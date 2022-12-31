package http

import "fmt"

type ErrorResponse struct {
	Status  int
	Message string `json:"message"`
}

func (error *ErrorResponse) Error() string {
	return fmt.Sprintf("Status: %d Message: %s", error.Status, error.Message)
}
