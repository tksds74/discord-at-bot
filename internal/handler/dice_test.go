package handler

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestDiceSlashCommand_CreateCommand(t *testing.T) {
	cmd := NewDiceSlashCommand(nil)
	command := cmd.CreateCommand()

	if command.Name != diceCommandName {
		t.Errorf("CreateCommand().Name = %v, want %v", command.Name, diceCommandName)
	}

	if command.Description == "" {
		t.Errorf("CreateCommand().Description is empty")
	}

	if len(command.Options) != 1 {
		t.Errorf("CreateCommand().Options length = %v, want 1", len(command.Options))
		return
	}

	opt := command.Options[0]
	if opt.Name != diceArgName {
		t.Errorf("CreateCommand().Options[0].Name = %v, want %v", opt.Name, diceArgName)
	}

	if opt.Type != discordgo.ApplicationCommandOptionInteger {
		t.Errorf("CreateCommand().Options[0].Type = %v, want ApplicationCommandOptionInteger", opt.Type)
	}

	if opt.Required {
		t.Errorf("CreateCommand().Options[0].Required = true, want false")
	}

	if opt.MinValue == nil || *opt.MinValue != 1.0 {
		t.Errorf("CreateCommand().Options[0].MinValue = %v, want 1.0", opt.MinValue)
	}

	if opt.MaxValue != 100.0 {
		t.Errorf("CreateCommand().Options[0].MaxValue = %v, want 100.0", opt.MaxValue)
	}
}
