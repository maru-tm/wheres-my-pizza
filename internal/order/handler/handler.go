package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"restaurant-system/internal/order/model"
	"restaurant-system/internal/order/service"
	"restaurant-system/pkg/logger"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	rid := fmt.Sprintf("req-%d", time.Now().UnixNano())
	ctx := context.WithValue(r.Context(), "request_id", rid)

	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log(logger.ERROR, "order-service", "request_parse_failed", "failed to parse request body", rid, nil, err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	order := &model.Order{
		CustomerName:    req.CustomerName,
		Type:            req.OrderType,
		TableNumber:     req.TableNumber,
		DeliveryAddress: req.DeliveryAddress,
	}

	for _, item := range req.Items {
		order.Items = append(order.Items, model.OrderItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	result, err := h.service.CreateOrder(ctx, order)
	if err != nil {
		logger.Log(logger.ERROR, "order-service", "order_creation_failed", "failed to create order", rid,
			map[string]interface{}{"customer_name": req.CustomerName}, err)

		if err == model.ValidationError {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	logger.Log(logger.DEBUG, "order-service", "order_received", "new order received", rid,
		map[string]interface{}{"order_number": result.Number, "customer_name": result.CustomerName}, nil)

	response := map[string]interface{}{
		"order_number": result.Number,
		"status":       result.Status,
		"total_amount": result.TotalAmount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
