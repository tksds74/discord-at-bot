package discord

import "testing"

func TestFormatMention(t *testing.T) {
	got := FormatMention("1234567890")
	want := "<@1234567890>"
	if got != want {
		t.Errorf("FormatMention(\"1234567890\") == %s, want %s", got, want)
	}
}

func TestFormatStrikethrough(t *testing.T) {
	got := FormatStrikethrough("1234567890")
	want := "~~1234567890~~"
	if got != want {
		t.Errorf("FormatStrikethrough(\"1234567890\") == %s, want %s", got, want)
	}
}

func TestFormatBold(t *testing.T) {
	got := FormatBold("1234567890")
	want := "**1234567890**"
	if got != want {
		t.Errorf("FormatBold(\"1234567890\") == %s, want %s", got, want)
	}
}
