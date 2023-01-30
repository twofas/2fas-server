package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/twofas/2fas-server/internal/common/logging"
)

const (
	RequestIdKey = "request_id"

	CorrelationIdKey    = "correlation_id"
	CorrelationIdHeader = "X-Correlation-ID"
)

var (
	RequestId     string
	CorrelationId string
)

func RequestIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		RequestId = uuid.New().String()

		c.Set(RequestIdKey, RequestId)

		logging.WithDefaultField(RequestIdKey, RequestId)
	}
}

func CorrelationIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(CorrelationIdKey, uuid.New().String())

		CorrelationId = c.Request.Header.Get(CorrelationIdHeader)

		if CorrelationId == "" {
			CorrelationId = uuid.New().String()
		}

		logging.WithDefaultField(CorrelationIdKey, CorrelationId)

		c.Set(CorrelationIdKey, CorrelationId)
	}
}
