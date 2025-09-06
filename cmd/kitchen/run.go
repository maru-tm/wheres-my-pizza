package kitchen

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"restaurant-system/internal/kitchen/model"
	workerRepository "restaurant-system/internal/kitchen/repository"
	"restaurant-system/internal/kitchen/service"
	"restaurant-system/internal/logger"
	orderRepository "restaurant-system/internal/order/repository"
	"restaurant-system/internal/rabbitmq"
)

func Run(ctx context.Context, pg *pgxpool.Pool, rmq *rabbitmq.RabbitMQ, workerName, types string, prefetch, heartbeat int, requestID string) {
	orderRepo := orderRepository.NewOrderRepository(pg)
	workerRepo := workerRepository.NewWorkerRepository(pg)
	workerService := service.NewKitchenService(workerRepo, orderRepo)

	if err := workerService.RegisterWorker(ctx, workerName, "kitchen-worker"); err != nil {
		if errors.Is(err, model.ErrWorkerAlreadyOnline) {
			logger.Log(logger.ERROR, "kitchen-worker", "worker_registration_failed",
				"worker already online with this name", requestID,
				map[string]interface{}{"worker_name": workerName}, err)
			os.Exit(1)
		}
		logger.Log(logger.ERROR, "kitchen-worker", "worker_registration_failed",
			"failed to register worker", requestID, nil, err)
		os.Exit(1)
	}

}
