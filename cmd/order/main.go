package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"restaurant-system/internal/config"
	"restaurant-system/internal/db"
	"restaurant-system/internal/logger"
	"restaurant-system/internal/order/handler"
	"restaurant-system/internal/order/repository"
	"restaurant-system/internal/order/service"
	"restaurant-system/internal/rabbitmq"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mode := flag.String("mode", "", "Service mode (must be 'order-service')")
	port := flag.Int("port", 3000, "HTTP port for the API")
	maxConcurrent := flag.Int("max-concurrent", 50, "Maximum number of concurrent orders to process")
	flag.Parse()

	requestID := "startup-" + fmt.Sprint(os.Getpid())

	if *mode != "order-service" {
		logger.Log(logger.ERROR, "order-service", "startup_failed", "must run with --mode=order-service", requestID, nil, nil)
		return
	}

	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "config_load_failed", "failed to load config", requestID, nil, err)
		return
	}

	postgres, err := db.New(ctx, cfg.Database)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "db_connection_failed", "failed to connect to PostgreSQL", requestID, nil, err)
		return
	}
	defer postgres.Close()

	if err := postgres.RunMigrations(ctx, "migrations"); err != nil {
		logger.Log(logger.ERROR, "order-service", "db_migrations_failed", "failed to run migrations", requestID, nil, err)
		return
	}
	logger.Log(logger.INFO, "order-service", "db_connected", "Connected to PostgreSQL database", requestID, nil, nil)

	rmq, err := rabbitmq.New(cfg.RabbitMQ)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "rabbitmq_connection_failed", "failed to connect to RabbitMQ", requestID, nil, err)
		return
	}
	logger.Log(logger.INFO, "order-service", "rabbitmq_connected", "Connected to RabbitMQ exchange 'orders_topic'", requestID, nil, nil)

	orderRepo := repository.NewOrderRepository(postgres.Pool)
	orderRMQ, err := repository.NewOrderPublisher(rmq)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "order_publisher_init_failed", "failed to initialize order publisher", requestID, nil, err)
		return
	}

	orderService := service.NewOrderService(orderRepo, orderRMQ)
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
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Log(logger.INFO, "order-service", "shutdown_initiated", "received termination signal, shutting down...", requestID, nil, nil)
		server.Shutdown(ctx)
	}()

	logger.Log(logger.INFO, "order-service", "service_started",
		fmt.Sprintf("Order Service started on port %d", *port),
		requestID,
		map[string]interface{}{
			"port":           *port,
			"max_concurrent": *maxConcurrent,
		},
		nil,
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log(logger.ERROR, "order-service", "server_error", "server failed", requestID, nil, err)
	}

	logger.Log(logger.INFO, "order-service", "service_stopped", "Order Service stopped gracefully", requestID, nil, nil)
}
