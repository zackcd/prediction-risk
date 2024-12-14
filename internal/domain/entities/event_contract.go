package entities

// Price of a contract in cents
// Contracts can be priced between $0.00 and $1.00, in increments of $0.01
// To make operations more accurate, we store the price in cents
type ContractPrice int

func NewContractPrice(value int) ContractPrice {
	if value < 0 || value > 100 {
		panic("invalid contract price")
	}

	return ContractPrice(value)
}

func (p ContractPrice) Value() int {
	return p.Value()
}

type Side string

const (
	SideYes Side = "YES"
	SideNo  Side = "NO"
)
