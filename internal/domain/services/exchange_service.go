package services

import (
	"fmt"
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/infrastructure/external/kalshi"
)

type ExchangeService struct {
	kalshiClient *kalshi.KalshiClient
}

func NewExchangeService(
	kalshiClient *kalshi.KalshiClient,
) *ExchangeService {
	return &ExchangeService{
		kalshiClient: kalshiClient,
	}
}

func (s *ExchangeService) GetPositions() {
	result := &PositionsResult{
        MarketPositions: make([]kalshi.MarketPosition, 0),
        EventPositions:  make([]kalshi.EventPosition, 0),
    }

    if err := s.getPositionsRecursive(initialParams, result); err != nil {
        return nil, fmt.Errorf("fetching positions: %w", err)
    }

    return result, nil
}

// Recursively get all positions
func (s *ExchangeService) getPositionsRecursive(
	params *kalshi.PositionsParams,
	marketPositions []kalshi.MarketPosition,
	eventPositions []kalshi.EventPosition,
) ([]kalshi.MarketPosition, []kalshi.EventPosition, error) {
	resp, err := s.kalshiClient.Portfolio.GetPositions(params)
	if err != nil {
		return nil, nil, err
	}

	marketPositions = append(marketPositions, resp.MarketPositions...)
	eventPositions = append(eventPositions, resp.EventPositions...)

	if resp.Cursor == nil {
		return marketPositions, eventPositions, nil
	}

	return s.getPositionsRecursive(
		params.WithCursor(*resp.Cursor),
		marketPositions,
		eventPositions,
	)
}

func (s *ExchangeService) CreateSellOrder(
	ticker string,
	side entities.Side,
	orderID string,
	count int,
) (*entities.Order, error) {
	// Convert entities.Side to kalshi.OrderSide
	var kalshiOrderSide kalshi.OrderSide
	if side == entities.SideYes {
		kalshiOrderSide = kalshi.OrderSideYes
	} else {
		kalshiOrderSide = kalshi.OrderSideNo
	}

	request := &kalshi.CreateOrderRequest{
		Ticker:            ticker,
		ClientOrderID:     orderID,
		Side:              kalshiOrderSide,
		Action:            kalshi.OrderActionSell,
		Count:             count,
		Type:              "market",
		YesPrice:          nil,
		NoPrice:           nil,
		ExpirationTs:      nil,
		SellPositionFloor: nil,
		BuyMaxCost:        nil,
	}

	// Call the Kalshi API
	orderResp, err := s.kalshiClient.Portfolio.CreateOrder(request)
	if err != nil {
		// Handle error
	}

	return &entities.Order{
		ExchangeOrderID: orderResp.Order.ID,
		Exchange:        entities.ExchangeKalshi,
		InternalOrderID: orderID,
		Ticker:          ticker,
		Side:            side,
		Action:          entities.OrderActionSell,
	}, nil
}
