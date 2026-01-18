package handler

import (
	"at-bot/internal/meta"

	"github.com/bwmarrin/discordgo"
)

const versionCommandName = "version"

type versionSlashCommand struct {
	baseSlashCommand
}

func NewVersionSlashCommand() *versionSlashCommand {
	return &versionSlashCommand{}
}

func (command *versionSlashCommand) CreateCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        versionCommandName,
		Description: "BOTのバージョン情報を表示します。",
	}
}

func (command *versionSlashCommand) InteractionType() discordgo.InteractionType {
	return discordgo.InteractionApplicationCommand
}

func (command *versionSlashCommand) InteractionID() string {
	return versionCommandName
}

func (command *versionSlashCommand) MatchInteractionID(interactionID string) bool {
	return command.InteractionID() == interactionID
}

func (command *versionSlashCommand) Handle(session *discordgo.Session, interaction *discordgo.Interaction) error {
	err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{toEmbed()},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func toEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🤖 @募っと",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Version", Value: meta.Version()},
			{Name: "Commit", Value: meta.CommitID()},
			{Name: "Built", Value: meta.BuildTime()},
			{Name: "Go(build)", Value: meta.GoBuild()},
		},
	}
}
