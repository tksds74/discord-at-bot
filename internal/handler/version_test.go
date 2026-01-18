package handler

import "testing"

func TestVersionSlashCommand_CreateCommand(t *testing.T) {
	cmd := NewVersionSlashCommand()
	command := cmd.CreateCommand()

	if command.Name != versionCommandName {
		t.Errorf("CreateCommand().Name = %v, want %v", command.Name, versionCommandName)
	}

	if command.Description == "" {
		t.Errorf("CreateCommand().Description is empty")
	}
}
