package exchange_domain

import (
	"prediction-risk/internal/app/contract"
	"time"
)

// Market represents the complete state of a tradeable market
type Market struct {
	Ticker      contract.Ticker // Unique market identifier
	Info        MarketInfo
	Status      MarketStatus
	Pricing     MarketPricing
	Constraints TradingConstraints
	Liquidity   LiquidityMetrics
}

// MarketInfo holds the identifying and descriptive information about a market
type MarketInfo struct {
	Title    string     // Human-readable market title
	Category string     // Market category (e.g., "Sports", "Politics")
	Type     MarketType // Type of market (e.g., Binary, Numeric)
}

// MarketStatus tracks the current state and timing of a market
type MarketStatus struct {
	State              MarketState // Current market state (Active, Closed, Settled)
	OpenTime           time.Time   // When trading begins
	CloseTime          time.Time   // When trading ends
	ExpirationTime     time.Time   // When the market resolves
	SettlementTime     *time.Time  // When the market was settled (if applicable)
	Result             *string     // Market outcome (if settled)
	AllowsEarlyClosing bool        // Whether the market can close before expiration
}

// MarketPricing holds the current price information for a market
type MarketPricing struct {
	YesSide PricingSide // Pricing for YES contracts
	NoSide  PricingSide // Pricing for NO contracts
}

// PricingSide represents the pricing for one side of a binary market
type PricingSide struct {
	Bid         contract.ContractPrice // Highest buy price
	Ask         contract.ContractPrice // Lowest sell price
	LastPrice   contract.ContractPrice // Last traded price
	PreviousBid contract.ContractPrice // Previous period's bid
	PreviousAsk contract.ContractPrice // Previous period's ask
}

// TradingConstraints defines the rules and limitations for trading
type TradingConstraints struct {
	NotionalValue contract.ContractPrice // Contract full value
	TickSize      contract.ContractPrice // Minimum price movement
	RiskLimit     contract.ContractPrice // Maximum position size in cents
}

// LiquidityMetrics provides information about market activity
type LiquidityMetrics struct {
	Volume       int // Total volume
	Volume24H    int // 24-hour volume
	OpenInterest int // Outstanding contracts
	Liquidity    int // Available liquidity
}

type MarketState string

const (
	MarketStateUnopened MarketState = "unopened"
	MarketStateOpen     MarketState = "open"
	MarketStateClosed   MarketState = "closed"
	MarketStateSettled  MarketState = "settled"
)

// MarketType represents the type of market
type MarketType string

const (
	MarketTypeBinary MarketType = "BINARY"
	MarketTypeScalar MarketType = "SCALAR"
)
