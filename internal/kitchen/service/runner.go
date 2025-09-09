package service

import (
	"context"
	"fmt"
	"restaurant-system/internal/kitchen/infrastructure/rmq"
	"restaurant-system/internal/logger"
	"restaurant-system/internal/order/model"
	"time"
)

type OrderConsumer interface {
	Consume(ctx context.Context) (*rmq.OrderMessage, uint64, error)
	Ack(deliveryTag uint64) error
	Nack(deliveryTag uint64, requeue bool) error
}

type StatusPublisher interface {
	PublishUpdatedStatus(ctx context.Context, msg rmq.StatusUpdateMessage) error
}
type KitchenWorker struct {
	service         *KitchenService
	orderConsumer   OrderConsumer
	statusPublisher StatusPublisher
	workerName      string
}

func NewKitchenWorker(service *KitchenService, orderConsumer OrderConsumer, statusPublisher StatusPublisher, workerName string) *KitchenWorker {
	return &KitchenWorker{
		service:         service,
		orderConsumer:   orderConsumer,
		statusPublisher: statusPublisher,
		workerName:      workerName,
	}
}

func (w *KitchenWorker) Run(ctx context.Context, requestID string) {
	logger.Log(logger.INFO, "kitchen-worker", "worker_started", "kitchen worker started", requestID,
		map[string]interface{}{"worker_name": w.workerName}, nil)

	for {
		select {
		case <-ctx.Done():
			logger.Log(logger.INFO, "kitchen-worker", "shutdown", "worker shutting down gracefully", requestID,
				map[string]interface{}{"worker_name": w.workerName}, nil)
			return

		default:
			msg, deliveryTag, err := w.orderConsumer.Consume(ctx)
			if err != nil {
				logger.Log(logger.ERROR, "kitchen-worker", "consume_failed", "failed to consume message", requestID,
					map[string]interface{}{"worker_name": w.workerName}, err)
				continue // ждём следующее сообщение
			}

			if err := w.ProcessOrder(ctx, msg, deliveryTag, requestID); err != nil {
				logger.Log(logger.ERROR, "kitchen-worker", "process_failed", "failed to process order", requestID,
					map[string]interface{}{"worker_name": w.workerName, "order_number": msg.OrderNumber}, err)
			}
		}
	}
}

func (w *KitchenWorker) ProcessOrder(ctx context.Context, msg *rmq.OrderMessage, deliveryTag uint64, requestID string) error {
	if msg.OrderType == "" {
		_ = w.orderConsumer.Nack(deliveryTag, false)
		return fmt.Errorf("invalid order_type for order %s", msg.OrderNumber)
	}

	if err := w.service.StartCooking(ctx, msg.OrderNumber, w.workerName, "started cooking", requestID); err != nil {
		_ = w.orderConsumer.Nack(deliveryTag, true)
		return fmt.Errorf("failed to start cooking: %w", err)
	}

	statusMsg := rmq.StatusUpdateMessage{
		OrderNumber:         msg.OrderNumber,
		OldStatus:           string(model.StatusReceived),
		NewStatus:           string(model.StatusCooking),
		ChangedBy:           w.workerName,
		Timestamp:           time.Now().UTC(),
		EstimatedCompletion: time.Now().Add(estimateCookingTime(msg.OrderType)).UTC(),
	}
	if err := w.statusPublisher.PublishUpdatedStatus(ctx, statusMsg); err != nil {
		_ = w.orderConsumer.Nack(deliveryTag, true)
		return fmt.Errorf("failed to publish cooking status: %w", err)
	}

	time.Sleep(estimateCookingTime(msg.OrderType))

	if err := w.service.CompleteOrder(ctx, msg.OrderNumber, w.workerName, "order ready", requestID); err != nil {
		_ = w.orderConsumer.Nack(deliveryTag, true)
		return fmt.Errorf("failed to complete order: %w", err)
	}

	readyMsg := rmq.StatusUpdateMessage{
		OrderNumber: msg.OrderNumber,
		OldStatus:   string(model.StatusCooking),
		NewStatus:   string(model.StatusReady),
		ChangedBy:   w.workerName,
		Timestamp:   time.Now().UTC(),
	}
	if err := w.statusPublisher.PublishUpdatedStatus(ctx, readyMsg); err != nil {
		_ = w.orderConsumer.Nack(deliveryTag, true)
		return fmt.Errorf("failed to publish ready status: %w", err)
	}

	if err := w.orderConsumer.Ack(deliveryTag); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}

	logger.Log(logger.INFO, "kitchen-worker", "order_processed", "order processed successfully", requestID, map[string]interface{}{
		"order_number": msg.OrderNumber,
	}, nil)

	return nil
}

func estimateCookingTime(orderType string) time.Duration {
	switch orderType {
	case "dine_in":
		return 8 * time.Second
	case "takeout":
		return 10 * time.Second
	case "delivery":
		return 12 * time.Second
	default:
		return 5 * time.Second
	}
}
