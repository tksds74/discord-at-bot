package recruit

import (
	"errors"
	"time"
)

type RecruitID int64
type GuildID string
type ChannelID string
type MessageID string
type UserID string
type RecruitStatus string

const (
	RecruitStatusOpened RecruitStatus = "opened"
	RecruitStatusClosed RecruitStatus = "closed"
)

type ParticipantStatus string

const (
	ParticipantStatusJoined   ParticipantStatus = "joined"
	ParticipantStatusDeclined ParticipantStatus = "declined"
	ParticipantStatusCanceled ParticipantStatus = "canceled"
)

type RecruitState struct {
	ID          RecruitID
	GuildID     GuildID
	ChannelID   ChannelID
	MessageID   MessageID
	AuthorID    UserID
	MaxCapacity int
	Status      RecruitStatus
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type Participant struct {
	RecruitID RecruitID
	UserID    UserID
	Status    ParticipantStatus
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type RecruitView struct {
	Meta          *RecruitState
	JoinedUsers   []UserID
	DeclinedUsers []UserID
}

func (v *RecruitView) RemainingSlots() int {
	remaining := v.Meta.MaxCapacity - len(v.JoinedUsers) + 1
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (v *RecruitView) IsFull() bool {
	return v.Meta.MaxCapacity <= len(v.JoinedUsers)-1
}

func (v *RecruitView) ExtraCount() int {
	extra := len(v.JoinedUsers) - v.Meta.MaxCapacity - 1
	if extra < 0 {
		return 0
	}
	return extra
}

var (
	ErrAlreadyJoined    = errors.New("既に参加済みです")
	ErrAlreadyDeclined  = errors.New("既に辞退済みです")
	ErrAuthorCannotJoin = errors.New("作成者は参加/辞退できません")
	ErrRecruitNotFound  = errors.New("募集が見つかりません")
)

type ParticipantStatusChangeResult struct {
	CurrentView    *RecruitView
	PreviousStatus *ParticipantStatus
}
