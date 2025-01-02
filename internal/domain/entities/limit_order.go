package entities

type LimitOrder struct {
	limitPrice ContractPrice
}

func (o *LimitOrder) GetLimitPrice() ContractPrice {
	return o.limitPrice
}

func (o *LimitOrder) SetLimitPrice(limitPrice ContractPrice) {
	o.limitPrice = limitPrice
}

type Limitable interface {
	LimitPrice() ContractPrice
	SetLimitPrice(ContractPrice)
}
