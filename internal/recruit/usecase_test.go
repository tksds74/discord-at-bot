package recruit

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Mock repositories
type mockRecruitRepository struct {
	getByMessageFunc func(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error)
	createFunc       func(ctx context.Context, state *RecruitState) (RecruitID, error)
	deleteFunc       func(ctx context.Context, id RecruitID) error
}

func (m *mockRecruitRepository) Get(ctx context.Context, id RecruitID) (*RecruitState, error) {
	return nil, nil
}

func (m *mockRecruitRepository) GetByMessage(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error) {
	if m.getByMessageFunc != nil {
		return m.getByMessageFunc(ctx, channelID, messageID)
	}
	return nil, nil
}

func (m *mockRecruitRepository) Create(ctx context.Context, state *RecruitState) (RecruitID, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, state)
	}
	return 1, nil
}

func (m *mockRecruitRepository) Update(ctx context.Context, state *RecruitState) error {
	return nil
}

func (m *mockRecruitRepository) Delete(ctx context.Context, id RecruitID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

type mockParticipantRepository struct {
	upsertFunc               func(ctx context.Context, recruitID RecruitID, userID UserID, status ParticipantStatus) error
	findByRecruitAndUserFunc func(ctx context.Context, recruitID RecruitID, userID UserID) (*Participant, error)
	listFunc                 func(ctx context.Context, recruitID RecruitID) ([]Participant, error)
}

func (m *mockParticipantRepository) Upsert(ctx context.Context, recruitID RecruitID, userID UserID, status ParticipantStatus) error {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, recruitID, userID, status)
	}
	return nil
}

func (m *mockParticipantRepository) FindByRecruitAndUser(ctx context.Context, recruitID RecruitID, userID UserID) (*Participant, error) {
	if m.findByRecruitAndUserFunc != nil {
		return m.findByRecruitAndUserFunc(ctx, recruitID, userID)
	}
	return nil, nil
}

func (m *mockParticipantRepository) List(ctx context.Context, recruitID RecruitID) ([]Participant, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, recruitID)
	}
	return []Participant{}, nil
}

func (m *mockParticipantRepository) DeleteAll(ctx context.Context, recruitID RecruitID) error {
	return nil
}

type mockUnitOfWork struct {
	doFunc func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.doFunc != nil {
		return m.doFunc(ctx, fn)
	}
	// デフォルトはトランザクション処理を実行
	return fn(ctx)
}

func TestRecruitUsecase_Start(t *testing.T) {
	ctx := context.Background()

	t.Run("募集開始が正常に作成される", func(t *testing.T) {
		recruitRepo := &mockRecruitRepository{
			createFunc: func(ctx context.Context, state *RecruitState) (RecruitID, error) {
				if state.GuildID != "guild-1" {
					t.Errorf("GuildID = %v, want guild-1", state.GuildID)
				}
				if state.ChannelID != "channel-1" {
					t.Errorf("ChannelID = %v, want channel-1", state.ChannelID)
				}
				if state.MessageID != "message-1" {
					t.Errorf("MessageID = %v, want message-1", state.MessageID)
				}
				if state.MaxCapacity != 5 {
					t.Errorf("MaxCapacity = %v, want 5", state.MaxCapacity)
				}
				if state.AuthorID != "author-1" {
					t.Errorf("AuthorID = %v, want author-1", state.AuthorID)
				}
				if state.Status != RecruitStatusOpened {
					t.Errorf("Status = %v, want %v", state.Status, RecruitStatusOpened)
				}
				return 1, nil
			},
		}

		participantRepo := &mockParticipantRepository{
			upsertFunc: func(ctx context.Context, recruitID RecruitID, userID UserID, status ParticipantStatus) error {
				if recruitID != 1 {
					t.Errorf("RecruitID = %v, want 1", recruitID)
				}
				if userID != "author-1" {
					t.Errorf("UserID = %v, want author-1", userID)
				}
				if status != ParticipantStatusJoined {
					t.Errorf("Status = %v, want %v", status, ParticipantStatusJoined)
				}
				return nil
			},
			listFunc: func(ctx context.Context, recruitID RecruitID) ([]Participant, error) {
				return []Participant{
					{RecruitID: recruitID, UserID: "author-1", Status: ParticipantStatusJoined},
				}, nil
			},
		}

		uow := &mockUnitOfWork{}
		uc := NewRecruitUsecase(recruitRepo, participantRepo, uow)

		view, err := uc.Start(ctx, "guild-1", "channel-1", "message-1", 5, "author-1")
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		if view == nil {
			t.Fatal("view is nil")
		}
		if len(view.JoinedUsers) != 1 {
			t.Errorf("JoinedUsers length = %v, want 1", len(view.JoinedUsers))
		}
		if view.JoinedUsers[0] != "author-1" {
			t.Errorf("JoinedUsers[0] = %v, want author-1", view.JoinedUsers[0])
		}
	})

	t.Run("リポジトリエラーの場合はエラーを返す", func(t *testing.T) {
		recruitRepo := &mockRecruitRepository{
			createFunc: func(ctx context.Context, state *RecruitState) (RecruitID, error) {
				return 0, errors.New("database error")
			},
		}
		participantRepo := &mockParticipantRepository{}
		uow := &mockUnitOfWork{}
		uc := NewRecruitUsecase(recruitRepo, participantRepo, uow)

		_, err := uc.Start(ctx, "guild-1", "channel-1", "message-1", 5, "author-1")
		if err == nil {
			t.Error("Start() error = nil, want error")
		}
	})
}

