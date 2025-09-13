package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"restaurant-system/internal/tracking/service"
	"restaurant-system/pkg/logger"
)

type TrackingHandler struct {
	service *service.TrackingService
}

func NewTrackingHandler(s *service.TrackingService) *TrackingHandler {
	return &TrackingHandler{service: s}
}

func (h *TrackingHandler) GetOrderStatus(w http.ResponseWriter, r *http.Request, orderNumber string) {
	rid := fmt.Sprintf("req-%d", time.Now().UnixNano())

	logger.Log(logger.DEBUG, "tracking-service", "request_received", "order status request received", rid,
		map[string]interface{}{"order_number": orderNumber, "endpoint": "status"}, nil)

	status, err := h.service.GetOrderStatus(r.Context(), orderNumber)
	if err != nil {
		logger.Log(logger.ERROR, "tracking-service", "order_status_failed", "failed to get order status", rid,
			map[string]interface{}{"order_number": orderNumber}, err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *TrackingHandler) GetOrderHistory(w http.ResponseWriter, r *http.Request, orderNumber string) {
	rid := fmt.Sprintf("req-%d", time.Now().UnixNano())

	logger.Log(logger.DEBUG, "tracking-service", "request_received", "order history request received", rid,
		map[string]interface{}{"order_number": orderNumber, "endpoint": "history"}, nil)

	history, err := h.service.GetOrderHistory(r.Context(), orderNumber)
	if err != nil {
		logger.Log(logger.ERROR, "tracking-service", "order_history_failed", "failed to get order history", rid,
			map[string]interface{}{"order_number": orderNumber}, err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *TrackingHandler) GetWorkersStatus(w http.ResponseWriter, r *http.Request) {
	rid := fmt.Sprintf("req-%d", time.Now().UnixNano())

	logger.Log(logger.DEBUG, "tracking-service", "request_received", "workers status request received", rid,
		map[string]interface{}{"endpoint": "workers/status"}, nil)

	workers, err := h.service.GetWorkersStatus(r.Context())
	if err != nil {
		logger.Log(logger.ERROR, "tracking-service", "workers_status_failed", "failed to get workers status", rid, nil, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workers)
}
