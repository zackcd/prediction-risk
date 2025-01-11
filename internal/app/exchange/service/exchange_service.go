package exchange_service

import (
	"prediction-risk/internal/app/contract"
	exchange_domain "prediction-risk/internal/app/exchange/domain"
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"
)

type ExchangeService interface {
	GetMarket(ticker string) (*kalshi.Market, error)
	GetPositions() (*kalshi.PositionsResult, error)
	CreateSellOrder(ticker string, count int, side contract.Side, limitPrice *contract.ContractPrice, orderID *exchange_domain.OrderID) (*exchange_domain.Order, error)
}

type MarketGetter interface {
	GetMarket(ticker string) (*kalshi.MarketResponse, error)
}

type PortfolioManager interface {
	GetPositions(opts kalshi.GetPositionsOptions) (*kalshi.PositionsResult, error)
	CreateOrder(order kalshi.CreateOrderRequest) (*kalshi.CreateOrderResponse, error)
}

type exchangeService struct {
	market    MarketGetter
	portfolio PortfolioManager
}

func NewExchangeService(market MarketGetter, portfolio PortfolioManager) ExchangeService {
	return &exchangeService{
		market:    market,
		portfolio: portfolio,
	}
}

func (es *exchangeService) GetMarket(ticker string) (*kalshi.Market, error) {
	resp, err := es.market.GetMarket(ticker)
	if err != nil {
		return nil, err
	}
	return &resp.Market, nil
}

func (es *exchangeService) GetPositions() (*kalshi.PositionsResult, error) {
	params := kalshi.GetPositionsOptions{}
	resp, err := es.portfolio.GetPositions(params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (es *exchangeService) CreateSellOrder(
	ticker string,
	count int,
	side contract.Side,
	limitPrice *contract.ContractPrice,
	orderID *exchange_domain.OrderID,
) (*exchange_domain.Order, error) {
	if orderID == nil {
		newID := exchange_domain.NewOrderID()
		orderID = &newID
	}

	var orderSide kalshi.OrderSide
	if side == contract.SideYes {
		orderSide = kalshi.OrderSideYes
	} else {
		orderSide = kalshi.OrderSideNo
	}

	var orderType string
	var yesPrice *int
	var noPrice *int
	if limitPrice != nil {
		orderType = "limit"
		value := limitPrice.Value()
		if side == contract.SideYes {
			yesPrice = &value
		} else {
			noPrice = &value
		}
	} else {
		orderType = "market"
	}

	request := kalshi.CreateOrderRequest{
		Ticker:        ticker,
		ClientOrderID: orderID.String(),
		Side:          orderSide,
		Action:        kalshi.OrderActionSell,
		Count:         count,
		Type:          orderType,
		YesPrice:      yesPrice,
		NoPrice:       noPrice,
	}
	resp, err := es.portfolio.CreateOrder(request)
	if err != nil {
		return nil, err
	}

	return &exchange_domain.Order{
		OrderID:         *orderID,
		ExchangeOrderID: resp.Order.ID,
		Exchange:        exchange_domain.ExchangeKalshi,
		Ticker:          resp.Order.Ticker,
		Side:            side,
		Action:          exchange_domain.OrderActionSell,
		OrderType:       exchange_domain.OrderTypeMarket,
		Status:          resp.Order.Status,
	}, nil
}
