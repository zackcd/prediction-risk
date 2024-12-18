package kalshi

import "fmt"

type HTTPError struct {
	Reason     string
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTPError(%d %s): %s", e.StatusCode, e.Reason, e.Body)
}
