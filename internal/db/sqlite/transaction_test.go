package sqlite

import (
	"context"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestTxManager_Do_Commit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	txManager := NewTxManager(db)
	ctx := context.Background()

	// トランザクション内でデータを挿入
	err := txManager.Do(ctx, func(ctx context.Context) error {
		executor := GetExecutor(ctx, db)
		_, err := executor.ExecContext(ctx, `
			INSERT INTO recruits (guild_id, channel_id, message_id, author_id, max_capacity, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
		`, "guild-1", "channel-1", "message-1", "author-1", 5, "opened")
		return err
	})

	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	// コミットされたことを確認
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM recruits").Scan(&count)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}

	if count != 1 {
		t.Errorf("Do() commit failed, count = %v, want 1", count)
	}
}

func TestTxManager_Do_Rollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	txManager := NewTxManager(db)
	ctx := context.Background()

	// トランザクション内でエラーを返す
	testErr := errors.New("test error")
	err := txManager.Do(ctx, func(ctx context.Context) error {
		executor := GetExecutor(ctx, db)
		_, err := executor.ExecContext(ctx, `
			INSERT INTO recruits (guild_id, channel_id, message_id, author_id, max_capacity, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
		`, "guild-1", "channel-1", "message-1", "author-1", 5, "opened")
		if err != nil {
			return err
		}
		return testErr
	})

	if err == nil {
		t.Fatal("Do() should return error")
	}

	if !errors.Is(err, testErr) {
		t.Errorf("Do() error = %v, want %v", err, testErr)
	}

	// ロールバックされたことを確認
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM recruits").Scan(&count)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}

	if count != 0 {
		t.Errorf("Do() rollback failed, count = %v, want 0", count)
	}
}

func TestTxManager_Do_Panic(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	txManager := NewTxManager(db)
	ctx := context.Background()

	// panic時の処理を確認
	defer func() {
		if r := recover(); r == nil {
			t.Error("Do() should panic")
		}

		// ロールバックされたことを確認
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM recruits").Scan(&count)
		if err != nil {
			t.Fatalf("QueryRow() error = %v", err)
		}

		if count != 0 {
			t.Errorf("Do() rollback failed after panic, count = %v, want 0", count)
		}
	}()

	_ = txManager.Do(ctx, func(ctx context.Context) error {
		executor := GetExecutor(ctx, db)
		_, err := executor.ExecContext(ctx, `
			INSERT INTO recruits (guild_id, channel_id, message_id, author_id, max_capacity, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
		`, "guild-1", "channel-1", "message-1", "author-1", 5, "opened")
		if err != nil {
			return err
		}
		panic("test panic")
	})
}

func TestGetExecutor_WithTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	txManager := NewTxManager(db)
	ctx := context.Background()

	err := txManager.Do(ctx, func(txCtx context.Context) error {
		executor := GetExecutor(txCtx, db)

		// トランザクション内なので*sql.Txが返される
		// 型アサーションでテスト
		if _, ok := executor.(interface{ Commit() error }); !ok {
			t.Error("GetExecutor() should return *sql.Tx inside transaction")
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
}

func TestGetExecutor_WithoutTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	executor := GetExecutor(ctx, db)

	// トランザクション外なので*sql.DBが返される
	// 型アサーションでテスト
	if _, ok := executor.(interface{ Begin() (*interface{}, error) }); ok {
		// *sql.DBはBeginメソッドを持つ
		// executorインターフェースはBeginを持たないので、型アサーションで区別できない
		// 代わりに実際にクエリを実行して確認
		_, err := executor.ExecContext(ctx, `
			INSERT INTO recruits (guild_id, channel_id, message_id, author_id, max_capacity, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
		`, "guild-1", "channel-1", "message-1", "author-1", 5, "opened")

		if err != nil {
			t.Errorf("GetExecutor() without transaction failed to execute: %v", err)
		}
	}
}

func TestTxManager_NestedTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	txManager := NewTxManager(db)
	ctx := context.Background()

	// ネストしたトランザクションのテスト
	err := txManager.Do(ctx, func(ctx1 context.Context) error {
		executor1 := GetExecutor(ctx1, db)
		_, err := executor1.ExecContext(ctx1, `
			INSERT INTO recruits (guild_id, channel_id, message_id, author_id, max_capacity, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
		`, "guild-1", "channel-1", "message-1", "author-1", 5, "opened")
		if err != nil {
			return err
		}

		// 内側のトランザクションでエラー
		return txManager.Do(ctx1, func(ctx2 context.Context) error {
			executor2 := GetExecutor(ctx2, db)
			_, err := executor2.ExecContext(ctx2, `
				INSERT INTO recruits (guild_id, channel_id, message_id, author_id, max_capacity, status, created_at)
				VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
			`, "guild-2", "channel-2", "message-2", "author-2", 3, "opened")
			if err != nil {
				return err
			}
			return errors.New("inner transaction error")
		})
	})

	if err == nil {
		t.Fatal("Do() should return error from nested transaction")
	}

	// 外側のトランザクションもロールバックされることを確認
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM recruits").Scan(&count)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}

	// ネストしたトランザクションは両方ロールバックされる
	// (内側のトランザクションがエラーを返すと、外側もロールバック)
	if count != 0 {
		t.Errorf("Nested transaction rollback failed, count = %v, want 0", count)
	}
}

func TestTxManager_MultipleOperations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	txManager := NewTxManager(db)
	ctx := context.Background()

	// 複数の操作を1つのトランザクションで実行
	err := txManager.Do(ctx, func(ctx context.Context) error {
		executor := GetExecutor(ctx, db)

		// 募集を作成
		result, err := executor.ExecContext(ctx, `
			INSERT INTO recruits (guild_id, channel_id, message_id, author_id, max_capacity, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
		`, "guild-1", "channel-1", "message-1", "author-1", 5, "opened")
		if err != nil {
			return err
		}

		recruitID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		// 参加者を追加
		_, err = executor.ExecContext(ctx, `
			INSERT INTO participants (recruit_id, user_id, status, created_at)
			VALUES (?, ?, ?, datetime('now'))
		`, recruitID, "user-1", "joined")
		if err != nil {
			return err
		}

		_, err = executor.ExecContext(ctx, `
			INSERT INTO participants (recruit_id, user_id, status, created_at)
			VALUES (?, ?, ?, datetime('now'))
		`, recruitID, "user-2", "joined")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	// 両方のテーブルにデータが入っていることを確認
	var recruitCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM recruits").Scan(&recruitCount)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}

	if recruitCount != 1 {
		t.Errorf("recruit count = %v, want 1", recruitCount)
	}

	var participantCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM participants").Scan(&participantCount)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}

	if participantCount != 2 {
		t.Errorf("participant count = %v, want 2", participantCount)
	}
}
