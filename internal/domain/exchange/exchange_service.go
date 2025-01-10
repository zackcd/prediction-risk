package exchange

import (
	"prediction-risk/internal/domain/contract"
	"prediction-risk/internal/infrastructure/external/kalshi"
)

type ExchangeService interface {
	GetMarket(ticker string) (*kalshi.Market, error)
	GetPositions() (*kalshi.PositionsResult, error)
	CreateSellOrder(ticker string, count int, side contract.Side, orderID string, limitPrice *contract.ContractPrice) (*ExchangeOrder, error)
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
	orderID string,
	limitPrice *contract.ContractPrice,
) (*ExchangeOrder, error) {
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
		ClientOrderID: orderID,
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

	return &ExchangeOrder{
		ExchangeOrderID: resp.Order.ID,
		Exchange:        ExchangeKalshi,
		InternalOrderID: resp.Order.ClientOrderID,
		Ticker:          resp.Order.Ticker,
		Side:            side,
		Action:          OrderActionSell,
		OrderType:       OrderTypeMarket,
		Status:          resp.Order.Status,
	}, nil
}
