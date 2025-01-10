package order

import (
	"prediction-risk/internal/domain/contract"
	"prediction-risk/internal/domain/exchange"
	"prediction-risk/internal/infrastructure/external/kalshi"
)

// ExchangeProvider defines the interface for exchange operations needed by StopOrderService
type ExchangeProvider interface {
	GetPositions() (*kalshi.PositionsResult, error)
	CreateSellOrder(ticker string, count int, side contract.Side, orderID string, limitPrice *contract.ContractPrice) (*exchange.ExchangeOrder, error)
}
