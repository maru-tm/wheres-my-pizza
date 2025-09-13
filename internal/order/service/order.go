package service

import (
	"context"
	"fmt"
	"time"

	"restaurant-system/internal/order/model"
	"restaurant-system/pkg/logger"

	"github.com/jackc/pgx/v5"
)

type OrderRepository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	CreateOrder(ctx context.Context, tx pgx.Tx, order *model.Order) (int, error)
	CreateItem(ctx context.Context, tx pgx.Tx, item *model.OrderItem) (int, error)
	CreateLog(ctx context.Context, tx pgx.Tx, logEntry *model.OrderStatusLog) (int, error)
	GetNextOrderSequence(ctx context.Context, tx pgx.Tx, date string) (int, error)
}

type OrderPublisher interface {
	PublishCreatedOrder(ctx context.Context, order *model.Order) error
}

type OrderService struct {
	repo OrderRepository
	rmq  OrderPublisher
}

func NewOrderService(r OrderRepository, rmq OrderPublisher) *OrderService {
	return &OrderService{repo: r, rmq: rmq}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	rid := ""
	if v := ctx.Value("request_id"); v != nil {
		if str, ok := v.(string); ok {
			rid = str
		}
	}
	if rid == "" {
		rid = fmt.Sprintf("req-%d", time.Now().UnixNano())
	}

	if err := order.Validate(); err != nil {
		logger.Log(logger.ERROR, "order-service", "validation_failed", "order validation failed", rid,
			map[string]interface{}{"error": err.Error()}, err)
		return nil, model.ValidationError
	}

	// Calculate total amount
	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	order.TotalAmount = total

	// Set priority based on total amount
	switch {
	case total > 100:
		order.Priority = 10
	case total >= 50:
		order.Priority = 5
	default:
		order.Priority = 1
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "db_transaction_failed", "failed to begin transaction", rid,
			map[string]interface{}{"error": err.Error()}, err)
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Generate order number
	today := time.Now().UTC().Format("20060102")
	seq, err := s.repo.GetNextOrderSequence(ctx, tx, today)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "sequence_fetch_failed", "failed to get next order sequence", rid,
			map[string]interface{}{"error": err.Error()}, err)
		return nil, fmt.Errorf("failed to get next order sequence: %w", err)
	}

	orderNumber := fmt.Sprintf("ORD_%s_%03d", today, seq)
	order.Number = orderNumber
	order.Status = model.StatusReceived
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	// Create order
	orderID, err := s.repo.CreateOrder(ctx, tx, order)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "db_insert_failed", "failed to insert order", rid,
			map[string]interface{}{"order_number": order.Number, "error": err.Error()}, err)
		return nil, err
	}
	order.ID = orderID

	// Create order items
	for i := range order.Items {
		order.Items[i].OrderID = orderID
		order.Items[i].CreatedAt = time.Now()
		itemID, err := s.repo.CreateItem(ctx, tx, &order.Items[i])
		if err != nil {
			logger.Log(logger.ERROR, "order-service", "db_insert_failed", "failed to insert order item", rid,
				map[string]interface{}{"order_number": order.Number, "item_name": order.Items[i].Name, "error": err.Error()}, err)
			return nil, err
		}
		order.Items[i].ID = itemID
	}

	// Create status log
	logEntry := &model.OrderStatusLog{
		OrderID:   orderID,
		Status:    model.StatusReceived,
		ChangedBy: "system",
		ChangedAt: time.Now(),
		Notes:     nil,
	}
	logID, err := s.repo.CreateLog(ctx, tx, logEntry)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "db_insert_failed", "failed to insert order status log", rid,
			map[string]interface{}{"order_number": order.Number, "error": err.Error()}, err)
		return nil, err
	}
	logEntry.ID = logID

	if err := tx.Commit(ctx); err != nil {
		logger.Log(logger.ERROR, "order-service", "db_transaction_failed", "failed to commit transaction", rid,
			map[string]interface{}{"order_number": order.Number, "error": err.Error()}, err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish to RabbitMQ
	if err := s.rmq.PublishCreatedOrder(ctx, order); err != nil {
		logger.Log(logger.ERROR, "order-service", "rabbitmq_publish_failed", "failed to publish order", rid,
			map[string]interface{}{"order_number": order.Number, "priority": order.Priority}, err)
		return order, fmt.Errorf("order saved but failed to publish message: %w", err)
	}

	logger.Log(logger.DEBUG, "order-service", "order_published", "order published to RabbitMQ", rid,
		map[string]interface{}{"order_number": order.Number, "priority": order.Priority}, nil)

	return order, nil
}
