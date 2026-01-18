package handler

import (
	"at-bot/internal/buildinfo"
	"fmt"
	"log"

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
	log.Printf(
		"[VERSION] user %s checked version %s (%s)",
		interaction.Member.User.ID,
		buildinfo.Version(),
		buildinfo.ShortCommitID(),
	)

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
		Title: fmt.Sprintf("🤖 %s", buildinfo.VersionWithPrefix()),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Commit", Value: buildinfo.CommitID()},
			{Name: "Built", Value: buildinfo.BuildTime()},
			{Name: "Go(build)", Value: buildinfo.GoBuild()},
		},
	}
}
