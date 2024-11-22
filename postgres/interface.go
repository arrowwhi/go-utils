package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBInterface определяет стандартные методы работы с базой данных
type DBInterface interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	BeginTransaction(ctx context.Context) (context.Context, error)
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error
}

// Реализация методов интерфейса

func (db *Database) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if tx, ok := GetTransaction(ctx); ok {
		return tx.Query(ctx, sql, args...)
	}
	return db.Pool.Query(ctx, sql, args...)
}

func (db *Database) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if tx, ok := GetTransaction(ctx); ok {
		return tx.Exec(ctx, sql, args...)
	}
	return db.Pool.Exec(ctx, sql, args...)
}
