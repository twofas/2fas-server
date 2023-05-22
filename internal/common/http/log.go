package http

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/twofas/2fas-server/internal/common/logging"
)

func RequestJsonLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer

		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := io.ReadAll(tee)

		c.Request.Body = io.NopCloser(&buf)

		logging.WithFields(logging.Fields{
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"headers":        c.Request.Header,
			"body":           string(body),
			"request_id":     c.GetString(RequestIdKey),
			"correlation_id": c.GetString(CorrelationIdKey),
		}).Info("Request")

		c.Next()

		logging.WithFields(logging.Fields{
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"request_id":     c.GetString(RequestIdKey),
			"correlation_id": c.GetString(CorrelationIdKey),
		}).Info("Response")
	}
}
