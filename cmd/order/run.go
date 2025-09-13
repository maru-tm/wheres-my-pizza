package order

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"restaurant-system/internal/order/handler"
	"restaurant-system/internal/order/infrastructure/pg"
	"restaurant-system/internal/order/infrastructure/rmq"
	"restaurant-system/internal/order/service"
	"restaurant-system/pkg/logger"
	"restaurant-system/pkg/rabbitmq"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(ctx context.Context, dbPool *pgxpool.Pool, rmqClient *rabbitmq.RabbitMQ, port int, maxConcurrent int, requestID string) {
	// Initialize repositories
	orderRepo := pg.NewOrderRepository(dbPool)

	// Initialize RabbitMQ publisher
	orderPublisher, err := rmq.NewOrderPublisher(rmqClient)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "order_publisher_init_failed", "failed to initialize order publisher", requestID, nil, err)
		return
	}

	// Initialize service
	orderService := service.NewOrderService(orderRepo, orderPublisher)

	// Initialize handler
	orderHandler := handler.NewOrderHandler(orderService)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "request_id", fmt.Sprintf("req-%d", time.Now().UnixNano()))
		orderHandler.CreateOrderHandler(w, r.WithContext(ctx))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	logger.Log(logger.INFO, "order-service", "service_started", "Order Service started", requestID,
		map[string]interface{}{"port": port, "max_concurrent": maxConcurrent}, nil)

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log(logger.ERROR, "order-service", "http_server_failed", "HTTP server failed", requestID,
				map[string]interface{}{"error": err.Error()}, err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log(logger.ERROR, "order-service", "shutdown_failed", "failed to shutdown server", requestID,
			map[string]interface{}{"error": err.Error()}, err)
	}

	logger.Log(logger.INFO, "order-service", "service_stopped", "Order Service stopped", requestID, nil, nil)
}
