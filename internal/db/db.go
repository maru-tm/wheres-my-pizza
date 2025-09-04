package db

import (
	"context"
	"fmt"
	"log"
	"restaurant-system/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("PostgreSQL не отвечает: %w", err)
	}

	log.Println("✅ Подключение к PostgreSQL успешно")
	return &DB{Pool: pool}, nil
}

func (d *DB) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}
