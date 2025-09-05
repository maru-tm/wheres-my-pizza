package handler

import (
	"encoding/json"
	"net/http"
	"restaurant-system/internal/order/model"
	"restaurant-system/internal/order/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	created, err := h.service.CreateOrder(r.Context(), &order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreateOrderResponse{
		OrderNumber: created.Number,
		Status:      string(created.Status),
		TotalAmount: created.TotalAmount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
