package pg

import (
	"context"
	"fmt"

	"restaurant-system/internal/kitchen/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkerRepository struct {
	db *pgxpool.Pool
}

func NewWorkerRepository(db *pgxpool.Pool) *WorkerRepository {
	return &WorkerRepository{db: db}
}

func (r *WorkerRepository) CreateOrUpdateWorker(ctx context.Context, name string, workerType string, orderTypes []string) (*model.Worker, error) {
	query := `
        INSERT INTO workers (name, type, status, last_seen, orders_processed)
        VALUES ($1, $2, 'online', NOW(), 0)
        ON CONFLICT (name) 
        DO UPDATE SET 
            type = EXCLUDED.type,
            status = 'online',
            last_seen = NOW()
        RETURNING id, name, type, status, last_seen, orders_processed
    `

	var worker model.Worker
	err := r.db.QueryRow(ctx, query, name, workerType).Scan(
		&worker.ID,
		&worker.Name,
		&worker.Type,
		&worker.Status,
		&worker.LastSeen,
		&worker.OrdersProcessed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update worker: %w", err)
	}

	worker.OrderTypes = orderTypes
	return &worker, nil
}

func (r *WorkerRepository) UpdateWorkerHeartbeat(ctx context.Context, id int) error {
	query := `UPDATE workers SET last_seen = NOW(), status = 'online' WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *WorkerRepository) MarkWorkerOffline(ctx context.Context, id int) error {
	query := `UPDATE workers SET status = 'offline', last_seen = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *WorkerRepository) IncrementOrdersProcessed(ctx context.Context, id int) error {
	query := `UPDATE workers SET orders_processed = orders_processed + 1 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderNumber string, status string, processedBy string) error {
	query := `UPDATE orders SET status = $1, processed_by = $2, updated_at = NOW() WHERE number = $3`
	_, err := r.db.Exec(ctx, query, status, processedBy, orderNumber)
	return err
}

func (r *OrderRepository) CreateStatusLog(ctx context.Context, orderNumber string, status string, changedBy string, notes *string) error {
	query := `
		INSERT INTO order_status_log (order_id, status, changed_by, changed_at, notes)
		SELECT id, $1, $2, NOW(), $3 FROM orders WHERE number = $4
	`
	_, err := r.db.Exec(ctx, query, status, changedBy, notes, orderNumber)
	return err
}

func (r *OrderRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	query := `
		SELECT id, number, customer_name, type, table_number, delivery_address, 
			   total_amount, priority, status, processed_by, completed_at
		FROM orders WHERE number = $1
	`

	var order model.Order
	err := r.db.QueryRow(ctx, query, orderNumber).Scan(
		&order.ID,
		&order.Number,
		&order.CustomerName,
		&order.Type,
		&order.TableNumber,
		&order.DeliveryAddress,
		&order.TotalAmount,
		&order.Priority,
		&order.Status,
		&order.ProcessedBy,
		&order.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

func (r *OrderRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.BeginTx(ctx, pgx.TxOptions{})
}
