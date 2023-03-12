package validation

import (
	"github.com/go-playground/validator/v10"
	"strings"
)

func NotBlank(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}
