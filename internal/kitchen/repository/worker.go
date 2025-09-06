package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"

	"github.com/jackc/pgx/v5"
)

type WorkerRepository struct {
	db *pgxpool.Pool
}

func NewWorkerRepository(pool *pgxpool.Pool) *WorkerRepository {
	return &WorkerRepository{db: pool}
}

func (r *WorkerRepository) FindStatus(ctx context.Context, name string) (string, error) {
	var status string
	err := r.db.QueryRow(ctx,
		`SELECT status FROM workers WHERE name = $1`,
		name,
	).Scan(&status)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return status, err
}

func (r *WorkerRepository) Insert(ctx context.Context, name, workerType string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO workers (name, type, status, last_seen) 
		 VALUES ($1, $2, 'online', $3)`,
		name, workerType, time.Now(),
	)
	return err
}

func (r *WorkerRepository) UpdateOnline(ctx context.Context, name string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE workers 
		 SET status = 'online', last_seen = $2 
		 WHERE name = $1`,
		name, time.Now(),
	)
	return err
}

func (r *WorkerRepository) IncrementOrdersProcessed(ctx context.Context, name string) error {
	query := `
		UPDATE workers
		SET orders_processed = orders_processed + 1,
		    updated_at = now()
		WHERE name = $1
	`
	_, err := r.db.Exec(ctx, query, name)
	if err != nil {
		return fmt.Errorf("failed to increment orders_processed for worker %s: %w", name, err)
	}
	return nil
}
