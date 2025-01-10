package order

import "prediction-risk/internal/domain/contract"

type LimitOrder struct {
	limitPrice contract.ContractPrice
}

func (o *LimitOrder) GetLimitPrice() contract.ContractPrice {
	return o.limitPrice
}

func (o *LimitOrder) SetLimitPrice(limitPrice contract.ContractPrice) {
	o.limitPrice = limitPrice
}

type Limitable interface {
	LimitPrice() contract.ContractPrice
	SetLimitPrice(contract.ContractPrice)
}
