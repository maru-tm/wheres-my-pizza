package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"restaurant-system/internal/logger"
	"restaurant-system/internal/order/model"
	"restaurant-system/internal/order/service"
	"time"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest

	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log(logger.ERROR, "order-service", "validation_failed", "invalid request body", requestID,
			map[string]interface{}{"body_error": err.Error()}, err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	order := model.Order{
		CustomerName: req.CustomerName,
		Type:         model.OrderType(req.OrderType),
		Status:       model.StatusReceived,
		Items:        []model.OrderItem{},
	}
	for _, item := range req.Items {
		order.Items = append(order.Items, model.OrderItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	logger.Log(logger.DEBUG, "order-service", "order_received", "new order request received", requestID,
		map[string]interface{}{"customer_name": req.CustomerName, "items_count": len(order.Items)}, nil)

	created, err := h.service.CreateOrder(r.Context(), &order)
	if err != nil {
		if err == model.ValidationError {
			logger.Log(logger.ERROR, "order-service", "validation_failed", "order validation failed", requestID,
				map[string]interface{}{"customer_name": req.CustomerName}, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			logger.Log(logger.ERROR, "order-service", "order_creation_failed", "failed to create order", requestID,
				map[string]interface{}{"customer_name": req.CustomerName}, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := CreateOrderResponse{
		OrderNumber: created.Number,
		Status:      string(created.Status),
		TotalAmount: created.TotalAmount,
	}

	logger.Log(logger.DEBUG, "order-service", "order_created", "order successfully created", requestID,
		map[string]interface{}{"order_number": created.Number, "total_amount": created.TotalAmount}, nil)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
