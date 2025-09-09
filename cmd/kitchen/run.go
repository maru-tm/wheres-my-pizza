package kitchen

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	workerRepository "restaurant-system/internal/kitchen/infrastructure/pg"
	"restaurant-system/internal/kitchen/infrastructure/rmq"
	"restaurant-system/internal/kitchen/model"
	"restaurant-system/internal/kitchen/service"
	"restaurant-system/internal/logger"
	orderRepository "restaurant-system/internal/order/infrastructure/pg"
	"restaurant-system/internal/rabbitmq"
)

func Run(ctx context.Context, pgxPool *pgxpool.Pool, rabbitmq *rabbitmq.RabbitMQ, workerName, types string, prefetch, heartbeat int, requestID string) {
	orderRepo := orderRepository.NewOrderRepository(pgxPool)
	workerRepo := workerRepository.NewWorkerRepository(pgxPool)
	workerService := service.NewKitchenService(workerRepo, orderRepo)

	if err := workerService.RegisterWorker(ctx, workerName, "kitchen-rmq"); err != nil {
		if errors.Is(err, model.ErrWorkerAlreadyOnline) {
			logger.Log(logger.ERROR, "kitchen-rmq", "worker_registration_failed",
				"rmq already online with this name", requestID,
				map[string]interface{}{"worker_name": workerName}, err)
			os.Exit(1)
		}
		logger.Log(logger.ERROR, "kitchen-rmq", "worker_registration_failed",
			"failed to register rmq", requestID, nil, err)
		os.Exit(1)
	}
	consumer, err := rmq.NewOrderConsumer(rabbitmq, workerName)
	if err != nil {
		logger.Log(logger.ERROR, "kitchen-service", "order_consumer_init_failed", "failed to initialize order consumer", requestID, nil, err)
		os.Exit(1)
	}
	statusPublisher, err := rmq.NewStatusPublisher(rabbitmq)
	if err != nil {
		logger.Log(logger.ERROR, "kitchen-service", "order_consumer_init_failed", "failed to initialize status publisher", requestID, nil, err)
		os.Exit(1)
	}
	kitchenWorker := service.NewKitchenWorker(workerService, consumer, statusPublisher, workerName)

}
