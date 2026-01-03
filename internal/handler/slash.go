package handler

import "github.com/bwmarrin/discordgo"

type baseSlashCommand struct{}

func (b *baseSlashCommand) getOptionMap(interaction *discordgo.Interaction) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	options := interaction.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	return optionMap
}
