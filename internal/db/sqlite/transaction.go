package sqlite

import (
	"at-bot/internal/uow"
	"context"
	"database/sql"
	"fmt"
)

type txKey struct{}

type txManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) uow.UnitOfWork {
	return &txManager{
		db: db,
	}
}

func (m *txManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	var fnErr error
	defer func() {
		if p := recover(); p != nil {
			// panicが発生した場合はロールバック
			_ = tx.Rollback()
			// panicを再度投げる
			panic(p)
		} else if fnErr != nil {
			// エラーが発生した場合はロールバック
			_ = tx.Rollback()
		} else {
			// 正常終了の場合はコミット
			fnErr = tx.Commit()
			if fnErr != nil {
				fnErr = fmt.Errorf("failed to commit transaction: %w", fnErr)
			}
		}
	}()

	fnErr = fn(txCtx)
	return fnErr
}

// getTx はcontextからトランザクションを取得する
// トランザクションが存在しない場合はnilを返す
func getTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return nil
}

// getExecutor はcontextからトランザクションまたはDBを取得
// Repository実装でトランザクション対応するために使用
func GetExecutor(ctx context.Context, db *sql.DB) executor {
	if tx := getTx(ctx); tx != nil {
		return tx
	}
	return db
}

// executor はsql.DBとsql.Txの共通インターフェース
type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