func TestRecruitUsecase_Join(t *testing.T) {
	ctx := context.Background()

	t.Run("ユーザーが正常に参加できる", func(t *testing.T) {
		recruitRepo := &mockRecruitRepository{
			getByMessageFunc: func(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error) {
				return &RecruitState{
					ID:          1,
					GuildID:     "guild-1",
					ChannelID:   channelID,
					MessageID:   messageID,
					AuthorID:    "author-1",
					MaxCapacity: 5,
					Status:      RecruitStatusOpened,
					CreatedAt:   time.Now(),
				}, nil
			},
		}

		participantRepo := &mockParticipantRepository{
			findByRecruitAndUserFunc: func(ctx context.Context, recruitID RecruitID, userID UserID) (*Participant, error) {
				// 初回参加なのでnil
				return nil, nil
			},
			upsertFunc: func(ctx context.Context, recruitID RecruitID, userID UserID, status ParticipantStatus) error {
				if userID != "user-1" {
					t.Errorf("UserID = %v, want user-1", userID)
				}
				if status != ParticipantStatusJoined {
					t.Errorf("Status = %v, want %v", status, ParticipantStatusJoined)
				}
				return nil
			},
			listFunc: func(ctx context.Context, recruitID RecruitID) ([]Participant, error) {
				return []Participant{
					{RecruitID: recruitID, UserID: "author-1", Status: ParticipantStatusJoined},
					{RecruitID: recruitID, UserID: "user-1", Status: ParticipantStatusJoined},
				}, nil
			},
		}

		uow := &mockUnitOfWork{}
		uc := NewRecruitUsecase(recruitRepo, participantRepo, uow)

		result, err := uc.Join(ctx, "channel-1", "message-1", "user-1")
		if err != nil {
			t.Fatalf("Join() error = %v", err)
		}

		if result == nil {
			t.Fatal("result is nil")
		}
		if result.PreviousStatus != nil {
			t.Errorf("PreviousStatus = %v, want nil", *result.PreviousStatus)
		}
		if len(result.CurrentView.JoinedUsers) != 2 {
			t.Errorf("JoinedUsers length = %v, want 2", len(result.CurrentView.JoinedUsers))
		}
	})

	t.Run("作成者は参加できない", func(t *testing.T) {
		recruitRepo := &mockRecruitRepository{
			getByMessageFunc: func(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error) {
				return &RecruitState{
					ID:          1,
					GuildID:     "guild-1",
					ChannelID:   channelID,
					MessageID:   messageID,
					AuthorID:    "author-1",
					MaxCapacity: 5,
					Status:      RecruitStatusOpened,
					CreatedAt:   time.Now(),
				}, nil
			},
		}

		participantRepo := &mockParticipantRepository{}
		uow := &mockUnitOfWork{}
		uc := NewRecruitUsecase(recruitRepo, participantRepo, uow)

		_, err := uc.Join(ctx, "channel-1", "message-1", "author-1")
		if !errors.Is(err, ErrAuthorCannotJoin) {
			t.Errorf("Join() error = %v, want ErrAuthorCannotJoin", err)
		}
	})

	t.Run("既に参加済みの場合はエラー", func(t *testing.T) {
		recruitRepo := &mockRecruitRepository{
			getByMessageFunc: func(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error) {
				return &RecruitState{
					ID:          1,
					GuildID:     "guild-1",
					ChannelID:   channelID,
					MessageID:   messageID,
					AuthorID:    "author-1",
					MaxCapacity: 5,
					Status:      RecruitStatusOpened,
					CreatedAt:   time.Now(),
				}, nil
			},
		}

		participantRepo := &mockParticipantRepository{
			findByRecruitAndUserFunc: func(ctx context.Context, recruitID RecruitID, userID UserID) (*Participant, error) {
				return &Participant{
					RecruitID: recruitID,
					UserID:    userID,
					Status:    ParticipantStatusJoined,
					CreatedAt: time.Now(),
				}, nil
			},
		}

		uow := &mockUnitOfWork{}
		uc := NewRecruitUsecase(recruitRepo, participantRepo, uow)

		_, err := uc.Join(ctx, "channel-1", "message-1", "user-1")
		if !errors.Is(err, ErrAlreadyJoined) {
			t.Errorf("Join() error = %v, want ErrAlreadyJoined", err)
		}
	})
}

