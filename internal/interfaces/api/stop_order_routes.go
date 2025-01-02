package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/domain/services"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type StopOrderRoutes struct {
	service services.StopOrderService
}

func NewStopOrderRoutes(service services.StopOrderService) *StopOrderRoutes {
	return &StopOrderRoutes{service: service}
}

func (routes *StopOrderRoutes) Register(router chi.Router) {
	router.Route("/api/stop-orders", func(r chi.Router) {
		r.Post("/", routes.CreateStopOrder)
		r.Get("/", routes.ListStopOrders)
		r.Get("/{id}", routes.GetStopOrder)
		r.Patch("/{id}", routes.UpdateStopOrder)
		r.Delete("/{id}", routes.CancelStopOrder)
	})
}

type CreateStopOrderRequest struct {
	Ticker       string `json:"ticker"`
	Side         string `json:"side"`
	TriggerPrice int    `json:"trigger_price"`
	LimitPrice   *int   `json:"limit_price"`
}

type UpdateStopOrderRequest struct {
	TriggerPrice *int `json:"trigger_price"`
	LimitPrice   *int `json:"limit_price"`
}

type StopOrderResponse struct {
	ID           string    `json:"id"`
	Ticker       string    `json:"ticker"`
	Side         string    `json:"side"`
	TriggerPrice int       `json:"trigger_price"`
	LimitPrice   *int      `json:"limit_price"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// In api/mappers.go
func ToStopOrderResponse(order *entities.StopOrder) StopOrderResponse {
	return StopOrderResponse{
		ID:           order.ID().String(),
		Ticker:       order.Ticker(),
		Side:         order.Side().String(),
		TriggerPrice: order.TriggerPrice().Value(),
		Status:       string(order.Status()),
		CreatedAt:    order.CreatedAt(),
		UpdatedAt:    order.UpdatedAt(),
	}
}

func (r *StopOrderRoutes) CreateStopOrder(w http.ResponseWriter, req *http.Request) {
	var request CreateStopOrderRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	side, err := entities.NewSide(request.Side)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	triggerPrice, err := entities.NewContractPrice(request.TriggerPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var limitPrice *entities.ContractPrice
	if request.LimitPrice != nil {
		*limitPrice, err = entities.NewContractPrice(*request.LimitPrice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	order, err := r.service.CreateOrder(request.Ticker, side, triggerPrice, limitPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ToStopOrderResponse(order)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *StopOrderRoutes) ListStopOrders(w http.ResponseWriter, req *http.Request) {
	orders, err := r.service.GetActiveOrders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := lo.Map(orders, func(order *entities.StopOrder, _ int) StopOrderResponse {
		return ToStopOrderResponse(order)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *StopOrderRoutes) GetStopOrder(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	orderID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := r.service.GetOrder(entities.OrderID(orderID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	response := ToStopOrderResponse(order)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *StopOrderRoutes) UpdateStopOrder(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	orderID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var request UpdateStopOrderRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var triggerPrice *entities.ContractPrice
	if request.TriggerPrice != nil {
		tp, err := entities.NewContractPrice(*request.TriggerPrice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return // Don't forget to return after writing error
		}
		triggerPrice = &tp
	}

	var limitPrice *entities.ContractPrice
	if request.LimitPrice != nil {
		*limitPrice, err = entities.NewContractPrice(*request.LimitPrice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	order, err := r.service.UpdateOrder(entities.OrderID(orderID), triggerPrice, limitPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ToStopOrderResponse(order)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *StopOrderRoutes) CancelStopOrder(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	orderID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := r.service.CancelOrder(entities.OrderID(orderID))
	if err != nil {
		var notFoundErr *entities.ErrNotFound // Note the pointer type
		if errors.As(err, &notFoundErr) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ToStopOrderResponse(order)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
