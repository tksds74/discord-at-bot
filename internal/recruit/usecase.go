package recruit

import (
	"at-bot/internal/uow"
	"context"
	"fmt"
	"time"
)

type RecruitUsecase struct {
	recruitRepos     RecruitRepository
	participantRepos ParticipantRepository
	uow              uow.UnitOfWork
}

func NewRecruitUsecase(
	recruitRepos RecruitRepository,
	participantRepos ParticipantRepository,
	uow uow.UnitOfWork,
) *RecruitUsecase {
	return &RecruitUsecase{
		recruitRepos:     recruitRepos,
		participantRepos: participantRepos,
		uow:              uow,
	}
}

func (uc *RecruitUsecase) Start(
	ctx context.Context,
	guildID GuildID,
	channelID ChannelID,
	messageID MessageID,
	maxCapacity int,
	authorID UserID,
) (*RecruitView, error) {
	var view *RecruitView
	err := uc.uow.Do(ctx, func(ctx context.Context) error {
		state := &RecruitState{
			GuildID:     guildID,
			ChannelID:   channelID,
			MessageID:   messageID,
			MaxCapacity: maxCapacity,
			AuthorID:    authorID,
			CreatedAt:   time.Now(),
			Status:      RecruitStatusOpened,
		}

		id, err := uc.recruitRepos.Create(ctx, state)
		if err != nil {
			return err
		}

		err = uc.participantRepos.Upsert(ctx, id, authorID, ParticipantStatusJoined)
		if err != nil {
			return err
		}

		state.ID = id
		view, err = uc.buildRecruitView(ctx, state)
		if err != nil {
			return err
		}
		return nil
	})
	return view, err
}

func (uc *RecruitUsecase) Join(
	ctx context.Context,
	channelID ChannelID,
	messageID MessageID,
	actorID UserID,
) (*ParticipantStatusChangeResult, error) {
	return uc.updateParticipantStatus(ctx, channelID, messageID, actorID, ParticipantStatusJoined)
}

func (uc *RecruitUsecase) Decline(
	ctx context.Context,
	channelID ChannelID,
	messageID MessageID,
	actorID UserID,
) (*ParticipantStatusChangeResult, error) {
	return uc.updateParticipantStatus(ctx, channelID, messageID, actorID, ParticipantStatusDeclined)
}

func (uc *RecruitUsecase) updateParticipantStatus(
	ctx context.Context,
	channelID ChannelID,
	messageID MessageID,
	actorID UserID,
	status ParticipantStatus,
) (*ParticipantStatusChangeResult, error) {
	var result *ParticipantStatusChangeResult
	err := uc.uow.Do(ctx, func(ctx context.Context) error {
		state, err := uc.recruitRepos.GetByMessage(ctx, channelID, messageID)
		if err != nil {
			return err
		}

		// 作成者は参加/辞退不可
		if state.AuthorID == actorID {
			return ErrAuthorCannotJoin
		}

		participant, err := uc.participantRepos.FindByRecruitAndUser(ctx, state.ID, actorID)
		if err != nil {
			return err
		}

		// 変更前のステータスを保存
		var previousStatus *ParticipantStatus
		if participant != nil {
			previousStatus = &participant.Status
		}

		// すでに同状態で登録済みの場合は独自エラー
		if participant != nil && participant.Status == status {
			switch status {
			case ParticipantStatusJoined:
				return ErrAlreadyJoined
			case ParticipantStatusDeclined:
				return ErrAlreadyDeclined
			}
		}

		// 参加状態を更新
		err = uc.participantRepos.Upsert(ctx, state.ID, actorID, status)
		if err != nil {
			return err
		}

		// 更新後のViewを構築
		view, err := uc.buildRecruitView(ctx, state)
		if err != nil {
			return err
		}

		result = &ParticipantStatusChangeResult{
			CurrentView:    view,
			PreviousStatus: previousStatus,
		}
		return nil
	})
	return result, err
}

func (uc *RecruitUsecase) buildRecruitView(
	ctx context.Context,
	state *RecruitState,
) (*RecruitView, error) {
	participants, err := uc.participantRepos.List(ctx, state.ID)
	if err != nil {
		return nil, err
	}

	var joinedUsers, declinedUsers []UserID
	for _, p := range participants {
		switch p.Status {
		case ParticipantStatusJoined:
			joinedUsers = append(joinedUsers, p.UserID)
		case ParticipantStatusDeclined:
			declinedUsers = append(declinedUsers, p.UserID)
		}
	}

	return &RecruitView{
		Meta:          state,
		JoinedUsers:   joinedUsers,
		DeclinedUsers: declinedUsers,
	}, nil
}

func (uc *RecruitUsecase) Cancel(
	ctx context.Context,
	channelID ChannelID,
	messageID MessageID,
	actorID UserID,
) (*ParticipantStatusChangeResult, error) {
	return uc.updateParticipantStatus(ctx, channelID, messageID, actorID, ParticipantStatusCanceled)
}

func (uc *RecruitUsecase) Delete(
	ctx context.Context,
	channelID ChannelID,
	messageID MessageID,
	actorID UserID,
) error {
	return uc.uow.Do(ctx, func(ctx context.Context) error {
		state, err := uc.recruitRepos.GetByMessage(ctx, channelID, messageID)
		if err != nil {
			return err
		}

		if state.AuthorID != actorID {
			return fmt.Errorf("作成者以外は募集を削除することはできません。")
		}

		err = uc.recruitRepos.Delete(ctx, state.ID)
		if err != nil {
			return err
		}

		return nil
	})
}
