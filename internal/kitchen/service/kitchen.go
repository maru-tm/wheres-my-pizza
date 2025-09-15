package service

import (
	"context"
	"fmt"
	"time"

	"restaurant-system/internal/kitchen/infrastructure/rmq"
	"restaurant-system/internal/kitchen/model"
	"restaurant-system/pkg/logger"
)

type WorkerRepository interface {
	CreateOrUpdateWorker(ctx context.Context, name string, workerType string, orderTypes []string) (*model.Worker, error)
	UpdateWorkerHeartbeat(ctx context.Context, id int) error
	MarkWorkerOffline(ctx context.Context, id int) error
	IncrementOrdersProcessed(ctx context.Context, id int) error
}

type OrderRepository interface {
	UpdateOrderStatus(ctx context.Context, orderNumber string, status string, processedBy string) error
	CreateStatusLog(ctx context.Context, orderNumber string, status string, changedBy string, notes *string) error
	GetOrderByNumber(ctx context.Context, orderNumber string) (*model.Order, error)
}

type StatusPublisher interface {
	PublishStatusUpdate(ctx context.Context, update *rmq.StatusUpdateMessage) error
}

type KitchenService struct {
	workerRepo WorkerRepository
	orderRepo  OrderRepository
	publisher  StatusPublisher
}

func NewKitchenService(wr WorkerRepository, or OrderRepository, sp StatusPublisher) *KitchenService {
	return &KitchenService{
		workerRepo: wr,
		orderRepo:  or,
		publisher:  sp,
	}
}

func (s *KitchenService) RegisterWorker(ctx context.Context, name string, orderTypes []string) (*model.Worker, error) {
	workerType := "general"
	if len(orderTypes) > 0 {
		workerType = "specialized"
	}

	worker, err := s.workerRepo.CreateOrUpdateWorker(ctx, name, workerType, orderTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to register worker: %w", err)
	}

	return worker, nil
}

func (s *KitchenService) SendHeartbeat(ctx context.Context, workerID int) error {
	return s.workerRepo.UpdateWorkerHeartbeat(ctx, workerID)
}

func (s *KitchenService) MarkWorkerOffline(ctx context.Context, workerID int) error {
	return s.workerRepo.MarkWorkerOffline(ctx, workerID)
}

func (s *KitchenService) ProcessOrder(ctx context.Context, worker *model.Worker, orderMsg *rmq.OrderMessage) error {
	rid := ""
	if v := ctx.Value("request_id"); v != nil {
		if str, ok := v.(string); ok {
			rid = str
		}
	}

	logger.Log(logger.DEBUG, "kitchen-worker", "order_processing_started", "started processing order", rid,
		map[string]interface{}{
			"order_number": orderMsg.OrderNumber,
			"worker_name":  worker.Name,
		}, nil)

	if len(worker.OrderTypes) > 0 {
		canHandle := false
		for _, t := range worker.OrderTypes {
			if t == orderMsg.OrderType {
				canHandle = true
				break
			}
		}
		if !canHandle {
			logger.Log(logger.DEBUG, "kitchen-worker", "order_rejected", "worker cannot handle this order type", rid,
				map[string]interface{}{
					"order_number": orderMsg.OrderNumber,
					"order_type":   orderMsg.OrderType,
					"worker_types": worker.OrderTypes,
				}, nil)
			return fmt.Errorf("worker cannot handle order type %s", orderMsg.OrderType)
		}
	}

	if err := s.orderRepo.UpdateOrderStatus(ctx, orderMsg.OrderNumber, "cooking", worker.Name); err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "status_update_failed", "failed to update order status to cooking", rid,
			map[string]interface{}{
				"order_number": orderMsg.OrderNumber,
				"error":        err.Error(),
			}, err)
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if err := s.orderRepo.CreateStatusLog(ctx, orderMsg.OrderNumber, "cooking", worker.Name, nil); err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "status_log_failed", "failed to create status log", rid,
			map[string]interface{}{
				"order_number": orderMsg.OrderNumber,
				"error":        err.Error(),
			}, err)
		return fmt.Errorf("failed to create status log: %w", err)
	}

	update := &rmq.StatusUpdateMessage{
		OrderNumber:         orderMsg.OrderNumber,
		OldStatus:           "received",
		NewStatus:           "cooking",
		ChangedBy:           worker.Name,
		Timestamp:           time.Now(),
		EstimatedCompletion: time.Now().Add(10 * time.Minute),
	}
	if err := s.publisher.PublishStatusUpdate(ctx, update); err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "status_publish_failed", "failed to publish status update", rid,
			map[string]interface{}{
				"order_number": orderMsg.OrderNumber,
				"error":        err.Error(),
			}, err)
	}

	var cookingTime time.Duration
	switch orderMsg.OrderType {
	case "dine_in":
		cookingTime = 8 * time.Second
	case "takeout":
		cookingTime = 10 * time.Second
	case "delivery":
		cookingTime = 12 * time.Second
	default:
		cookingTime = 10 * time.Second
	}

	time.Sleep(cookingTime)

	if err := s.orderRepo.UpdateOrderStatus(ctx, orderMsg.OrderNumber, "ready", worker.Name); err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "status_update_failed", "failed to update order status to ready", rid,
			map[string]interface{}{
				"order_number": orderMsg.OrderNumber,
				"error":        err.Error(),
			}, err)
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if err := s.orderRepo.CreateStatusLog(ctx, orderMsg.OrderNumber, "ready", worker.Name, nil); err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "status_log_failed", "failed to create status log", rid,
			map[string]interface{}{
				"order_number": orderMsg.OrderNumber,
				"error":        err.Error(),
			}, err)
		return fmt.Errorf("failed to create status log: %w", err)
	}

	if err := s.workerRepo.IncrementOrdersProcessed(ctx, worker.ID); err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "worker_update_failed", "failed to increment orders processed", rid,
			map[string]interface{}{
				"worker_id":    worker.ID,
				"order_number": orderMsg.OrderNumber,
				"error":        err.Error(),
			}, err)
	}

	update = &rmq.StatusUpdateMessage{
		OrderNumber:         orderMsg.OrderNumber,
		OldStatus:           "cooking",
		NewStatus:           "ready",
		ChangedBy:           worker.Name,
		Timestamp:           time.Now(),
		EstimatedCompletion: time.Now(),
	}
	if err := s.publisher.PublishStatusUpdate(ctx, update); err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "status_publish_failed", "failed to publish status update", rid,
			map[string]interface{}{
				"order_number": orderMsg.OrderNumber,
				"error":        err.Error(),
			}, err)
	}

	logger.Log(logger.DEBUG, "kitchen-worker", "order_completed", "order processing completed", rid,
		map[string]interface{}{
			"order_number": orderMsg.OrderNumber,
			"worker_name":  worker.Name,
			"cooking_time": cookingTime.String(),
		}, nil)

	return nil
}
