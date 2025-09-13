package tracking

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"restaurant-system/internal/tracking/handler"
	"restaurant-system/internal/tracking/service"
	"restaurant-system/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(ctx context.Context, pgxPool *pgxpool.Pool, rabbitmq interface{}, port int, rid string) {
	trackingService := service.NewTrackingService(pgxPool)
	trackingHandler := handler.NewTrackingHandler(trackingService)

	mux := http.NewServeMux()
	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/orders/"):]

		if r.Method == http.MethodGet {
			if path == "" {
				http.Error(w, "Order number required", http.StatusBadRequest)
				return
			}

			if len(path) > len("/status") && path[len(path)-len("/status"):] == "/status" {
				orderNumber := path[:len(path)-len("/status")]
				trackingHandler.GetOrderStatus(w, r, orderNumber)
			} else if len(path) > len("/history") && path[len(path)-len("/history"):] == "/history" {
				orderNumber := path[:len(path)-len("/history")]
				trackingHandler.GetOrderHistory(w, r, orderNumber)
			} else {
				http.Error(w, "Not found", http.StatusNotFound)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/workers/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			trackingHandler.GetWorkersStatus(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		logger.Log(logger.INFO, "tracking-service", "shutdown_initiated", "received termination signal, shutting down...", rid, nil, nil)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	logger.Log(logger.INFO, "tracking-service", "service_started",
		fmt.Sprintf("Tracking Service started on port %d", port),
		rid,
		map[string]interface{}{"port": port},
		nil,
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log(logger.ERROR, "tracking-service", "server_error", "server failed", rid, nil, err)
	}

	logger.Log(logger.INFO, "tracking-service", "service_stopped", "Tracking Service stopped gracefully", rid, nil, nil)
}
