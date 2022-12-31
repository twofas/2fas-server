package http

import (
	"errors"
	"github.com/2fas/api/config"
	"github.com/2fas/api/internal/common/api"
	"github.com/2fas/api/internal/common/logging"
	"github.com/gin-gonic/gin"
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
