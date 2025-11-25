package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/twofas/2fas-server/internal/common/api"
	"github.com/twofas/2fas-server/internal/common/logging"
)

func Validate(c *gin.Context, v *validator.Validate, a any) bool {
	err := v.Struct(a)

	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			logging.FromContext(c.Request.Context()).Errorf("unexpected validation error: %v", err)
			c.JSON(500, api.NewInternalServerError(fmt.Errorf("unexpected validation error")))
			return false
		}

		c.JSON(400, api.NewBadRequestError(validationErrors))
		return false
	}

	return true
}
