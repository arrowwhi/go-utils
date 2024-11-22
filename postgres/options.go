package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Option функция для настройки базы данных
type Option func(*pgxpool.Config)

// WithMaxConns устанавливает максимальное количество соединений в пуле
func WithMaxConns(maxConns int32) Option {
	return func(config *pgxpool.Config) {
		config.MaxConns = maxConns
	}
}

// WithMinConns устанавливает минимальное количество соединений в пуле
func WithMinConns(minConns int32) Option {
	return func(config *pgxpool.Config) {
		config.MinConns = minConns
	}
}

// WithMaxConnLifetime устанавливает максимальное время жизни соединения
func WithMaxConnLifetime(d time.Duration) Option {
	return func(config *pgxpool.Config) {
		config.MaxConnLifetime = d
	}
}

// WithHealthCheckPeriod устанавливает период проверки здоровья соединений
func WithHealthCheckPeriod(d time.Duration) Option {
	return func(config *pgxpool.Config) {
		config.HealthCheckPeriod = d
	}
}
