package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type txKeyType struct{}

var txKey = txKeyType{}

// BeginTransaction начинает транзакцию внутри контекста
func (db *Database) BeginTransaction(ctx context.Context) (context.Context, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return WithTransaction(ctx, tx), nil
}

func (db *Database) CommitTransaction(ctx context.Context) error {
	tx, ok := GetTransaction(ctx)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}
	return tx.Commit(ctx)
}

func (db *Database) RollbackTransaction(ctx context.Context) error {
	tx, ok := GetTransaction(ctx)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}
	return tx.Rollback(ctx)
}

// WithTransaction сохраняет транзакцию в контексте
func WithTransaction(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// GetTransaction извлекает транзакцию из контекста
func GetTransaction(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}
