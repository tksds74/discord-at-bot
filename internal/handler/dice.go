package handler

import (
	"at-bot/internal/dice"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// ダイスコマンド用の固定値
const (
	diceCommandName = "dice"
	diceArgName     = "個数"
)

type diceSlashCommand struct {
	baseSlashCommand
	service *dice.DiceUsecase
}

func NewDiceSlashCommand(service *dice.DiceUsecase) *diceSlashCommand {
	return &diceSlashCommand{
		service: service,
	}
}

func (command *diceSlashCommand) CreateCommand() *discordgo.ApplicationCommand {
	minValue := 1.0
	maxValue := 100.0
	return &discordgo.ApplicationCommand{
		Name:        diceCommandName,
		Description: "6面ダイスを振った結果を返します。(オプションでダイスの個数指定可)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        diceArgName,
				Description: "振るダイスの個数を入力します。(省略時: 1個)",
				Required:    false,
				MinValue:    &minValue,
				MaxValue:    maxValue,
			},
		},
	}
}

func (command *diceSlashCommand) InteractionType() discordgo.InteractionType {
	return discordgo.InteractionApplicationCommand
}

func (command *diceSlashCommand) InteractionID() string {
	return diceCommandName
}

func (command *diceSlashCommand) MatchInteractionID(interactionID string) bool {
	return command.InteractionID() == interactionID
}

func (command *diceSlashCommand) Handle(session *discordgo.Session, interaction *discordgo.Interaction) error {
	optionMap := command.getOptionMap(interaction)
	opt, ok := optionMap[diceArgName]

	diceCount := 1
	if ok && opt != nil {
		diceCount = int(opt.IntValue())
	}

	results, err := command.service.Dice(diceCount)
	if err != nil {
		return err
	}

	err = session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: strings.Join(results, " "),
		},
	})

	if err != nil {
		return err
	}

	return nil
}
