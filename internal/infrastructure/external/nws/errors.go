package nws

import "fmt"

type NWSError struct {
	Reason     string
	StatusCode int
	Body       string
}

func (e *NWSError) Error() string {
	return fmt.Sprintf("NWSError(%d %s): %s", e.StatusCode, e.Reason, e.Body)
}
