package sqlite

import (
	"at-bot/internal/recruit"
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB はテスト用のインメモリSQLiteデータベースを作成
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// テーブル作成
	schema := `
	CREATE TABLE IF NOT EXISTS recruits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		channel_id TEXT NOT NULL,
		message_id TEXT NOT NULL,
		author_id TEXT NOT NULL,
		max_capacity INTEGER NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS participants (
		recruit_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME,
		PRIMARY KEY (recruit_id, user_id),
		FOREIGN KEY (recruit_id) REFERENCES recruits(id) ON DELETE CASCADE
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	return db
}

func TestRecruitRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRecruitRepository(db)
	ctx := context.Background()

	state := &recruit.RecruitState{
		GuildID:     "guild-1",
		ChannelID:   "channel-1",
		MessageID:   "message-1",
		AuthorID:    "author-1",
		MaxCapacity: 5,
		Status:      recruit.RecruitStatusOpened,
		CreatedAt:   time.Now(),
	}

	id, err := repo.Create(ctx, state)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if id == 0 {
		t.Error("Create() returned invalid ID")
	}
}

func TestRecruitRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRecruitRepository(db)
	ctx := context.Background()

	// データを作成
	state := &recruit.RecruitState{
		GuildID:     "guild-1",
		ChannelID:   "channel-1",
		MessageID:   "message-1",
		AuthorID:    "author-1",
		MaxCapacity: 5,
		Status:      recruit.RecruitStatusOpened,
		CreatedAt:   time.Now(),
	}

	id, err := repo.Create(ctx, state)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 取得
	got, err := repo.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.ID != id {
		t.Errorf("Get() ID = %v, want %v", got.ID, id)
	}
	if got.GuildID != state.GuildID {
		t.Errorf("Get() GuildID = %v, want %v", got.GuildID, state.GuildID)
	}
	if got.ChannelID != state.ChannelID {
		t.Errorf("Get() ChannelID = %v, want %v", got.ChannelID, state.ChannelID)
	}
	if got.MessageID != state.MessageID {
		t.Errorf("Get() MessageID = %v, want %v", got.MessageID, state.MessageID)
	}
	if got.AuthorID != state.AuthorID {
		t.Errorf("Get() AuthorID = %v, want %v", got.AuthorID, state.AuthorID)
	}
	if got.MaxCapacity != state.MaxCapacity {
		t.Errorf("Get() MaxCapacity = %v, want %v", got.MaxCapacity, state.MaxCapacity)
	}
	if got.Status != state.Status {
		t.Errorf("Get() Status = %v, want %v", got.Status, state.Status)
	}
}

func TestRecruitRepository_GetByMessage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRecruitRepository(db)
	ctx := context.Background()

	state := &recruit.RecruitState{
		GuildID:     "guild-1",
		ChannelID:   "channel-1",
		MessageID:   "message-1",
		AuthorID:    "author-1",
		MaxCapacity: 5,
		Status:      recruit.RecruitStatusOpened,
		CreatedAt:   time.Now(),
	}

	id, err := repo.Create(ctx, state)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// メッセージIDで取得
	got, err := repo.GetByMessage(ctx, "channel-1", "message-1")
	if err != nil {
		t.Fatalf("GetByMessage() error = %v", err)
	}

	if got.ID != id {
		t.Errorf("GetByMessage() ID = %v, want %v", got.ID, id)
	}
	if got.ChannelID != state.ChannelID {
		t.Errorf("GetByMessage() ChannelID = %v, want %v", got.ChannelID, state.ChannelID)
	}
	if got.MessageID != state.MessageID {
		t.Errorf("GetByMessage() MessageID = %v, want %v", got.MessageID, state.MessageID)
	}
}

func TestRecruitRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRecruitRepository(db)
	ctx := context.Background()

	state := &recruit.RecruitState{
		GuildID:     "guild-1",
		ChannelID:   "channel-1",
		MessageID:   "message-1",
		AuthorID:    "author-1",
		MaxCapacity: 5,
		Status:      recruit.RecruitStatusOpened,
		CreatedAt:   time.Now(),
	}

	id, err := repo.Create(ctx, state)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 更新
	state.ID = id
	state.Status = recruit.RecruitStatusClosed
	state.MaxCapacity = 10

	err = repo.Update(ctx, state)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// 取得して確認
	got, err := repo.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Status != recruit.RecruitStatusClosed {
		t.Errorf("Update() Status = %v, want %v", got.Status, recruit.RecruitStatusClosed)
	}
	if got.MaxCapacity != 10 {
		t.Errorf("Update() MaxCapacity = %v, want 10", got.MaxCapacity)
	}
}

func TestRecruitRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRecruitRepository(db)
	ctx := context.Background()

	state := &recruit.RecruitState{
		GuildID:     "guild-1",
		ChannelID:   "channel-1",
		MessageID:   "message-1",
		AuthorID:    "author-1",
		MaxCapacity: 5,
		Status:      recruit.RecruitStatusOpened,
		CreatedAt:   time.Now(),
	}

	id, err := repo.Create(ctx, state)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 削除
	err = repo.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 取得してエラーになることを確認
	_, err = repo.Get(ctx, id)
	if err == nil {
		t.Error("Get() should return error after delete")
	}
}

func TestParticipantRepository_Upsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	recruitRepo := NewRecruitRepository(db)
	participantRepo := NewParticipantRepository(db)
	ctx := context.Background()

	// 募集を作成
	state := &recruit.RecruitState{
		GuildID:     "guild-1",
		ChannelID:   "channel-1",
		MessageID:   "message-1",
		AuthorID:    "author-1",
		MaxCapacity: 5,
		Status:      recruit.RecruitStatusOpened,
		CreatedAt:   time.Now(),
	}

	recruitID, err := recruitRepo.Create(ctx, state)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 参加者を追加
	err = participantRepo.Upsert(ctx, recruitID, "user-1", recruit.ParticipantStatusJoined)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	// 同じユーザーのステータスを更新
	err = participantRepo.Upsert(ctx, recruitID, "user-1", recruit.ParticipantStatusDeclined)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	// 確認
	participant, err := participantRepo.FindByRecruitAndUser(ctx, recruitID, "user-1")
	if err != nil {
		t.Fatalf("FindByRecruitAndUser() error = %v", err)
	}

	if participant.Status != recruit.ParticipantStatusDeclined {
		t.Errorf("Upsert() Status = %v, want %v", participant.Status, recruit.ParticipantStatusDeclined)
	}
}

func TestParticipantRepository_FindByRecruitAndUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	recruitRepo := NewRecruitRepository(db)
	participantRepo := NewParticipantRepository(db)
	ctx := context.Background()

	// 募集を作成
	state := &recruit.RecruitState{
		GuildID:     "guild-1",
		ChannelID:   "channel-1",
		MessageID:   "message-1",
		AuthorID:    "author-1",
		MaxCapacity: 5,
		Status:      recruit.RecruitStatusOpened,
		CreatedAt:   time.Now(),
	}

	recruitID, err := recruitRepo.Create(ctx, state)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 参加者を追加
	err = participantRepo.Upsert(ctx, recruitID, "user-1", recruit.ParticipantStatusJoined)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	// 存在する参加者を取得
	participant, err := participantRepo.FindByRecruitAndUser(ctx, recruitID, "user-1")
	if err != nil {
		t.Fatalf("FindByRecruitAndUser() error = %v", err)
	}

	if participant == nil {
		t.Fatal("FindByRecruitAndUser() returned nil")
	}

	if participant.UserID != "user-1" {
		t.Errorf("FindByRecruitAndUser() UserID = %v, want user-1", participant.UserID)
	}

	// 存在しない参加者を取得
	participant, err = participantRepo.FindByRecruitAndUser(ctx, recruitID, "user-999")
	if err != nil {
		t.Fatalf("FindByRecruitAndUser() error = %v", err)
	}

	if participant != nil {
		t.Error("FindByRecruitAndUser() should return nil for non-existent user")
	}
}

func TestParticipantRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	recruitRepo := NewRecruitRepository(db)
	participantRepo := NewParticipantRepository(db)
	ctx := context.Background()

	// 募集を作成
	state := &recruit.RecruitState{
		GuildID:     "guild-1",
		ChannelID:   "channel-1",
		MessageID:   "message-1",
		AuthorID:    "author-1",
		MaxCapacity: 5,
		Status:      recruit.RecruitStatusOpened,
		CreatedAt:   time.Now(),
	}

	recruitID, err := recruitRepo.Create(ctx, state)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 複数の参加者を追加
	err = participantRepo.Upsert(ctx, recruitID, "user-1", recruit.ParticipantStatusJoined)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	err = participantRepo.Upsert(ctx, recruitID, "user-2", recruit.ParticipantStatusJoined)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	err = participantRepo.Upsert(ctx, recruitID, "user-3", recruit.ParticipantStatusDeclined)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	// リスト取得
	participants, err := participantRepo.List(ctx, recruitID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(participants) != 3 {
		t.Errorf("List() length = %v, want 3", len(participants))
	}

	// ステータス確認
	joinedCount := 0
	declinedCount := 0
	for _, p := range participants {
		switch p.Status {
		case recruit.ParticipantStatusJoined:
			joinedCount++
		case recruit.ParticipantStatusDeclined:
			declinedCount++
		}
	}

	if joinedCount != 2 {
		t.Errorf("List() joined count = %v, want 2", joinedCount)
	}
	if declinedCount != 1 {
		t.Errorf("List() declined count = %v, want 1", declinedCount)
	}
}

func TestParticipantRepository_DeleteAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	recruitRepo := NewRecruitRepository(db)
	participantRepo := NewParticipantRepository(db)
	ctx := context.Background()

	// 募集を作成
	state := &recruit.RecruitState{
		GuildID:     "guild-1",
		ChannelID:   "channel-1",
		MessageID:   "message-1",
		AuthorID:    "author-1",
		MaxCapacity: 5,
		Status:      recruit.RecruitStatusOpened,
		CreatedAt:   time.Now(),
	}

	recruitID, err := recruitRepo.Create(ctx, state)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 参加者を追加
	err = participantRepo.Upsert(ctx, recruitID, "user-1", recruit.ParticipantStatusJoined)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	err = participantRepo.Upsert(ctx, recruitID, "user-2", recruit.ParticipantStatusJoined)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	// 全削除
	err = participantRepo.DeleteAll(ctx, recruitID)
	if err != nil {
		t.Fatalf("DeleteAll() error = %v", err)
	}

	// リスト取得して空であることを確認
	participants, err := participantRepo.List(ctx, recruitID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(participants) != 0 {
		t.Errorf("List() length = %v, want 0 after DeleteAll", len(participants))
	}
}
