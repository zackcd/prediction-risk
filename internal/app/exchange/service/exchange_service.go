package exchange_service

import (
	"prediction-risk/internal/app/contract"
	exchange_domain "prediction-risk/internal/app/exchange/domain"
)

type OrderParams struct {
	ContractID contract.ContractIdentifier
	Quantity   *uint
	Action     exchange_domain.OrderAction
	Reference  string
	LimitPrice *contract.ContractPrice
	// Future fields can be added without breaking the interface
}

type ExchangeService interface {
	GetMarket(ticker contract.Ticker) (*exchange_domain.Market, error)
	GetPositions() (*exchange_domain.Position, error)
	CreateOrder(orderParams OrderParams) (*exchange_domain.Order, error)
}
