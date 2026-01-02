package recruit

import (
	"context"
)

type RecruitRepository interface {
	Get(ctx context.Context, id RecruitID) (*RecruitState, error)
	GetByMessage(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error)
	Create(ctx context.Context, recruit *RecruitState) (RecruitID, error)
	Update(ctx context.Context, recruit *RecruitState) error
	Delete(ctx context.Context, id RecruitID) error
}

type ParticipantRepository interface {
	Upsert(ctx context.Context, recruitID RecruitID, userID UserID, status ParticipantStatus) error
	FindByRecruitAndUser(ctx context.Context, recruitID RecruitID, userID UserID) (*Participant, error)
	List(ctx context.Context, recruitID RecruitID) ([]Participant, error)
	DeleteAll(ctx context.Context, recruitID RecruitID) error
}
