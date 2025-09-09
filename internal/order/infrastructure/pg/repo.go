package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"restaurant-system/internal/order/model"
	"time"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.BeginTx(ctx, pgx.TxOptions{})
}

func (r *OrderRepository) CreateOrder(ctx context.Context, tx pgx.Tx, order *model.Order) (int, error) {
	query := `
		INSERT INTO orders
			(number, customer_name, type, table_number, delivery_address, total_amount, priority, status)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int
	err := tx.QueryRow(ctx,
		query,
		order.Number,
		order.CustomerName,
		order.Type,
		order.TableNumber,
		order.DeliveryAddress,
		order.TotalAmount,
		order.Priority,
		order.Status,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	return id, nil
}

func (r *OrderRepository) CreateItem(ctx context.Context, tx pgx.Tx, item *model.OrderItem) (int, error) {
	query := `
		INSERT INTO order_items
			(order_id, name, quantity, price)
		VALUES
			($1, $2, $3, $4)
		RETURNING id
	`

	var id int
	err := tx.QueryRow(ctx,
		query,
		item.OrderID,
		item.Name,
		item.Quantity,
		item.Price,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create order item: %w", err)
	}

	return id, nil
}

func (r *OrderRepository) CreateLog(ctx context.Context, tx pgx.Tx, logEntry *model.OrderStatusLog) (int, error) {
	query := `
		INSERT INTO order_status_log
			(order_id, status, changed_by, notes)
		VALUES
			($1, $2, $3, $4)
		RETURNING id
	`

	var id int
	err := tx.QueryRow(ctx,
		query,
		logEntry.OrderID,
		logEntry.Status,
		logEntry.ChangedBy,
		logEntry.Notes,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create order status log: %w", err)
	}

	return id, nil
}

func (r *OrderRepository) GetNextOrderSequence(ctx context.Context, tx pgx.Tx, date string) (int, error) {
	var seq int

	query := `
        INSERT INTO order_sequences (seq_date, last_seq)
        VALUES ($1, 1)
        ON CONFLICT (seq_date)
        DO UPDATE SET last_seq = order_sequences.last_seq + 1
        RETURNING last_seq
    `

	err := tx.QueryRow(ctx, query, date).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("failed to get next order sequence: %w", err)
	}

	return seq, nil
}

func (r *OrderRepository) GetByNumber(ctx context.Context, tx pgx.Tx, orderNumber string) (*model.Order, error) {
	query := `SELECT id, number, customer_name, type, table_number, delivery_address, 
	                 total_amount, priority, status, processed_by, completed_at, 
	                 created_at, updated_at
	          FROM orders WHERE number = $1`

	var order model.Order
	err := tx.QueryRow(ctx, query, orderNumber).Scan(
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
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order by number: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) GetStatus(ctx context.Context, orderNumber string) (model.OrderStatus, error) {
	var status model.OrderStatus

	query := `SELECT status FROM orders WHERE number = $1`

	err := r.db.QueryRow(ctx, query, orderNumber).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("order with id %d not found", orderNumber)
		}
		return "", fmt.Errorf("failed to get order status: %w", err)
	}

	return status, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, tx pgx.Tx, orderNumber string, status model.OrderStatus, workerName string) error {
	query := `
		UPDATE orders
		SET status = $1,
		    processed_by = $2,
		    completed_at = CASE WHEN $1 = 'ready' THEN $3 ELSE completed_at END,
		    updated_at = now()
		WHERE id = $4
	`
	_, err := tx.Exec(ctx, query, status, workerName, time.Now(), orderNumber)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

func (r *OrderRepository) InsertStatusLog(ctx context.Context, tx pgx.Tx, log model.OrderStatusLog) error {
	query := `
		INSERT INTO order_status_log (order_id, status, changed_by, notes)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.Exec(ctx, query,
		log.OrderID,
		log.Status,
		log.ChangedBy,
		log.Notes,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order status log: %w", err)
	}
	return nil
}
