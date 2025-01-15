package exchange_service

import (
	"fmt"
	"prediction-risk/internal/app/contract"
	exchange_domain "prediction-risk/internal/app/exchange/domain"
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"

	"github.com/samber/lo"
)

type marketGetter interface {
	GetMarket(ticker string) (*kalshi.MarketResponse, error)
}

type positionGetter interface {
	GetPositions(params kalshi.GetPositionsOptions) (*kalshi.PositionsResult, error)
}

type orderCreator interface {
	CreateOrder(request kalshi.CreateOrderRequest) (*kalshi.CreateOrderResponse, error)
}

type KalshiExchangeService struct {
	markets   marketGetter
	positions positionGetter
	orders    orderCreator
}

func NewExchangeService(kalshiClient *kalshi.KalshiClient) *KalshiExchangeService {
	return &KalshiExchangeService{
		markets:   kalshiClient.Market,
		positions: kalshiClient.Portfolio,
		orders:    kalshiClient.Portfolio,
	}
}

func (es *KalshiExchangeService) GetMarket(ticker string) (*exchange_domain.Market, error) {
	_, err := es.markets.GetMarket(ticker)
	if err != nil {
		return nil, err
	}

	// TODO: Implement market mapping
	market := exchange_domain.Market{}
	return &market, nil
}

func (es *KalshiExchangeService) GetPositions() ([]*exchange_domain.Position, error) {
	params := kalshi.GetPositionsOptions{}
	resp, err := es.positions.GetPositions(params)
	if err != nil {
		return nil, err
	}

	positions := lo.Map(resp.MarketPositions, func(p kalshi.MarketPosition, _ int) *exchange_domain.Position {
		// TODO: Implement position mapping
		return &exchange_domain.Position{}
	})

	return positions, nil
}

func (es *KalshiExchangeService) CreateOrder(
	orderParams OrderParams,
) (*exchange_domain.Order, error) {
	switch orderParams.Action {
	case exchange_domain.OrderActionBuy:
		return es.createBuyOrder(
			orderParams.ContractID,
			orderParams.Reference,
			orderParams.LimitPrice,
		)
	case exchange_domain.OrderActionSell:
		return es.createSellOrder(
			orderParams.ContractID,
			orderParams.Reference,
			orderParams.Quantity,
			orderParams.LimitPrice,
		)
	default:
		return nil, fmt.Errorf("invalid order action: %s", orderParams.Action)
	}
}

func (es *KalshiExchangeService) createBuyOrder(
	contractID contract.ContractIdentifier,
	reference string,
	limitPrice *contract.ContractPrice,
) (*exchange_domain.Order, error) {
	return nil, fmt.Errorf("buy orders not yet supported")
}

func (es *KalshiExchangeService) createSellOrder(
	contractID contract.ContractIdentifier,
	reference string,
	quantity *uint,
	limitPrice *contract.ContractPrice,
) (*exchange_domain.Order, error) {
	position, err := es.findPosition(contractID.Ticker)
	if err != nil {
		return nil, fmt.Errorf("find position: %w", err)
	}

	var orderSide kalshi.OrderSide
	if contractID.Side == contract.SideYes {
		orderSide = kalshi.OrderSideYes
	} else {
		orderSide = kalshi.OrderSideNo
	}

	var orderType string
	var marketOrderType exchange_domain.MarketOrderType
	var yesPrice *int
	var noPrice *int
	if limitPrice != nil {
		orderType = "limit"
		marketOrderType = exchange_domain.OrderTypeLimit
		value := limitPrice.Value()
		if contractID.Side == contract.SideYes {
			yesPrice = &value
		} else {
			noPrice = &value
		}
	} else {
		orderType = "market"
		marketOrderType = exchange_domain.OrderTypeMarket
	}

	sellQuantity, err := es.calculateSellQuantity(position.Position, quantity)
	if err != nil {
		return nil, fmt.Errorf("calculate sell quantity: %w", err)
	}

	request := kalshi.CreateOrderRequest{
		Ticker:        string(contractID.Ticker),
		ClientOrderID: reference,
		Side:          orderSide,
		Action:        kalshi.OrderActionSell,
		Count:         int(sellQuantity),
		Type:          orderType,
		YesPrice:      yesPrice,
		NoPrice:       noPrice,
	}
	resp, err := es.orders.CreateOrder(request)
	if err != nil {
		return nil, err
	}

	order := exchange_domain.NewOrder(
		resp.Order.ID,
		exchange_domain.ExchangeKalshi,
		reference,
		resp.Order.Ticker,
		contractID.Side,
		exchange_domain.OrderActionSell,
		marketOrderType,
		resp.Order.Status,
	)

	return order, nil
}

func (es *KalshiExchangeService) findPosition(ticker contract.Ticker) (*kalshi.MarketPosition, error) {
	tickerStr := string(ticker)
	positions, err := es.positions.GetPositions(kalshi.GetPositionsOptions{Ticker: &tickerStr})
	if err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}

	// Check if the position is available
	if positions == nil {
		return nil, fmt.Errorf("position not found for ticker: %s", tickerStr)
	}

	// Check if the position present
	position, isPresent := lo.Find(positions.MarketPositions, func(p kalshi.MarketPosition) bool {
		return p.Ticker == tickerStr
	})
	if !isPresent {
		return nil, fmt.Errorf("position not found for ticker: %s", tickerStr)
	}

	return &position, nil
}

// If size is specified, it will be set to the minimum of the position and the size
// Otherwise it will be set to the full position
func (es *KalshiExchangeService) calculateSellQuantity(
	position int,
	requestedQuantity *uint,
) (uint, error) {
	if position <= 0 {
		return 0, fmt.Errorf("position must be positive, got %d", position)
	}

	availableQuantity := uint(abs(position))
	if requestedQuantity == nil {
		return availableQuantity, nil
	}

	return min(availableQuantity, *requestedQuantity), nil
}

func abs(x int) uint {
	if x < 0 {
		return uint(-x)
	}
	return uint(x)
}

func min(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}
