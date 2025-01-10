package contract

import "fmt"

// Price of a contract in cents
// Contracts can be priced between $0.00 and $1.00, in increments of $0.01
// To make operations more accurate, we store the price in cents
type ContractPrice int

func NewContractPrice(value int) (ContractPrice, error) {
	if value < 0 || value > 100 {
		return 0, fmt.Errorf("contract price must be between 0 and 100, got: %d", value)
	}
	return ContractPrice(value), nil
}

func (p ContractPrice) Value() int {
	return int(p)
}

func (p ContractPrice) IsValid() bool {
	return p >= 0 && p <= 100
}

type side string

type Side interface {
	String() string
}

const (
	SideYes side = "YES"
	SideNo  side = "NO"
)

func (s side) String() string {
	return string(s)
}

func NewSide(value string) (Side, error) {
	switch value {
	case string(SideYes), string(SideNo):
		return side(value), nil
	default:
		return nil, fmt.Errorf("invalid side value: %q", value)
	}
}
