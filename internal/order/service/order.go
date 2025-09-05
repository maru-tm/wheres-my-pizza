package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"restaurant-system/internal/order/model"
	"restaurant-system/internal/rabbitmq"
	"time"
)

type PGInterface interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)

	CreateOrder(ctx context.Context, tx pgx.Tx, order *model.Order) (int, error)
	CreateItem(ctx context.Context, tx pgx.Tx, item *model.OrderItem) (int, error)
	CreateLog(ctx context.Context, tx pgx.Tx, logEntry *model.OrderStatusLog) (int, error)

	GetNextOrderSequence(ctx context.Context, tx pgx.Tx, date string) (int, error)
}
type RabbitMQInterface interface{}
type OrderService struct {
	repo PGInterface
	rmq  *rabbitmq.RabbitMQ
}

func NewOrderService(r PGInterface, rmq *rabbitmq.RabbitMQ) *OrderService {
	return &OrderService{repo: r, rmq: rmq}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	if err := order.Validate(); err != nil {
		return nil, err
	}

	var total float64
	for _, item := range order.Items {
		total += item.Price * float64(item.Quantity)
	}
	order.TotalAmount = total
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

	today := time.Now().UTC().Format("20060102")

	seq, err := s.repo.GetNextOrderSequence(ctx, tx, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get next order sequence: %w", err)
	}

	orderNumber := fmt.Sprintf("ORD_%s_%03d", today, seq)
	order.Number = orderNumber

	orderID, err := s.repo.CreateOrder(ctx, tx, order)
	if err != nil {
		return nil, err
	}
	order.ID = orderID

	for i := range order.Items {
		order.Items[i].OrderID = orderID
		itemID, err := s.repo.CreateItem(ctx, tx, &order.Items[i])
		if err != nil {
			return nil, err
		}
		order.Items[i].ID = itemID
	}

	logEntry := &model.OrderStatusLog{
		OrderID:   orderID,
		Status:    model.StatusReceived,
		ChangedBy: "system",
	}
	logID, err := s.repo.CreateLog(ctx, tx, logEntry)
	if err != nil {
		return nil, err
	}
	logEntry.ID = logID

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}
