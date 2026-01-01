package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}
	// データベース接続
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	// 外部キー制約を有効化
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	// テーブル作成
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	-- 募集テーブル
	CREATE TABLE IF NOT EXISTS recruits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL,
		channel_id TEXT NOT NULL,
		message_id TEXT NOT NULL,
		author_id TEXT NOT NULL,
		max_capacity INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'opened',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP
	);

	-- 参加者テーブル
	CREATE TABLE IF NOT EXISTS participants (
		recruit_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP,
		PRIMARY KEY (recruit_id, user_id),
		FOREIGN KEY (recruit_id) REFERENCES recruits(id) ON DELETE CASCADE
	);

	-- インデックス
	CREATE INDEX IF NOT EXISTS idx_recruits_status ON recruits(status);
	CREATE INDEX IF NOT EXISTS idx_recruits_guild_id ON recruits(guild_id);
	CREATE INDEX IF NOT EXISTS idx_participants_recruit_id ON participants(recruit_id);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}
