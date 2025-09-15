package kitchen

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"restaurant-system/internal/kitchen/infrastructure/pg"
	"restaurant-system/internal/kitchen/infrastructure/rmq"
	"restaurant-system/internal/kitchen/service"
	"restaurant-system/pkg/logger"
	"restaurant-system/pkg/rabbitmq"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(ctx context.Context, pgxPool *pgxpool.Pool, rabbitmq *rabbitmq.RabbitMQ, workerName string, orderTypes []string, prefetch int, heartbeatInterval int, rid string) {
	workerRepo := pg.NewWorkerRepository(pgxPool)
	orderRepo := pg.NewOrderRepository(pgxPool)
	statusPublisher, err := rmq.NewStatusPublisher(rabbitmq)
	if err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "status_publisher_init_failed", "failed to initialize status publisher", rid, nil, err)
		return
	}

	kitchenService := service.NewKitchenService(workerRepo, orderRepo, statusPublisher)

	worker, err := kitchenService.RegisterWorker(ctx, workerName, orderTypes)
	if err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "worker_registration_failed", "failed to register worker", rid,
			map[string]interface{}{"worker_name": workerName}, err)
		os.Exit(1)
	}

	logger.Log(logger.INFO, "kitchen-worker", "worker_registered", "worker registered successfully", rid,
		map[string]interface{}{
			"worker_name":  worker.Name,
			"worker_type":  worker.Type,
			"order_types":  orderTypes,
			"prefetch":     prefetch,
			"heartbeat_ms": heartbeatInterval * 1000,
		}, nil)

	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(time.Duration(heartbeatInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := kitchenService.SendHeartbeat(ctx, worker.ID); err != nil {
					logger.Log(logger.ERROR, "kitchen-worker", "heartbeat_failed", "failed to send heartbeat", rid,
						map[string]interface{}{"worker_id": worker.ID}, err)
				}
			case <-heartbeatCtx.Done():
				return
			}
		}
	}()

	consumer, err := rmq.NewOrderConsumer(rabbitmq, prefetch, orderTypes)
	if err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "consumer_init_failed", "failed to initialize consumer", rid, nil, err)
		stopHeartbeat()
		return
	}

	msgs, err := consumer.ConsumeOrders(ctx)
	if err != nil {
		logger.Log(logger.ERROR, "kitchen-worker", "consume_failed", "failed to start consuming messages", rid, nil, err)
		stopHeartbeat()
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Log(logger.INFO, "kitchen-worker", "shutdown_initiated", "received termination signal", rid, nil, nil)
		stopHeartbeat()

		if err := kitchenService.MarkWorkerOffline(ctx, worker.ID); err != nil {
			logger.Log(logger.ERROR, "kitchen-worker", "shutdown_failed", "failed to mark worker offline", rid,
				map[string]interface{}{"worker_id": worker.ID}, err)
		}

		os.Exit(0)
	}()

	for msg := range msgs {
		processCtx := context.WithValue(ctx, "request_id", fmt.Sprintf("msg-%d", time.Now().UnixNano()))

		if err := kitchenService.ProcessOrder(processCtx, worker, msg); err != nil {
			logger.Log(logger.ERROR, "kitchen-worker", "order_processing_failed", "failed to process order", rid,
				map[string]interface{}{"order_number": msg.OrderNumber}, err)

			if err := consumer.NackMessage(msg.DeliveryTag, true); err != nil {
				logger.Log(logger.ERROR, "kitchen-worker", "nack_failed", "failed to nack message", rid, nil, err)
			}
		} else {
			if err := consumer.AckMessage(msg.DeliveryTag); err != nil {
				logger.Log(logger.ERROR, "kitchen-worker", "ack_failed", "failed to ack message", rid, nil, err)
			}
		}
	}
}
