package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"prediction-risk/internal/app/contract"
	"prediction-risk/internal/app/core"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
	trigger_service "prediction-risk/internal/app/risk/trigger/service"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type StopTriggerRoutes struct {
	service *trigger_service.TriggerService
}

func NewStopTriggerRoutes(service *trigger_service.TriggerService) *StopTriggerRoutes {
	return &StopTriggerRoutes{service: service}
}

func (routes *StopTriggerRoutes) Register(router chi.Router) {
	router.Route("/api/stop-triggers", func(r chi.Router) {
		r.Post("/", routes.CreateStopTrigger)
		r.Get("/", routes.ListStopTriggers)
		r.Get("/{id}", routes.GetStopTrigger)
		r.Patch("/{id}", routes.UpdateStopTrigger)
		r.Delete("/{id}", routes.CancelStopTrigger)
	})
}

type CreateStopTriggerRequest struct {
	Contract struct {
		Ticker string `json:"ticker"`
		Side   string `json:"side"`
	} `json:"contract"`
	TriggerPrice int  `json:"trigger_price"`
	LimitPrice   *int `json:"limit_price"`
}

type UpdateStopTriggerRequest struct {
	TriggerPrice *int `json:"trigger_price"`
	LimitPrice   *int `json:"limit_price"`
}

type ContractIDResponse struct {
	Ticker string `json:"ticker"`
	Side   string `json:"side"`
}

type StopTriggerResponse struct {
	TriggerID    string             `json:"trigger_id"`
	TriggerType  string             `json:"trigger_type"`
	Contract     ContractIDResponse `json:"contract"`
	Status       string             `json:"status"`
	TriggerPrice int                `json:"trigger_price"`
	LimitPrice   *int               `json:"limit_price"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// In api/mappers.go
func ToStopTriggerResponse(trigger *trigger_domain.Trigger) StopTriggerResponse {
	var limitPrice *int
	if trigger.Actions[0].LimitPrice != nil {
		value := trigger.Actions[0].LimitPrice.Value()
		limitPrice = &value
	}

	return StopTriggerResponse{
		TriggerID:   trigger.TriggerID.String(),
		TriggerType: trigger.TriggerType.String(),
		Contract: ContractIDResponse{
			Ticker: string(trigger.Condition.Contract.Ticker),
			Side:   trigger.Condition.Contract.Side.String(),
		},
		Status:       trigger.Status.String(),
		TriggerPrice: trigger.Condition.Price.Threshold.Value(),
		LimitPrice:   limitPrice,
		CreatedAt:    trigger.CreatedAt,
		UpdatedAt:    trigger.UpdatedAt,
	}
}

func (r *StopTriggerRoutes) CreateStopTrigger(w http.ResponseWriter, req *http.Request) {
	var request CreateStopTriggerRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	side, err := contract.NewSide(request.Contract.Side)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	contractIdentifier := contract.ContractIdentifier{
		Ticker: contract.Ticker(request.Contract.Ticker),
		Side:   side,
	}

	triggerPrice, err := contract.NewContractPrice(request.TriggerPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var limitPrice *contract.ContractPrice
	if request.LimitPrice != nil {
		cp, err := contract.NewContractPrice(*request.LimitPrice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		limitPrice = &cp
	}

	trigger, err := r.service.CreateStopTrigger(contractIdentifier, triggerPrice, limitPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ToStopTriggerResponse(trigger)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *StopTriggerRoutes) ListStopTriggers(w http.ResponseWriter, req *http.Request) {
	triggers, err := r.service.Get()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := lo.Map(triggers, func(trigger *trigger_domain.Trigger, _ int) StopTriggerResponse {
		return ToStopTriggerResponse(trigger)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *StopTriggerRoutes) GetStopTrigger(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	triggerID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trigger, err := r.service.GetByID(trigger_domain.TriggerID(triggerID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if trigger == nil {
		http.Error(w, "trigger not found", http.StatusNotFound)
		return
	}

	response := ToStopTriggerResponse(trigger)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *StopTriggerRoutes) UpdateStopTrigger(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	triggerID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var request UpdateStopTriggerRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var triggerPrice *contract.ContractPrice
	if request.TriggerPrice != nil {
		tp, err := contract.NewContractPrice(*request.TriggerPrice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return // Don't forget to return after writing error
		}
		triggerPrice = &tp
	}

	var limitPrice *contract.ContractPrice
	if request.LimitPrice != nil {
		cp, err := contract.NewContractPrice(*request.LimitPrice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		limitPrice = &cp
	}

	trigger, err := r.service.UpdateStopTrigger(trigger_domain.TriggerID(triggerID), triggerPrice, limitPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ToStopTriggerResponse(trigger)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (r *StopTriggerRoutes) CancelStopTrigger(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	triggerID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trigger, err := r.service.CancelTrigger(trigger_domain.TriggerID(triggerID))
	if err != nil {
		var notFoundErr *core.ErrNotFound // Note the pointer type
		if errors.As(err, &notFoundErr) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ToStopTriggerResponse(trigger)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
