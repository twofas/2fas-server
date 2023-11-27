package db

import (
	"fmt"
	"testing"
)

func TestIsDBError(t *testing.T) {
	tests := []struct {
		name      string
		target    error
		isDBError bool
	}{
		{
			name:      "plain error",
			target:    fmt.Errorf("error"),
			isDBError: false,
		},
		{
			name:      "wrapped error",
			target:    WrapError(fmt.Errorf("error")),
			isDBError: true,
		},
		{
			name:      "error wrapped twice",
			target:    fmt.Errorf("error: %w", WrapError(fmt.Errorf("error"))),
			isDBError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsDBError(tc.target)
			if tc.isDBError != result {
				t.Fatalf("For %v, expected %v, got %v", tc.target, tc.isDBError, result)
			}
		})
	}

}
