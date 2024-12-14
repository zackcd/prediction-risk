package entities

type StopLossOrderStatus string

// Active means the order is currently being monitored
// Executed means the order has been triggered
// Canceled means the order has been canceled
// Expired means the event has passed and the order is no longer valid
const (
	StatusActive   StopLossOrderStatus = "ACTIVE"
	StatusExecuted StopLossOrderStatus = "EXECUTED"
	StatusCanceled StopLossOrderStatus = "CANCELED"
	StatusExpired  StopLossOrderStatus = "EXPIRED"
)
