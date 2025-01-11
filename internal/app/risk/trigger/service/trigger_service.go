package trigger_service

import (
	"context"
	"errors"
	"fmt"
	"prediction-risk/internal/app/contract"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
)

var (
	ErrInvalidTrigger     = errors.New("invalid trigger")
	ErrTriggerNotFound    = errors.New("trigger not found")
	ErrInvalidTriggerType = errors.New("invalid trigger type")
)

type TriggerRepository interface {
	Save(ctx context.Context, trigger *trigger_domain.Trigger) error
	Get(ctx context.Context, id trigger_domain.TriggerID) (*trigger_domain.Trigger, error)
	GetAll(ctx context.Context) ([]*trigger_domain.Trigger, error)
	Update(ctx context.Context, trigger *trigger_domain.Trigger) error
}

type TriggerService struct {
	repository TriggerRepository
}

func NewTriggerService(repository TriggerRepository) *TriggerService {
	return &TriggerService{
		repository: repository,
	}
}

// GetByID retrieves a specific trigger by its ID
func (s *TriggerService) GetByID(triggerID trigger_domain.TriggerID) (*trigger_domain.Trigger, error) {
	trigger, err := s.repository.Get(context.Background(), triggerID)
	if err != nil {
		return nil, fmt.Errorf("get trigger: %w", err)
	}
	return trigger, nil
}

// Get retrieves all triggers
func (s *TriggerService) Get() ([]*trigger_domain.Trigger, error) {
	triggers, err := s.repository.GetAll(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get all triggers: %w", err)
	}
	return triggers, nil
}

// CancelTrigger cancels an active trigger
func (s *TriggerService) CancelTrigger(triggerID trigger_domain.TriggerID) (*trigger_domain.Trigger, error) {
	trigger, err := s.repository.Get(context.Background(), triggerID)
	if err != nil {
		return nil, fmt.Errorf("get trigger: %w", err)
	}

	if err := s.validateStatusTransition(trigger.Status, trigger_domain.StatusCancelled); err != nil {
		return nil, fmt.Errorf("invalid status transition: %w", err)
	}

	trigger.Status = trigger_domain.StatusCancelled
	err = s.repository.Update(context.Background(), trigger)
	if err != nil {
		return nil, fmt.Errorf("update trigger: %w", err)
	}

	updatedTrigger, err := s.repository.Get(context.Background(), triggerID)
	if err != nil {
		return nil, fmt.Errorf("get cancelled trigger: %w", err)
	}

	return updatedTrigger, nil
}

// CreateStopTrigger creates a new stop trigger with optional limit price
func (s *TriggerService) CreateStopTrigger(
	contract contract.ContractIdentifier,
	triggerPrice contract.ContractPrice,
	limitPrice *contract.ContractPrice,
) (*trigger_domain.Trigger, error) {
	// Create base stop trigger
	trigger, err := trigger_domain.NewStopTrigger(contract, triggerPrice, limitPrice)
	if err != nil {
		return nil, fmt.Errorf("create stop trigger: %w", err)
	}

	// Validate trigger
	if err := trigger_domain.ValidateStopTrigger(trigger); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTrigger, err)
	}

	// Save trigger
	err = s.repository.Save(context.Background(), trigger)
	if err != nil {
		return nil, fmt.Errorf("save trigger: %w", err)
	}

	savedTrigger, err := s.repository.Get(context.Background(), trigger.TriggerID)
	if err != nil {
		return nil, fmt.Errorf("get saved trigger: %w", err)
	}

	return savedTrigger, nil
}

// UpdateStopTrigger updates an existing stop trigger's prices
func (s *TriggerService) UpdateStopTrigger(
	triggerID trigger_domain.TriggerID,
	triggerPrice *contract.ContractPrice,
	limitPrice *contract.ContractPrice,
) (*trigger_domain.Trigger, error) {
	trigger, err := s.repository.Get(context.Background(), triggerID)
	if err != nil {
		return nil, fmt.Errorf("get trigger: %w", err)
	}

	// Verify it's a stop trigger
	if trigger.TriggerType != trigger_domain.TriggerTypeStop {
		return nil, fmt.Errorf("%w: expected stop trigger", ErrInvalidTriggerType)
	}

	// Update trigger price if provided
	if triggerPrice != nil {
		trigger.Condition.Price.Threshold = *triggerPrice
	}

	// Update limit price if provided
	if limitPrice != nil {
		trigger.Actions[0].LimitPrice = limitPrice
	}

	// Validate updated trigger
	if err := trigger_domain.ValidateStopTrigger(trigger); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTrigger, err)
	}

	// Save updates
	err = s.repository.Update(context.Background(), trigger)
	if err != nil {
		return nil, fmt.Errorf("update trigger: %w", err)
	}

	updatedTrigger, err := s.repository.Get(context.Background(), triggerID)
	if err != nil {
		return nil, fmt.Errorf("get updated trigger: %w", err)
	}

	return updatedTrigger, nil
}

// validateStatusTransition checks if a status transition is valid
func (s *TriggerService) validateStatusTransition(
	currentStatus trigger_domain.TriggerStatus,
	newStatus trigger_domain.TriggerStatus,
) error {
	// Validate that both statuses are valid
	if !currentStatus.IsValid() || !newStatus.IsValid() {
		return errors.New("invalid status")
	}

	// Cannot transition from a terminal state
	if currentStatus.IsTerminal() {
		return fmt.Errorf("cannot transition from terminal status %s", currentStatus)
	}

	// Cannot transition from non-active state to another non-active state
	if currentStatus != trigger_domain.StatusActive && newStatus != trigger_domain.StatusActive {
		return fmt.Errorf("invalid transition from %s to %s", currentStatus, newStatus)
	}

	return nil
}
