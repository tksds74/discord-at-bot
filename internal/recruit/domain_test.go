package recruit

import (
	"testing"
)

func TestRecruitView_RemainingSlots(t *testing.T) {
	tests := []struct {
		name        string
		maxCapacity int
		joinedUsers []UserID
		want        int
	}{
		{
			name:        "募集人数5で作成者のみ参加(1名)の場合、残り5枠",
			maxCapacity: 5,
			joinedUsers: []UserID{"author"},
			want:        5,
		},
		{
			name:        "募集人数5で3名参加の場合、残り3枠",
			maxCapacity: 5,
			joinedUsers: []UserID{"author", "user1", "user2"},
			want:        3,
		},
		{
			name:        "募集人数5で6名参加（募集人数ちょうど)の場合、残り0枠",
			maxCapacity: 5,
			joinedUsers: []UserID{"author", "user1", "user2", "user3", "user4", "user5"},
			want:        0,
		},
		{
			name:        "募集人数5で7名参加（募集人数オーバー)の場合、残り0枠",
			maxCapacity: 5,
			joinedUsers: []UserID{"author", "user1", "user2", "user3", "user4", "user5", "user6"},
			want:        0,
		},
		{
			name:        "募集人数1で作成者のみの場合、残り1枠",
			maxCapacity: 1,
			joinedUsers: []UserID{"author"},
			want:        1,
		},
		{
			name:        "募集人数1で2名参加の場合、残り0枠",
			maxCapacity: 1,
			joinedUsers: []UserID{"author", "user1"},
			want:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := &RecruitView{
				Meta: &RecruitState{
					MaxCapacity: tt.maxCapacity,
				},
				JoinedUsers: tt.joinedUsers,
			}
			if got := view.RemainingSlots(); got != tt.want {
				t.Errorf("RemainingSlots() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecruitView_IsFull(t *testing.T) {
	tests := []struct {
		name        string
		maxCapacity int
		joinedUsers []UserID
		want        bool
	}{
		{
			name:        "募集人数5で作成者のみ参加（1名)の場合、募集中",
			maxCapacity: 5,
			joinedUsers: []UserID{"author"},
			want:        false,
		},
		{
			name:        "募集人数5で5名参加の場合、募集中",
			maxCapacity: 5,
			joinedUsers: []UserID{"author", "user1", "user2", "user3", "user4"},
			want:        false,
		},
		{
			name:        "募集人数5で6名参加（募集人数ちょうど)の場合、満員",
			maxCapacity: 5,
			joinedUsers: []UserID{"author", "user1", "user2", "user3", "user4", "user5"},
			want:        true,
		},
		{
			name:        "募集人数5で7名参加（募集人数オーバー)の場合、満員",
			maxCapacity: 5,
			joinedUsers: []UserID{"author", "user1", "user2", "user3", "user4", "user5", "user6"},
			want:        true,
		},
		{
			name:        "募集人数1で作成者のみの場合、募集中",
			maxCapacity: 1,
			joinedUsers: []UserID{"author"},
			want:        false,
		},
		{
			name:        "募集人数1で2名参加（募集人数ちょうど)の場合、満員",
			maxCapacity: 1,
			joinedUsers: []UserID{"author", "user1"},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := &RecruitView{
				Meta: &RecruitState{
					MaxCapacity: tt.maxCapacity,
				},
				JoinedUsers: tt.joinedUsers,
			}
			if got := view.IsFull(); got != tt.want {
				t.Errorf("IsFull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecruitView_ExtraCount(t *testing.T) {
	tests := []struct {
		name        string
		maxCapacity int
		joinedUsers []UserID
		want        int
	}{
		{
			name:        "募集人数5で作成者のみ参加（1名)の場合、超過0名",
			maxCapacity: 5,
			joinedUsers: []UserID{"author"},
			want:        0,
		},
		{
			name:        "募集人数5で6名参加（募集人数ちょうど)の場合、超過0名",
			maxCapacity: 5,
			joinedUsers: []UserID{"author", "user1", "user2", "user3", "user4", "user5"},
			want:        0,
		},
		{
			name:        "募集人数5で7名参加（1名超過)の場合、超過1名",
			maxCapacity: 5,
			joinedUsers: []UserID{"author", "user1", "user2", "user3", "user4", "user5", "user6"},
			want:        1,
		},
		{
			name:        "募集人数5で9名参加（3名超過)の場合、超過3名",
			maxCapacity: 5,
			joinedUsers: []UserID{"author", "user1", "user2", "user3", "user4", "user5", "user6", "user7", "user8"},
			want:        3,
		},
		{
			name:        "募集人数1で2名参加（募集人数ちょうど)の場合、超過0名",
			maxCapacity: 1,
			joinedUsers: []UserID{"author", "user1"},
			want:        0,
		},
		{
			name:        "募集人数1で3名参加（1名超過)の場合、超過1名",
			maxCapacity: 1,
			joinedUsers: []UserID{"author", "user1", "user2"},
			want:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := &RecruitView{
				Meta: &RecruitState{
					MaxCapacity: tt.maxCapacity,
				},
				JoinedUsers: tt.joinedUsers,
			}
			if got := view.ExtraCount(); got != tt.want {
				t.Errorf("ExtraCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
