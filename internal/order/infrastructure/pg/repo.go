package pg

import (
	"context"
	"fmt"

	"restaurant-system/internal/order/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

func (r *OrderRepository) CreateOrder(ctx context.Context, tx pgx.Tx, order *model.Order) (int, error) {
	query := `
		INSERT INTO orders (
			number, customer_name, type, table_number, delivery_address,
			total_amount, priority, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var id int
	err := tx.QueryRow(ctx, query,
		order.Number,
		order.CustomerName,
		string(order.Type),
		order.TableNumber,
		order.DeliveryAddress,
		order.TotalAmount,
		order.Priority,
		string(order.Status),
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	return id, nil
}

func (r *OrderRepository) CreateItem(ctx context.Context, tx pgx.Tx, item *model.OrderItem) (int, error) {
	query := `
		INSERT INTO order_items (order_id, name, quantity, price, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int
	err := tx.QueryRow(ctx, query,
		item.OrderID,
		item.Name,
		item.Quantity,
		item.Price,
		item.CreatedAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create order item: %w", err)
	}

	return id, nil
}

func (r *OrderRepository) CreateLog(ctx context.Context, tx pgx.Tx, logEntry *model.OrderStatusLog) (int, error) {
	query := `
		INSERT INTO order_status_log (order_id, status, changed_by, changed_at, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int
	err := tx.QueryRow(ctx, query,
		logEntry.OrderID,
		string(logEntry.Status),
		logEntry.ChangedBy,
		logEntry.ChangedAt,
		logEntry.Notes,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create status log: %w", err)
	}

	return id, nil
}

func (r *OrderRepository) GetNextOrderSequence(ctx context.Context, tx pgx.Tx, date string) (int, error) {
	query := `
		INSERT INTO order_sequences (seq_date, last_seq)
		VALUES ($1, 1)
		ON CONFLICT (seq_date) 
		DO UPDATE SET last_seq = order_sequences.last_seq + 1
		RETURNING last_seq
	`

	var seq int
	err := tx.QueryRow(ctx, query, date).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence: %w", err)
	}

	return seq, nil
}

func (r *OrderRepository) GetOrders(ctx context.Context, page, limit int) ([]*model.Order, int, error) {
	offset := (page - 1) * limit
	query := `
		SELECT id, number, customer_name, type, table_number, delivery_address, 
			total_amount, priority, status, processed_by, completed_at, created_at, updated_at
		FROM orders 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID, &order.Number, &order.CustomerName, &order.Type, &order.TableNumber, &order.DeliveryAddress,
			&order.TotalAmount, &order.Priority, &order.Status, &order.ProcessedBy, &order.CompletedAt, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
	}

	countQuery := `SELECT COUNT(*) FROM orders`
	var total int
	err = r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	return orders, total, nil
}

func (r *OrderRepository) GetOrder(ctx context.Context, orderNumber string) (*model.Order, error) {
	query := `
		SELECT id, number, customer_name, type, table_number, delivery_address, 
			total_amount, priority, status, processed_by, completed_at, created_at, updated_at
		FROM orders 
		WHERE number = $1
	`
	var order model.Order
	err := r.db.QueryRow(ctx, query, orderNumber).Scan(
		&order.ID, &order.Number, &order.CustomerName, &order.Type, &order.TableNumber, &order.DeliveryAddress,
		&order.TotalAmount, &order.Priority, &order.Status, &order.ProcessedBy, &order.CompletedAt, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	itemsQuery := `SELECT id, name, quantity, price, created_at FROM order_items WHERE order_id = $1`
	itemsRows, err := r.db.Query(ctx, itemsQuery, order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer itemsRows.Close()

	for itemsRows.Next() {
		var item model.OrderItem
		err := itemsRows.Scan(&item.ID, &item.Name, &item.Quantity, &item.Price, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}
