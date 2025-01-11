package kalshi

import "fmt"

type KalshiError struct {
	Reason     string
	StatusCode int
	Body       string
}

func (e *KalshiError) Error() string {
	return fmt.Sprintf("KalshiError(%d %s): %s", e.StatusCode, e.Reason, e.Body)
}
