package order

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"restaurant-system/internal/logger"
	"restaurant-system/internal/order/handler"
	"restaurant-system/internal/order/infrastructure/pg"
	"restaurant-system/internal/order/infrastructure/rmq"
	"restaurant-system/internal/order/service"
	"restaurant-system/internal/rabbitmq"
)

func Run(ctx context.Context, pgxPool *pgxpool.Pool, rabbitmq *rabbitmq.RabbitMQ, port int, maxConcurrent int, rid string) {
	orderRepo := pg.NewOrderRepository(pgxPool)
	orderPublisher, err := rmq.NewOrderPublisher(rabbitmq)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "order_publisher_init_failed", "failed to initialize order publisher", rid, nil, err)
		return
	}

	orderService := service.NewOrderService(orderRepo, orderPublisher)
	orderHandler := handler.NewOrderHandler(orderService)

	mux := http.NewServeMux()
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		orderHandler.CreateOrderHandler(w, r)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		logger.Log(logger.INFO, "order-service", "shutdown_initiated", "received termination signal, shutting down...", rid, nil, nil)
		server.Shutdown(context.Background())
	}()

	logger.Log(logger.INFO, "order-service", "service_started",
		fmt.Sprintf("Order Service started on port %d", port),
		rid,
		map[string]interface{}{
			"port":           port,
			"max_concurrent": maxConcurrent,
		},
		nil,
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log(logger.ERROR, "order-service", "server_error", "server failed", rid, nil, err)
	}

	logger.Log(logger.INFO, "order-service", "service_stopped", "Order Service stopped gracefully", rid, nil, nil)
}
