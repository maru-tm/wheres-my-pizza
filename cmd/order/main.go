package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"restaurant-system/internal/config"
	"restaurant-system/internal/db"
	"restaurant-system/internal/order/handler"
	"restaurant-system/internal/order/repository"
	"restaurant-system/internal/order/service"
	"restaurant-system/internal/rabbitmq"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mode := flag.String("mode", "", "Service mode (must be 'order-service')")
	port := flag.Int("port", 3000, "HTTP port for the API")
	maxConcurrent := flag.Int("max-concurrent", 50, "Maximum number of concurrent orders to process")
	flag.Parse()

	if *mode != "order-service" {
		log.Fatal("must run with --mode=order-service")
	}

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	postgres, err := db.New(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("❌ ошибка подключения к PostgreSQL: %v", err)
	}
	defer postgres.Close()

	if err := postgres.RunMigrations(ctx, "migrations"); err != nil {
		log.Fatalf("❌ ошибка миграций: %v", err)
	}

	rmq, err := rabbitmq.New(cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("❌ ошибка подключения к RabbitMQ: %v", err)
	}
	defer rmq.Close()

	orderRepo := repository.NewOrderRepository(postgres.Pool)
	orderService := service.NewOrderService(orderRepo, rmq)
	orderHandler := handler.NewOrderHandler(orderService)

	mux := http.NewServeMux()
	mux.HandleFunc("/orders", orderHandler.CreateOrder)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("🛑 Получен сигнал завершения, останавливаем Order Service...")
		server.Shutdown(ctx)
	}()

	log.Printf("✅ Order Service started on port %d (max_concurrent=%d)\n", *port, *maxConcurrent)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}

	log.Println("Order Service stopped gracefully")
}