func TestRecruitUsecase_Decline(t *testing.T) {
	ctx := context.Background()

	t.Run("ユーザーが正常に辞退できる", func(t *testing.T) {
		recruitRepo := &mockRecruitRepository{
			getByMessageFunc: func(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error) {
				return &RecruitState{
					ID:          1,
					GuildID:     "guild-1",
					ChannelID:   channelID,
					MessageID:   messageID,
					AuthorID:    "author-1",
					MaxCapacity: 5,
					Status:      RecruitStatusOpened,
					CreatedAt:   time.Now(),
				}, nil
			},
		}

		participantRepo := &mockParticipantRepository{
			findByRecruitAndUserFunc: func(ctx context.Context, recruitID RecruitID, userID UserID) (*Participant, error) {
				return nil, nil
			},
			upsertFunc: func(ctx context.Context, recruitID RecruitID, userID UserID, status ParticipantStatus) error {
				if status != ParticipantStatusDeclined {
					t.Errorf("Status = %v, want %v", status, ParticipantStatusDeclined)
				}
				return nil
			},
			listFunc: func(ctx context.Context, recruitID RecruitID) ([]Participant, error) {
				return []Participant{
					{RecruitID: recruitID, UserID: "author-1", Status: ParticipantStatusJoined},
					{RecruitID: recruitID, UserID: "user-1", Status: ParticipantStatusDeclined},
				}, nil
			},
		}

		uow := &mockUnitOfWork{}
		uc := NewRecruitUsecase(recruitRepo, participantRepo, uow)

		result, err := uc.Decline(ctx, "channel-1", "message-1", "user-1")
		if err != nil {
			t.Fatalf("Decline() error = %v", err)
		}

		if len(result.CurrentView.DeclinedUsers) != 1 {
			t.Errorf("DeclinedUsers length = %v, want 1", len(result.CurrentView.DeclinedUsers))
		}
	})

	t.Run("既に辞退済みの場合はエラー", func(t *testing.T) {
		recruitRepo := &mockRecruitRepository{
			getByMessageFunc: func(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error) {
				return &RecruitState{
					ID:          1,
					GuildID:     "guild-1",
					ChannelID:   channelID,
					MessageID:   messageID,
					AuthorID:    "author-1",
					MaxCapacity: 5,
					Status:      RecruitStatusOpened,
					CreatedAt:   time.Now(),
				}, nil
			},
		}

		participantRepo := &mockParticipantRepository{
			findByRecruitAndUserFunc: func(ctx context.Context, recruitID RecruitID, userID UserID) (*Participant, error) {
				return &Participant{
					RecruitID: recruitID,
					UserID:    userID,
					Status:    ParticipantStatusDeclined,
					CreatedAt: time.Now(),
				}, nil
			},
		}

		uow := &mockUnitOfWork{}
		uc := NewRecruitUsecase(recruitRepo, participantRepo, uow)

		_, err := uc.Decline(ctx, "channel-1", "message-1", "user-1")
		if !errors.Is(err, ErrAlreadyDeclined) {
			t.Errorf("Decline() error = %v, want ErrAlreadyDeclined", err)
		}
	})
}

