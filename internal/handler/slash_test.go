package handler

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestBaseSlashCommand_GetOptionMap(t *testing.T) {
	base := &baseSlashCommand{}

	tests := []struct {
		name    string
		options []*discordgo.ApplicationCommandInteractionDataOption
		want    map[string]*discordgo.ApplicationCommandInteractionDataOption
	}{
		{
			name: "複数のオプションを正しくマップ化",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Name: "option1", Type: discordgo.ApplicationCommandOptionString, Value: "value1"},
				{Name: "option2", Type: discordgo.ApplicationCommandOptionInteger, Value: int64(42)},
				{Name: "option3", Type: discordgo.ApplicationCommandOptionBoolean, Value: true},
			},
			want: map[string]*discordgo.ApplicationCommandInteractionDataOption{
				"option1": {Name: "option1", Type: discordgo.ApplicationCommandOptionString, Value: "value1"},
				"option2": {Name: "option2", Type: discordgo.ApplicationCommandOptionInteger, Value: int64(42)},
				"option3": {Name: "option3", Type: discordgo.ApplicationCommandOptionBoolean, Value: true},
			},
		},
		{
			name: "単一のオプション",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Name: "test", Type: discordgo.ApplicationCommandOptionInteger, Value: int64(5)},
			},
			want: map[string]*discordgo.ApplicationCommandInteractionDataOption{
				"test": {Name: "test", Type: discordgo.ApplicationCommandOptionInteger, Value: int64(5)},
			},
		},
		{
			name:    "オプションが空",
			options: []*discordgo.ApplicationCommandInteractionDataOption{},
			want:    map[string]*discordgo.ApplicationCommandInteractionDataOption{},
		},
		{
			name:    "オプションがnil",
			options: nil,
			want:    map[string]*discordgo.ApplicationCommandInteractionDataOption{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interaction := &discordgo.Interaction{
				Type: discordgo.InteractionApplicationCommand,
				Data: discordgo.ApplicationCommandInteractionData{
					Options: tt.options,
				},
			}

			got := base.getOptionMap(interaction)

			if len(got) != len(tt.want) {
				t.Errorf("getOptionMap() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for key, wantOpt := range tt.want {
				gotOpt, ok := got[key]
				if !ok {
					t.Errorf("getOptionMap() missing key %v", key)
					continue
				}
				if gotOpt.Name != wantOpt.Name {
					t.Errorf("getOptionMap()[%v].Name = %v, want %v", key, gotOpt.Name, wantOpt.Name)
				}
				if gotOpt.Type != wantOpt.Type {
					t.Errorf("getOptionMap()[%v].Type = %v, want %v", key, gotOpt.Type, wantOpt.Type)
				}
				if gotOpt.Value != wantOpt.Value {
					t.Errorf("getOptionMap()[%v].Value = %v, want %v", key, gotOpt.Value, wantOpt.Value)
				}
			}
		})
	}
}
