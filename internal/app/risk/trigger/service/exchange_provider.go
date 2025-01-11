package trigger_service

import (
	"prediction-risk/internal/app/contract"
	exchange_domain "prediction-risk/internal/app/exchange/domain"
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"
)

// ExchangeProvider defines the interface for exchange operations needed by StopOrderService
type ExchangeProvider interface {
	GetPositions() (*kalshi.PositionsResult, error)
	CreateSellOrder(
		ticker string,
		count int,
		side contract.Side,
		orderID string,
		limitPrice *contract.ContractPrice,
	) (*exchange_domain.Order, error)
}
