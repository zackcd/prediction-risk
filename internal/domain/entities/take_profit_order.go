package entities

type TakeProfitOrder struct {
	order
}

func NewTakeProfitOrder(
	ticker string,
	side Side,
	triggerPrice ContractPrice,
) *TakeProfitOrder {
	return &TakeProfitOrder{
		order: newOrder(OrderTypeTakeProfit, ticker, side, triggerPrice),
	}
}

var _ Order = (*TakeProfitOrder)(nil)
