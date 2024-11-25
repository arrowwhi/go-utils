package postgres

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/arrowwhi/go-utils/postgres/db_config"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/jackc/pgx/v5"
)

type Database struct {
	Pool    *pgxpool.Pool
	Builder squirrel.StatementBuilderType
}

// NewDatabase создает новое подключение к базе данных с базовыми настройками и опциями
func NewDatabase(cfg db_config.DBConfig, opts ...Option) (*Database, error) {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	pgConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Применение опций
	for _, opt := range opts {
		opt(pgConfig)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		return nil, err
	}

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return &Database{Pool: pool, Builder: builder}, nil
}

// Close закрывает пул соединений к базе данных
func (db *Database) Close() {
	db.Pool.Close()
}
