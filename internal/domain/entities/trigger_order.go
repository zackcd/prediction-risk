package entities

type Triggerable interface {
	TriggerPrice() ContractPrice
}
