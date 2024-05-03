package http

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/twofas/2fas-server/internal/common/logging"
)

const (
	CorrelationIdHeader = "X-Correlation-ID"
)

var (
	RequestId     string
	CorrelationId string
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := uuid.New().String()
		correlationId := c.Request.Header.Get(CorrelationIdHeader)
		if correlationId == "" {
			correlationId = uuid.New().String()
		}

		ctxWithLog := logging.AddToContext(c.Request.Context(), logging.WithFields(map[string]any{
			"correlation_id": correlationId,
			"request_id":     requestId,
		}))
		c.Request = c.Request.WithContext(ctxWithLog)

	}
}

func RequestJsonLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer

		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := io.ReadAll(tee)

		c.Request.Body = io.NopCloser(&buf)

		log := logging.FromContext(c.Request.Context())

		log.WithFields(logging.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"body":   string(body),
		}).Info("Request")

		c.Next()

		log.WithFields(logging.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		}).Info("Response")
	}
}
