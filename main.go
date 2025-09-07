package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"restaurant-system/cmd/kitchen"
	"restaurant-system/cmd/notification"
	"restaurant-system/cmd/order"
	"restaurant-system/cmd/tracking"
	"restaurant-system/internal/config"
	"restaurant-system/internal/db"
	"restaurant-system/internal/logger"
	"restaurant-system/internal/rabbitmq"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mode := flag.String("mode", "", "Service mode (order-service, kitchen-rmq, tracking-service, notification-subscriber)")

	orderPort := flag.Int("port", 3000, "HTTP port for order service")
	maxConcurrent := flag.Int("max-concurrent", 10, "Maximum number of concurrent requests")

	workerName := flag.String("rmq-name", "", "Unique kitchen rmq name")
	orderTypes := flag.String("order-types", "", "Comma-separated list of order types for this rmq")
	prefetch := flag.Int("prefetch", 1, "RabbitMQ prefetch count")
	heartbeat := flag.Int("heartbeat-interval", 30, "Heartbeat interval in seconds")

	trackingPort := flag.Int("tracking-port", 3002, "HTTP port for tracking service")

	flag.Parse()

	if *mode == "" {
		fmt.Println("Must provide --mode")
		os.Exit(1)
	}

	requestID := "startup-" + fmt.Sprint(os.Getpid())

	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Log(logger.ERROR, *mode, "config_load_failed", "failed to load config", requestID, nil, err)
		os.Exit(1)
	}

	pg, err := db.New(ctx, cfg.Database)
	if err != nil {
		logger.Log(logger.ERROR, *mode, "db_connection_failed", "failed to connect to PostgreSQL", requestID, nil, err)
		os.Exit(1)
	}
	defer pg.Close()

	logger.Log(logger.INFO, *mode, "db_connected", "Connected to PostgreSQL database", requestID, nil, nil)

	rmq, err := rabbitmq.New(cfg.RabbitMQ)
	if err != nil {
		logger.Log(logger.ERROR, *mode, "rabbitmq_connection_failed", "failed to connect to RabbitMQ", requestID, nil, err)
		os.Exit(1)
	}
	defer rmq.Close()
	logger.Log(logger.INFO, *mode, "rabbitmq_connected", "Connected to RabbitMQ", requestID, nil, nil)

	switch *mode {
	case "order-service":
		order.Run(ctx, pg.Pool, rmq, *orderPort, *maxConcurrent, requestID)
	case "kitchen-rmq":
		if *workerName == "" {
			fmt.Println("Error: --rmq-name is required for kitchen-rmq")
			os.Exit(1)
		}
		kitchen.Run(ctx, pg.Pool, rmq, *workerName, *orderTypes, *prefetch, *heartbeat, requestID)
	case "tracking-service":
		tracking.Run(ctx, pg.Pool, rmq, *trackingPort, requestID)
	case "notification-subscriber":
		notification.Run(ctx, rmq, requestID)
	default:
		fmt.Printf("Error: unknown mode '%s'\n", *mode)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	logger.Log(logger.INFO, *mode, "shutdown_initiated", "received termination signal, shutting down...", requestID, nil, nil)
}
