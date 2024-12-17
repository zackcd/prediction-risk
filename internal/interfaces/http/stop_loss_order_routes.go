package http

import (
	"encoding/json"
	"net/http"
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/domain/services"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type StopLossRoutes struct {
	service services.StopLossService
}

func NewStopLossRoutes(service services.StopLossService) *StopLossRoutes {
	return &StopLossRoutes{service: service}
}

func (routes *StopLossRoutes) Register(router chi.Router) {
	router.Route("/api/stop-loss", func(r chi.Router) {
		r.Post("/", routes.CreateStopLoss)
		r.Get("/", routes.ListStopLoss)
		r.Get("/{id}", routes.GetStopLoss)
		r.Patch("/{id}", routes.UpdateStopLoss)
		r.Delete("/{id}", routes.CancelStopLoss)
	})
}

type CreateStopLossRequest struct {
	Ticker    string `json:"ticker"`
	Side      string `json:"side"`
	Threshold int    `json:"threshold"`
}

type UpdateStopLossRequest struct {
	Threshold int `json:"threshold"`
}

func (r *StopLossRoutes) CreateStopLoss(w http.ResponseWriter, req *http.Request) {
	var request CreateStopLossRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	side, err := entities.NewSide(request.Side)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	threshold, err := entities.NewContractPrice(request.Threshold)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	order, err := r.service.CreateOrder(request.Ticker, side, threshold)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (r *StopLossRoutes) ListStopLoss(w http.ResponseWriter, req *http.Request) {
	orders, err := r.service.GetActiveOrders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (r *StopLossRoutes) GetStopLoss(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	orderID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := r.service.GetOrder(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (r *StopLossRoutes) UpdateStopLoss(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	orderID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var request UpdateStopLossRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	threshold, err := entities.NewContractPrice(request.Threshold)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	order, err := r.service.UpdateOrder(orderID, threshold)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (r *StopLossRoutes) CancelStopLoss(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	orderID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := r.service.CancelOrder(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
