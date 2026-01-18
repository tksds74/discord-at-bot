package buildinfo

import "testing"

func TestVersionWithPrefix(t *testing.T) {
	original := version
	t.Cleanup(func() { version = original })

	tests := []struct {
		in   string
		want string
	}{
		{"dev", "dev"},
		{"v1.2.3", "v1.2.3"},
		{"1.2.3", "v1.2.3"},
		{"tagname", "vtagname"},
	}

	for _, tt := range tests {
		version = tt.in
		if got := VersionWithPrefix(); got != tt.want {
			t.Errorf("VersionWithPrefix() = %v, want %v", got, tt.want)
		}
	}
}

func TestShortCommitID(t *testing.T) {
	original := commitID
	t.Cleanup(func() { commitID = original })

	tests := []struct {
		in   string
		want string
	}{
		{"unknown", "unknown"},
		{"abc", "abc"},
		{"0123456", "0123456"},
		{"01234567", "0123456"},
		{"012345678", "0123456"},
		{"abcdefghi", "abcdefg"},
	}

	for _, tt := range tests {
		commitID = tt.in
		if got := ShortCommitID(); got != tt.want {
			t.Errorf("ShortCommitID() = %v, want %v", got, tt.want)
		}
	}
}
