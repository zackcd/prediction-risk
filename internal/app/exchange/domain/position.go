package exchange_domain

import "prediction-risk/internal/app/contract"

type Position struct {
	ContractID contract.ContractIdentifier
	Quantity   uint
}
