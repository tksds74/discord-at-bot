package discord

import "fmt"

func FormatMention(userID string) string {
	return fmt.Sprintf("<@%s>", userID)
}

func FormatStrikethrough(text string) string {
	return fmt.Sprintf("~~%s~~", text)
}

func FormatBold(text string) string {
	return fmt.Sprintf("**%s**", text)
}
