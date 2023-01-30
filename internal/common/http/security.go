package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/twofas/2fas-server/config"
	"github.com/twofas/2fas-server/internal/common/api"
	"github.com/twofas/2fas-server/internal/common/logging"
	"net/http"
)

func IPWhitelistMiddleware(config config.SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestIp := c.ClientIP()

		if config.IsIpTrusted(requestIp) == false {
			err := errors.New("Request from not trusted IP " + requestIp)

			logging.Warning("Trying to access from untrusted IP ", requestIp)

			c.AbortWithStatusJSON(http.StatusForbidden, api.AccessForbiddenError(err))
		}
	}
}
