package main

import (
	"at-bot/internal/db/sqlite"
	"at-bot/internal/dice"
	"at-bot/internal/discord"
	"at-bot/internal/handler"
	"at-bot/internal/recruit"
	"at-bot/internal/shutdown"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")

	db, err := sqlite.InitDB("./data/bot.db")
	if err != nil {
		log.Fatalf("[INIT] failed to initialize database: %v", err)
	}
	defer db.Close()

	// infra
	recruitRepo := sqlite.NewRecruitRepository(db)
	participantRepos := sqlite.NewParticipantRepository(db)
	txManager := sqlite.NewTxManager(db)
	// usecase
	recruitUsecase := recruit.NewRecruitUsecase(recruitRepo, participantRepos, txManager)
	diceUsecase := dice.NewDiceUsecase()
	// handler
	openCmd := handler.NewOpenRecruitCommand(recruitUsecase)
	openSlashCmd := handler.NewOpenRecruitSlashCommand(recruitUsecase)
	joinCmd := handler.NewJoinRecruitCommand(recruitUsecase)
	declineCmd := handler.NewDeclineRecruitCommand(recruitUsecase)
	cancelCmd := handler.NewCancelRecruitCommand(recruitUsecase)
	closeCmd := handler.NewCloseRecruitCommand(recruitUsecase)
	diceCmd := handler.NewDiceSlashCommand(diceUsecase)

	prefixCommandDispatcher := &discord.PrefixCommandDispatcher{
		Listeners: []discord.PrefixCommandListener{
			openCmd,
		},
	}

	interactionDispatcher := &discord.InteractionDispatcher{
		Listeners: []discord.InteractionListener{
			joinCmd,
			declineCmd,
			cancelCmd,
			closeCmd,
			openSlashCmd,
			diceCmd,
		},
	}

	config, err := discord.
		NewSessionConfig(
			discord.WithToken(os.Getenv("DISCORD_BOT_TOKEN")),
			discord.WithIntent(discordgo.IntentGuildMessages),
			discord.WithIntent(discordgo.IntentMessageContent),
			discord.WithMessageCreateHandler(prefixCommandDispatcher.OnMessageCreate),
			discord.WithInteractionCreateHandler(interactionDispatcher.OnInteractionCreate),
			discord.WithSlashCommand(openSlashCmd),
			discord.WithSlashCommand(diceCmd),
		)

	if err != nil {
		log.Fatalf("[INIT] failed to create session config: %v", err)
	}

	var sm discord.SessionManager
	if err := sm.Open(config); err != nil {
		log.Fatalf("[INIT] failed to connect to Discord: %v", err)
	}
	defer sm.Close()

	log.Println("[INIT] discord bot started successfully")
	fmt.Println("Press Ctrl+C to exit")
	shutdown.WaitForExitSignal()
}
