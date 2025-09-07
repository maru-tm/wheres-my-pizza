package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	workerModel "restaurant-system/internal/kitchen/model"
	"restaurant-system/internal/logger"
	orderModel "restaurant-system/internal/order/model"
)

type WorkerRepo interface {
	GetStatus(ctx context.Context, name string) (string, error)
	Create(ctx context.Context, name, workerType string) error
	SetOnline(ctx context.Context, name string) error
	IncrementProcessedOrders(ctx context.Context, name string) error
}

type OrderRepo interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)

	GetStatus(ctx context.Context, orderID int) (orderModel.OrderStatus, error)
	UpdateStatus(ctx context.Context, tx pgx.Tx, orderID int, status orderModel.OrderStatus, workerName string) error
	InsertStatusLog(ctx context.Context, tx pgx.Tx, log orderModel.OrderStatusLog) error
}
type KitchenService struct {
	wRepo WorkerRepo
	oRepo OrderRepo
}

func NewKitchenService(wRepo WorkerRepo, oRepo OrderRepo) *KitchenService {
	return &KitchenService{wRepo: wRepo, oRepo: oRepo}
}

func (s *KitchenService) RegisterWorker(ctx context.Context, workerName, workerType string) error {
	status, err := s.wRepo.GetStatus(ctx, workerName)
	if err != nil {
		return err
	}

	if status == "" {
		return s.wRepo.Create(ctx, workerName, workerType)
	}

	if status == "online" {
		return workerModel.ErrWorkerAlreadyOnline
	}

	return s.wRepo.SetOnline(ctx, workerName)
}

func (s *KitchenService) StartCooking(ctx context.Context, orderID int, workerName string, notes string, rid string) error {
	tx, err := s.oRepo.BeginTx(ctx)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "db_transaction_failed", "failed to begin transaction", rid,
			map[string]interface{}{"error": err.Error()}, err)
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		}
	}()
	currentStatus, err := s.oRepo.GetStatus(ctx, orderID)
	if currentStatus == "cooking" {
		return workerModel.ErrAlreadyCooking
	}
	err = s.oRepo.UpdateStatus(ctx, tx, orderID, orderModel.StatusCooking, workerName)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	log := orderModel.OrderStatusLog{
		OrderID:   orderID,
		Status:    currentStatus,
		ChangedBy: workerName,
		Notes:     &notes,
	}
	err = s.oRepo.InsertStatusLog(ctx, tx, log)
	if err != nil {
		return fmt.Errorf("failed to insert log: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *KitchenService) CompleteOrder(ctx context.Context, orderID int, workerName string, notes string, rid string) error {
	tx, err := s.oRepo.BeginTx(ctx)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "db_transaction_failed", "failed to begin transaction", rid,
			map[string]interface{}{"error": err.Error()}, err)
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		}
	}()
	currentStatus, err := s.oRepo.GetStatus(ctx, orderID)
	if currentStatus == "cooking" {
		return workerModel.ErrAlreadyCooking
	}
	err = s.oRepo.UpdateStatus(ctx, tx, orderID, orderModel.StatusCompleted, workerName)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	log := orderModel.OrderStatusLog{
		OrderID:   orderID,
		Status:    currentStatus,
		ChangedBy: workerName,
		Notes:     &notes,
	}
	err = s.oRepo.InsertStatusLog(ctx, tx, log)
	if err != nil {
		return fmt.Errorf("failed to insert log: %w", err)
	}
	err = s.wRepo.IncrementProcessedOrders(ctx, workerName)
	if err != nil {
		return fmt.Errorf("failed to increment rmq stats: %w", err)
	}
	return tx.Commit(ctx)
}
