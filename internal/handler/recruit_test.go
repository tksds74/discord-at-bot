package handler

import (
	"at-bot/internal/recruit"
	"reflect"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestEncodeCustomID(t *testing.T) {
	tests := []struct {
		name    string
		items   map[string]string
		want    string
		wantErr bool
	}{
		{
			name: "複数のキーバリューを正しくエンコード",
			items: map[string]string{
				"key1":     "val:ue",
				"te:st":    "test23Value1",
				"customID": "recruit/fake",
			},
			want:    "8:customID12:recruit/fake4:key16:val:ue5:te:st12:test23Value1",
			wantErr: false,
		},
		{
			name: "単一のキーバリューをエンコード",
			items: map[string]string{
				"customID": "recruit/join",
			},
			want:    "8:customID12:recruit/join",
			wantErr: false,
		},
		{
			name: "空文字の値をエンコード",
			items: map[string]string{
				"key": "",
			},
			want:    "3:key0:",
			wantErr: false,
		},
		{
			name: "エンコード結果が100文字を大きく超える場合はエラー",
			items: map[string]string{
				"key": strings.Repeat("a", 100),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "エンコード結果が101文字の場合はエラー",
			items: map[string]string{
				"key": strings.Repeat("a", 93),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "エンコード結果が100文字なら許可",
			items: map[string]string{
				"key": strings.Repeat("a", 92),
			},
			want:    "3:key92:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodeCustomID(tt.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeCustomID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("encodeCustomID() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDecodeCustomID(t *testing.T) {
	tests := []struct {
		name       string
		encodedStr string
		want       map[string]string
		wantErr    bool
	}{
		{
			name:       "複数のキーバリューを正しくデコード",
			encodedStr: "8:customID12:recruit/fake4:key16:val:ue5:te:st12:test23Value1",
			want: map[string]string{
				"key1":     "val:ue",
				"te:st":    "test23Value1",
				"customID": "recruit/fake",
			},
			wantErr: false,
		},
		{
			name:       "単一のキーバリューをデコード",
			encodedStr: "8:customID12:recruit/join",
			want: map[string]string{
				"customID": "recruit/join",
			},
			wantErr: false,
		},
		{
			name:       "空文字の値をデコード",
			encodedStr: "3:key0:",
			want: map[string]string{
				"key": "",
			},
			wantErr: false,
		},
		{
			name:       "不正なフォーマット（コロンなし）",
			encodedStr: "invalidformat",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "不正なフォーマット（長さが数値でない）",
			encodedStr: "abc:key",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "不正なフォーマット（長さが実際の文字列より長い）",
			encodedStr: "10:key",
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeCustomID(tt.encodedStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeCustomID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeCustomID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeDecodeCustomID_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		items map[string]string
	}{
		{
			name: "シンプルなケース",
			items: map[string]string{
				"customID": "recruit/join",
			},
		},
		{
			name: "特殊文字を含むケース",
			items: map[string]string{
				"customID":  "recruit/close",
				"messageID": "123456789",
			},
		},
		{
			name: "コロンを含む値",
			items: map[string]string{
				"key": "value:with:colons",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := encodeCustomID(tt.items)
			if err != nil {
				t.Fatalf("encodeCustomID() error = %v", err)
			}

			decoded, err := decodeCustomID(encoded)
			if err != nil {
				t.Fatalf("decodeCustomID() error = %v", err)
			}

			if !reflect.DeepEqual(decoded, tt.items) {
				t.Errorf("round trip failed: got %v, want %v", decoded, tt.items)
			}
		})
	}
}

func TestRecruitState_ToEmbed(t *testing.T) {
	state := &recruitState{
		maxCapacity:  5,
		author:       "author-id",
		joinUsers:    []recruit.UserID{"author-id", "user1", "user2"},
		declineUsers: []recruit.UserID{"user3"},
	}

	embed := state.toEmbed()

	if embed.Title != "📢 募集開始 @5" {
		t.Errorf("Title = %v, want '📢 募集開始 @5'", embed.Title)
	}

	if !strings.Contains(embed.Description, "<@author-id>") {
		t.Errorf("Description should contain mention of author")
	}

	if len(embed.Fields) != 2 {
		t.Errorf("Fields length = %v, want 2", len(embed.Fields))
	}

	if embed.Fields[0].Name != joinLabel {
		t.Errorf("Fields[0].Name = %v, want %v", embed.Fields[0].Name, joinLabel)
	}

	if embed.Fields[1].Name != declineLabel {
		t.Errorf("Fields[1].Name = %v, want %v", embed.Fields[1].Name, declineLabel)
	}

	if embed.Color != 0xffa500 {
		t.Errorf("Color = %v, want 0xffa500", embed.Color)
	}
}

func TestRecruitState_ToComponent(t *testing.T) {
	state := &recruitState{
		maxCapacity: 5,
		author:      "author-id",
	}

	component := state.toComponent()

	if len(component.Components) != 2 {
		t.Errorf("Components length = %v, want 2", len(component.Components))
	}
}

func TestRecruitState_ToUsersString(t *testing.T) {
	tests := []struct {
		name    string
		userIds []recruit.UserID
		want    string
	}{
		{
			name:    "ユーザーが1人の場合",
			userIds: []recruit.UserID{"user1"},
			want:    "<@user1>",
		},
		{
			name:    "ユーザーが複数の場合",
			userIds: []recruit.UserID{"user1", "user2", "user3"},
			want:    "<@user1>\n<@user2>\n<@user3>",
		},
		{
			name:    "ユーザーがいない場合",
			userIds: []recruit.UserID{},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &recruitState{}
			got := state.toUsersString(tt.userIds)
			if got != tt.want {
				t.Errorf("toUsersString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromRecruitView(t *testing.T) {
	view := &recruit.RecruitView{
		Meta: &recruit.RecruitState{
			MaxCapacity: 5,
			AuthorID:    "author-id",
		},
		JoinedUsers:   []recruit.UserID{"author-id", "user1"},
		DeclinedUsers: []recruit.UserID{"user2"},
	}

	state := fromRecruitView(view)

	if state.maxCapacity != 5 {
		t.Errorf("maxCapacity = %v, want 5", state.maxCapacity)
	}

	if state.author != "author-id" {
		t.Errorf("author = %v, want author-id", state.author)
	}

	if len(state.joinUsers) != 2 {
		t.Errorf("joinUsers length = %v, want 2", len(state.joinUsers))
	}

	if len(state.declineUsers) != 1 {
		t.Errorf("declineUsers length = %v, want 1", len(state.declineUsers))
	}
}

func TestInitState(t *testing.T) {
	state := InitState("author-id", 5)

	if state.maxCapacity != 5 {
		t.Errorf("maxCapacity = %v, want 5", state.maxCapacity)
	}

	if state.author != "author-id" {
		t.Errorf("author = %v, want author-id", state.author)
	}

	if len(state.joinUsers) != 1 {
		t.Errorf("joinUsers length = %v, want 1", len(state.joinUsers))
	}

	if state.joinUsers[0] != "author-id" {
		t.Errorf("joinUsers[0] = %v, want author-id", state.joinUsers[0])
	}

	if len(state.declineUsers) != 0 {
		t.Errorf("declineUsers length = %v, want 0", len(state.declineUsers))
	}
}

func TestOpenRecruitSlashCommand_CreateCommand(t *testing.T) {
	cmd := NewOpenRecruitSlashCommand(nil)
	command := cmd.CreateCommand()

	if command.Name != recruitOpenCommandName {
		t.Errorf("CreateCommand().Name = %v, want %v", command.Name, recruitOpenCommandName)
	}

	if command.Description == "" {
		t.Errorf("CreateCommand().Description is empty")
	}

	if len(command.Options) != 1 {
		t.Errorf("CreateCommand().Options length = %v, want 1", len(command.Options))
		return
	}

	opt := command.Options[0]
	if opt.Name != recruitArgName {
		t.Errorf("CreateCommand().Options[0].Name = %v, want %v", opt.Name, recruitArgName)
	}

	if opt.Type != discordgo.ApplicationCommandOptionInteger {
		t.Errorf("CreateCommand().Options[0].Type = %v, want ApplicationCommandOptionInteger", opt.Type)
	}

	if !opt.Required {
		t.Errorf("CreateCommand().Options[0].Required = false, want true")
	}

	if opt.MinValue == nil || *opt.MinValue != 1.0 {
		t.Errorf("CreateCommand().Options[0].MinValue = %v, want 1.0", opt.MinValue)
	}
}

func TestOpenRecruitCommand_ExtractArgNumber(t *testing.T) {
	cmd := &openRecruitCommand{}

	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name:    "半角スペース区切りで数値を抽出",
			content: "@5",
			want:    5,
			wantErr: false,
		},
		{
			name:    "全角スペース区切りで数値を抽出",
			content: "@　10",
			want:    10,
			wantErr: false,
		},
		{
			name:    "半角スペースで区切られた数値",
			content: "@ 3",
			want:    3,
			wantErr: false,
		},
		{
			name:    "数値でない場合はエラー",
			content: "@abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "空の場合はエラー",
			content: "@",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cmd.extractArgNumber(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractArgNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("extractArgNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateJoinMessage(t *testing.T) {
	tests := []struct {
		name        string
		actorID     recruit.UserID
		maxCapacity int
		joinedUsers []recruit.UserID
		wantContain []string
	}{
		{
			name:        "残り枠がある場合",
			actorID:     "user-1",
			maxCapacity: 5,
			joinedUsers: []recruit.UserID{"author", "user-1"},
			wantContain: []string{
				"<@user-1> が参加しました。",
				"@4", // 残り4枠
			},
		},
		{
			name:        "定員ちょうどの場合（満員）",
			actorID:     "user-5",
			maxCapacity: 5,
			joinedUsers: []recruit.UserID{"author", "user-1", "user-2", "user-3", "user-4", "user-5"},
			wantContain: []string{
				"<@user-5> が参加しました。",
				"**[募集終了]**",
				"<@author>",
				"<@user-1>",
				"<@user-5>",
			},
		},
		{
			name:        "定員超過の場合",
			actorID:     "user-6",
			maxCapacity: 5,
			joinedUsers: []recruit.UserID{"author", "user-1", "user-2", "user-3", "user-4", "user-5", "user-6"},
			wantContain: []string{
				"<@user-6> が参加しました。",
			},
		},
		{
			name:        "定員1で2人目が参加（満員）",
			actorID:     "user-1",
			maxCapacity: 1,
			joinedUsers: []recruit.UserID{"author", "user-1"},
			wantContain: []string{
				"<@user-1> が参加しました。",
				"**[募集終了]**",
				"<@author>",
				"<@user-1>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := &recruit.RecruitView{
				Meta: &recruit.RecruitState{
					MaxCapacity: tt.maxCapacity,
				},
				JoinedUsers: tt.joinedUsers,
			}

			got := createJoinMessage(tt.actorID, view)

			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("createJoinMessage() = %v, should contain %v", got, want)
				}
			}
		})
	}
}

func TestBaseInteractionCommand_MatchCustomID(t *testing.T) {
	tests := []struct {
		name     string
		customID string
		target   string
		want     bool
	}{
		{
			name:     "マッチする場合",
			customID: "8:customID12:recruit/join",
			target:   "recruit/join",
			want:     true,
		},
		{
			name:     "マッチしない場合",
			customID: "8:customID13:recruit/close",
			target:   "recruit/join",
			want:     false,
		},
		{
			name:     "不正なフォーマット",
			customID: "invalid",
			target:   "recruit/join",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &customIDInteractionCommand{
				customID: tt.target,
			}
			got := cmd.MatchInteractionID(tt.customID)
			if got != tt.want {
				t.Errorf("MatchCustomID() = %v, want %v", got, tt.want)
			}
		})
	}
}
