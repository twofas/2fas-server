package http

import (
	"github.com/gin-gonic/gin"
	"github.com/twofas/2fas-server/internal/common/logging"
)

func RequestJsonLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestLogger := logging.WithFields(logging.Fields{
			"client_ip":      c.ClientIP(),
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"request_id":     c.GetString(RequestIdKey),
			"correlation_id": c.GetString(CorrelationIdKey),
		})

		requestLogger.Info("Request")

		c.Next()

		requestLogger.Info("Response")
	}
}
