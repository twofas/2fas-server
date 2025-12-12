package http

import "fmt"

type ResponseError struct {
	Status  int
	Message string `json:"message"`
}

func (error *ResponseError) Error() string {
	return fmt.Sprintf("Status: %d Message: %s", error.Status, error.Message)
}
