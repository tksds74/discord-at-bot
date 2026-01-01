package main

import (
	"at-bot/internal/db/sqlite"
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
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("環境変数の読み込みに失敗しました。: %v", err)
	}

	db, err := sqlite.InitDB("./data/bot.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// infra
	recruitRepo := sqlite.NewRecruitRepository(db)
	participantRepos := sqlite.NewParticipantRepository(db)
	txManager := sqlite.NewTxManager(db)
	// usecase
	recruitUsecase := recruit.NewRecruitUsecase(recruitRepo, participantRepos, txManager)
	// hadler
	startCmd := handler.NewStartRecruitCommand(recruitUsecase)
	joinCmd := handler.NewJoinRecruitCommand(recruitUsecase)
	declineCmd := handler.NewDeclineRecruitCommand(recruitUsecase)
	cancelCmd := handler.NewCancelRecruitCommand(recruitUsecase)
	deleteCmd := handler.NewDeleteRecruitCommand(recruitUsecase)

	prefixCommandDispatcher := &discord.PrefixCommandDispatcher{
		Listeners: []discord.PrefixCommandListener{
			startCmd,
		},
	}

	interactionDispatcher := &discord.InteractionDispatcher{
		Listeners: []discord.InteractionListener{
			joinCmd,
			declineCmd,
			cancelCmd,
			deleteCmd,
		},
	}

	config, err := discord.
		NewSessionConfig(
			discord.WithToken(os.Getenv("DISCORD_BOT_TOKEN")),
			discord.WithIntent(discordgo.IntentGuildMessages),
			discord.WithIntent(discordgo.IntentMessageContent),
			discord.WithMessageCreateHandler(prefixCommandDispatcher.OnMessageCreate),
			discord.WithInteractionCreateHandler(interactionDispatcher.OnInteractionCreate),
		)

	if err != nil {
		log.Fatalf("構成ファイルの構築に失敗しました。: %v", err)
	}

	var sm discord.SessionManager
	if err := sm.Open(config); err != nil {
		// TODO: os.Exit(1) で終了。deferは実行されない。
		log.Fatalf("DiscordBOTとの接続に失敗しました。: %v", err)
	}
	defer sm.Close()

	fmt.Println("Press Ctrl+C to exit")
	shutdown.WaitForExitSignal()
}
