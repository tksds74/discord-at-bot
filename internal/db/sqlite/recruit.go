package sqlite

import (
	"at-bot/internal/recruit"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type sqliteRecruitRepository struct {
	db *sql.DB
}

func NewRecruitRepository(db *sql.DB) recruit.RecruitRepository {
	return &sqliteRecruitRepository{
		db: db,
	}
}

func (r *sqliteRecruitRepository) Get(ctx context.Context, id recruit.RecruitID) (*recruit.RecruitState, error) {
	executor := GetExecutor(ctx, r.db)

	query := `
		SELECT id, guild_id, channel_id, message_id, author_id, max_capacity, status, created_at, updated_at
		FROM recruits
		WHERE id = ?
	`

	var state recruit.RecruitState
	var updatedAt sql.NullTime
	err := executor.
		QueryRowContext(ctx, query, id).
		Scan(
			&state.ID,
			&state.GuildID,
			&state.ChannelID,
			&state.MessageID,
			&state.AuthorID,
			&state.MaxCapacity,
			&state.Status,
			&state.CreatedAt,
			&updatedAt,
		)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("recruit not found: %d", id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get recruit: %w", err)
	}

	if updatedAt.Valid {
		state.UpdatedAt = &updatedAt.Time
	}

	return &state, nil
}

func (r *sqliteRecruitRepository) GetByMessage(
	ctx context.Context,
	channelID recruit.ChannelID,
	messageID recruit.MessageID,
) (*recruit.RecruitState, error) {
	executor := GetExecutor(ctx, r.db)

	query := `
		SELECT id, guild_id, channel_id, message_id, author_id, max_capacity, status, created_at, updated_at
		FROM recruits
		WHERE channel_id = ? AND message_id = ?
	`

	var state recruit.RecruitState
	var updatedAt sql.NullTime
	err := executor.
		QueryRowContext(ctx, query, channelID, messageID).
		Scan(
			&state.ID,
			&state.GuildID,
			&state.ChannelID,
			&state.MessageID,
			&state.AuthorID,
			&state.MaxCapacity,
			&state.Status,
			&state.CreatedAt,
			&updatedAt,
		)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("recruit not found: %s, %s", channelID, messageID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get recruit: %w", err)
	}

	if updatedAt.Valid {
		state.UpdatedAt = &updatedAt.Time
	}

	return &state, nil
}

func (r *sqliteRecruitRepository) Create(ctx context.Context, state *recruit.RecruitState) (recruit.RecruitID, error) {
	executor := GetExecutor(ctx, r.db)

	query := `
		INSERT INTO recruits (guild_id, channel_id, message_id, author_id, max_capacity, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := executor.ExecContext(
		ctx,
		query,
		state.GuildID,
		state.ChannelID,
		state.MessageID,
		state.AuthorID,
		state.MaxCapacity,
		state.Status,
		state.CreatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create recruit: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return recruit.RecruitID(id), nil
}

func (r *sqliteRecruitRepository) Update(ctx context.Context, state *recruit.RecruitState) error {
	executor := GetExecutor(ctx, r.db)

	now := time.Now()
	query := `
		UPDATE recruits
		SET guild_id = ?, channel_id = ?, message_id = ?, author_id = ?,
		    max_capacity = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := executor.ExecContext(
		ctx,
		query,
		state.GuildID,
		state.ChannelID,
		state.MessageID,
		state.AuthorID,
		state.MaxCapacity,
		state.Status,
		now,
		state.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update recruit: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("recruit not found: %d", state.ID)
	}

	return nil
}

func (r *sqliteRecruitRepository) Delete(ctx context.Context, id recruit.RecruitID) error {
	executor := GetExecutor(ctx, r.db)

	query := `DELETE FROM recruits WHERE id = ?`

	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete recruit: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("recruit not found: %d", id)
	}

	return nil
}

type sqliteParticipantRepository struct {
	db *sql.DB
}

func NewParticipantRepository(db *sql.DB) recruit.ParticipantRepository {
	return &sqliteParticipantRepository{
		db: db,
	}
}

func (r *sqliteParticipantRepository) Upsert(ctx context.Context, recruitID recruit.RecruitID, userID recruit.UserID, status recruit.ParticipantStatus) error {
	executor := GetExecutor(ctx, r.db)

	now := time.Now()
	query := `
		INSERT INTO participants (recruit_id, user_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(recruit_id, user_id) DO UPDATE SET
			status = excluded.status,
			updated_at = excluded.updated_at
	`

	_, err := executor.ExecContext(ctx, query, recruitID, userID, status, now, now)
	if err != nil {
		return fmt.Errorf("failed to upsert participant: %w", err)
	}

	return nil
}

func (r *sqliteParticipantRepository) FindByRecruitAndUser(ctx context.Context, recruitID recruit.RecruitID, userID recruit.UserID) (*recruit.Participant, error) {
	executor := GetExecutor(ctx, r.db)

	query := `
		SELECT recruit_id, user_id, status, created_at, updated_at
		FROM participants
		WHERE recruit_id = ? AND user_id = ?
	`
	var p recruit.Participant
	var updatedAt sql.NullTime
	err := executor.
		QueryRowContext(ctx, query, recruitID, userID).
		Scan(
			&p.RecruitID,
			&p.UserID,
			&p.Status,
			&p.CreatedAt,
			&updatedAt,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	if updatedAt.Valid {
		p.UpdatedAt = &updatedAt.Time
	}

	return &p, nil
}

func (r *sqliteParticipantRepository) List(ctx context.Context, recruitID recruit.RecruitID) ([]recruit.Participant, error) {
	executor := GetExecutor(ctx, r.db)

	query := `
		SELECT recruit_id, user_id, status, created_at, updated_at
		FROM participants
		WHERE recruit_id = ?
		ORDER BY created_at ASC
	`

	rows, err := executor.QueryContext(ctx, query, recruitID)
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}
	defer rows.Close()

	var participants []recruit.Participant
	for rows.Next() {
		var p recruit.Participant
		var updatedAt sql.NullTime

		err := rows.Scan(
			&p.RecruitID,
			&p.UserID,
			&p.Status,
			&p.CreatedAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		if updatedAt.Valid {
			p.UpdatedAt = &updatedAt.Time
		}

		participants = append(participants, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate participants: %w", err)
	}

	return participants, nil
}

func (r *sqliteParticipantRepository) DeleteAll(ctx context.Context, recruitID recruit.RecruitID) error {
	executor := GetExecutor(ctx, r.db)

	query := `
		DELETE FROM participants
		WHERE recruit_id = ?
	`

	result, err := executor.ExecContext(ctx, query, recruitID)
	if err != nil {
		return fmt.Errorf("failed to delete participants: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("participants not found: %d", recruitID)
	}

	return nil
}
