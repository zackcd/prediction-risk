package entities

type StopLossOrder struct {
	order
}

func NewStopLossOrder(
	ticker string,
	side Side,
	triggerPrice ContractPrice,
) *StopLossOrder {
	return &StopLossOrder{
		order: newOrder(OrderTypeStopLoss, ticker, side, triggerPrice),
	}
}

var _ Order = (*StopLossOrder)(nil)
