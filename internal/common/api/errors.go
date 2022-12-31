package api

import (
	"fmt"
	"net/http"
)

type ApiError struct {
	Code        int
	Type        string
	Description string
	Reason      string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("%d - %s", e.Code, e.Type)
}

func NewBadRequestError(err error) error {
	return &ApiError{
		Code:        http.StatusBadRequest,
		Type:        "BadRequest",
		Description: "Malformed request syntax.",
		Reason:      err.Error(),
	}
}

func NewInternalServerError(err error) error {
	return &ApiError{
		Code:        http.StatusInternalServerError,
		Type:        "InternalServerError",
		Description: "Unexpected condition was encounteredn",
		Reason:      err.Error(),
	}
}

func NotFoundError(err error) error {
	return &ApiError{
		Code:        http.StatusNotFound,
		Type:        "NotFound",
		Description: "Requested resource can not be found",
		Reason:      err.Error(),
	}
}

func AccessForbiddenError(err error) error {
	return &ApiError{
		Code:        http.StatusForbidden,
		Type:        "AccessForbidden",
		Description: "You are not allowed to access requested resource",
		Reason:      err.Error(),
	}
}

func ConflictError(err error) error {
	return &ApiError{
		Code:        http.StatusConflict,
		Type:        "Conflict",
		Description: "The request could not be completed due to a conflict with the current state of the target resource",
		Reason:      err.Error(),
	}
}

func GoneError(err error) error {
	return &ApiError{
		Code:        http.StatusGone,
		Type:        "Gone",
		Description: "Access is no longer available",
		Reason:      err.Error(),
	}
}