func TestRecruitUsecase_Cancel(t *testing.T) {
	ctx := context.Background()

	t.Run("参加をキャンセルできる", func(t *testing.T) {
		recruitRepo := &mockRecruitRepository{
			getByMessageFunc: func(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error) {
				return &RecruitState{
					ID:          1,
					GuildID:     "guild-1",
					ChannelID:   channelID,
					MessageID:   messageID,
					AuthorID:    "author-1",
					MaxCapacity: 5,
					Status:      RecruitStatusOpened,
					CreatedAt:   time.Now(),
				}, nil
			},
		}

		participantRepo := &mockParticipantRepository{
			findByRecruitAndUserFunc: func(ctx context.Context, recruitID RecruitID, userID UserID) (*Participant, error) {
				return &Participant{
					RecruitID: recruitID,
					UserID:    userID,
					Status:    ParticipantStatusJoined,
					CreatedAt: time.Now(),
				}, nil
			},
			upsertFunc: func(ctx context.Context, recruitID RecruitID, userID UserID, status ParticipantStatus) error {
				if status != ParticipantStatusCanceled {
					t.Errorf("Status = %v, want %v", status, ParticipantStatusCanceled)
				}
				return nil
			},
			listFunc: func(ctx context.Context, recruitID RecruitID) ([]Participant, error) {
				return []Participant{
					{RecruitID: recruitID, UserID: "author-1", Status: ParticipantStatusJoined},
				}, nil
			},
		}

		uow := &mockUnitOfWork{}
		uc := NewRecruitUsecase(recruitRepo, participantRepo, uow)

		result, err := uc.Cancel(ctx, "channel-1", "message-1", "user-1")
		if err != nil {
			t.Fatalf("Cancel() error = %v", err)
		}

		if result.PreviousStatus == nil || *result.PreviousStatus != ParticipantStatusJoined {
			t.Error("PreviousStatus should be ParticipantStatusJoined")
		}
	})
}

func TestRecruitUsecase_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("作成者は募集を削除できる", func(t *testing.T) {
		recruitRepo := &mockRecruitRepository{
			getByMessageFunc: func(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error) {
				return &RecruitState{
					ID:          1,
					GuildID:     "guild-1",
					ChannelID:   channelID,
					MessageID:   messageID,
					AuthorID:    "author-1",
					MaxCapacity: 5,
					Status:      RecruitStatusOpened,
					CreatedAt:   time.Now(),
				}, nil
			},
			deleteFunc: func(ctx context.Context, id RecruitID) error {
				if id != 1 {
					t.Errorf("RecruitID = %v, want 1", id)
				}
				return nil
			},
		}

		participantRepo := &mockParticipantRepository{}
		uow := &mockUnitOfWork{}
		uc := NewRecruitUsecase(recruitRepo, participantRepo, uow)

		err := uc.Delete(ctx, "channel-1", "message-1", "author-1")
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}
	})

	t.Run("作成者以外は削除できない", func(t *testing.T) {
		recruitRepo := &mockRecruitRepository{
			getByMessageFunc: func(ctx context.Context, channelID ChannelID, messageID MessageID) (*RecruitState, error) {
				return &RecruitState{
					ID:          1,
					GuildID:     "guild-1",
					ChannelID:   channelID,
					MessageID:   messageID,
					AuthorID:    "author-1",
					MaxCapacity: 5,
					Status:      RecruitStatusOpened,
					CreatedAt:   time.Now(),
				}, nil
			},
		}

		participantRepo := &mockParticipantRepository{}
		uow := &mockUnitOfWork{}
		uc := NewRecruitUsecase(recruitRepo, participantRepo, uow)

		err := uc.Delete(ctx, "channel-1", "message-1", "user-1")
		if err == nil {
			t.Error("Delete() error = nil, want error")
		}
	})
}
