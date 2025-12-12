package db

import (
	"errors"
	"fmt"
)

var dbError = errors.New("database error") //nolint:errname

func WrapError(err error) error {
	return fmt.Errorf("%w: %w", dbError, err)
}

func QueryPrepError(err error) error {
	return WrapError(fmt.Errorf("failed to prepare query: %w", err))
}

func IsDBError(err error) bool {
	return errors.Is(err, dbError)
}
