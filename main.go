package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"restaurant-system/cmd/kitchen"
	"restaurant-system/cmd/notification"
	"restaurant-system/cmd/order"
	"restaurant-system/cmd/tracking"
	"restaurant-system/config"
	"restaurant-system/pkg/logger"
	"restaurant-system/pkg/postgres"
	"restaurant-system/pkg/rabbitmq"

	"gopkg.in/yaml.v3"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mode := flag.String("mode", "", "Service mode (order-service, kitchen-worker, tracking-service, notification-subscriber)")
	orderPort := flag.Int("port", 3000, "HTTP port for order service")
	maxConcurrent := flag.Int("max-concurrent", 10, "Maximum number of concurrent requests")
	workerName := flag.String("worker-name", "", "Unique kitchen worker name")
	orderTypes := flag.String("order-types", "", "Comma-separated list of order types for this worker")
	prefetch := flag.Int("prefetch", 1, "RabbitMQ prefetch count")
	heartbeat := flag.Int("heartbeat-interval", 30, "Heartbeat interval in seconds")
	trackingPort := flag.Int("tracking-port", 3002, "HTTP port for tracking service")
	configPath := flag.String("config", "config/config.yaml", "Path to config file")

	flag.Parse()

	if *mode == "" {
		fmt.Println("Error: must provide --mode")
		os.Exit(1)
	}

	requestID := fmt.Sprintf("startup-%d-%d", os.Getpid(), time.Now().UnixNano())

	cfg, err := loadConfig(*configPath)
	if err != nil {
		logger.Log(logger.ERROR, *mode, "config_load_failed", "failed to load config", requestID, nil, err)
		os.Exit(1)
	}

	pg, err := postgres.New(ctx, cfg.Database)
	if err != nil {
		logger.Log(logger.ERROR, *mode, "db_connection_failed", "failed to connect to PostgreSQL", requestID, nil, err)
		os.Exit(1)
	}
	defer pg.Close()

	if err := pg.RunMigrations(ctx, "./migrations"); err != nil {
		logger.Log(logger.ERROR, *mode, "migration_failed", "failed to run migrations", requestID, nil, err)
		os.Exit(1)
	}

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
	case "kitchen-worker":
		if *workerName == "" {
			fmt.Println("Error: --worker-name is required for kitchen-worker")
			os.Exit(1)
		}
		var typesList []string
		if *orderTypes != "" {
			typesList = strings.Split(*orderTypes, ",")
			for i := range typesList {
				typesList[i] = strings.TrimSpace(typesList[i])
			}
		}
		kitchen.Run(ctx, pg.Pool, rmq, *workerName, typesList, *prefetch, *heartbeat, requestID)
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
	cancel()
	time.Sleep(1 * time.Second)
}

func loadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}
