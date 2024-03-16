package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BodySizeLimitMiddleware(requestBytesLimit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		var w http.ResponseWriter = c.Writer

		c.Request.Body = http.MaxBytesReader(w, c.Request.Body, requestBytesLimit)
	}
}
