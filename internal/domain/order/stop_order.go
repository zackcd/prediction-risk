package order

import (
	"prediction-risk/internal/domain/contract"
)

type StopOrder struct {
	order
	triggerPrice contract.ContractPrice
	limitPrice   *contract.ContractPrice
}

func NewStopOrder(
	ticker string,
	side contract.Side,
	triggerPrice contract.ContractPrice,
	limitPrice *contract.ContractPrice,
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

func (o *StopOrder) TriggerPrice() contract.ContractPrice {
	return o.triggerPrice
}

func (o *StopOrder) SetTriggerPrice(price contract.ContractPrice) {
	o.triggerPrice = price
}

func (o *StopOrder) LimitPrice() *contract.ContractPrice {
	return o.limitPrice
}

func (o *StopOrder) SetLimitPrice(price *contract.ContractPrice) {
	o.limitPrice = price
}
