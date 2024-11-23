package postgres

import (
	"context"
	"fmt"
	"github.com/arrowwhi/go-utils/postgres/config"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/jackc/pgx/v5"
)

type Database struct {
	Pool *pgxpool.Pool
}

// NewDatabase создает новое подключение к базе данных с базовыми настройками и опциями
func NewDatabase(cfg config.DBConfig, opts ...Option) (*Database, error) {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Применение опций
	for _, opt := range opts {
		opt(config)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &Database{Pool: pool}, nil
}

// Close закрывает пул соединений к базе данных
func (db *Database) Close() {
	db.Pool.Close()
}
