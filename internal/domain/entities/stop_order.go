package entities

type StopOrder struct {
	order
	triggerPrice ContractPrice
	limitPrice   *ContractPrice
}

func NewStopOrder(
	ticker string,
	side Side,
	triggerPrice ContractPrice,
	limitPrice *ContractPrice,
	orderId *OrderID,
) *StopOrder {
	order := newOrder(OrderTypeStop, ticker, side, orderId)
	return &StopOrder{
		order,
		triggerPrice,
		limitPrice,
	}
}

func (o *StopOrder) Order() order {
	return o.order
}

func (o *StopOrder) TriggerPrice() ContractPrice {
	return o.triggerPrice
}

func (o *StopOrder) SetTriggerPrice(price ContractPrice) {
	o.triggerPrice = price
}

func (o *StopOrder) LimitPrice() *ContractPrice {
	return o.limitPrice
}

func (o *StopOrder) SetLimitPrice(price *ContractPrice) {
	o.limitPrice = price
}
