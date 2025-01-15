package trigger_service

import (
	"fmt"
	exchange_domain "prediction-risk/internal/app/exchange/domain"
	exchange_service "prediction-risk/internal/app/exchange/service"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
)

type TriggerExecutor struct {
	triggerService  *TriggerService
	exchangeService exchange_service.ExchangeService
}

func NewTriggerExecutor(
	triggerService *TriggerService,
	exchangeService exchange_service.ExchangeService,
) *TriggerExecutor {
	return &TriggerExecutor{
		triggerService:  triggerService,
		exchangeService: exchangeService,
	}
}

func (t *TriggerExecutor) ExecuteTrigger(trigger *trigger_domain.Trigger) (*trigger_domain.Trigger, error) {
	// Check if the trigger is exeutable (it must be active)
	if trigger.Status != trigger_domain.StatusActive {
		return nil, fmt.Errorf("trigger is not active, status: %s", trigger.Status)
	}

	// Execute all the actions in the trigger
	_, err := t.executeActions(trigger.TriggerID, trigger.Actions)
	if err != nil {
		return nil, fmt.Errorf("execute actions: %w", err)
	}

	// Update the trigger status to executed
	updatedTrigger, err := t.triggerService.UpdateTriggerStatus(trigger.TriggerID, trigger_domain.StatusTriggered)
	if err != nil {
		return nil, fmt.Errorf("update trigger status: %w", err)
	}

	return updatedTrigger, nil
}

func (t *TriggerExecutor) executeActions(
	triggerID trigger_domain.TriggerID,
	actions []trigger_domain.TriggerAction,
) ([]*exchange_domain.Order, error) {
	var orders []*exchange_domain.Order
	for _, action := range actions {
		order, err := t.executeAction(triggerID, action)
		if err != nil {
			return nil, fmt.Errorf("execute action: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (t *TriggerExecutor) executeAction(
	triggerID trigger_domain.TriggerID,
	action trigger_domain.TriggerAction,
) (*exchange_domain.Order, error) {
	// Map trigger action side to exchange order action
	var orderAction exchange_domain.OrderAction
	switch action.Side {
	case trigger_domain.Buy:
		orderAction = exchange_domain.OrderActionBuy
	case trigger_domain.Sell:
		orderAction = exchange_domain.OrderActionSell
	default:
		return nil, fmt.Errorf("invalid action side: %s", action.Side)
	}

	// Create the order parameters once we know the action is valid
	orderParams := exchange_service.OrderParams{
		ContractID: action.Contract,
		Quantity:   action.Size,
		Action:     orderAction,
		Reference:  triggerID.String(),
		LimitPrice: action.LimitPrice,
	}

	order, err := t.exchangeService.CreateOrder(orderParams)
	if err != nil {
		return nil, fmt.Errorf("create %s order: %w", string(orderAction), err)
	}

	return order, nil
}
