package security

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/rate_limit"
	"time"
)

var browserExtensionApiBandwidthAbuseThreshold = 100

func BrowserExtensionBandwidthAuditMiddleware(rateLimiter rate_limit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		extensionId := c.Param("extension_id")

		if extensionId == "" {
			return
		}

		key := fmt.Sprintf("security.api.browser_extension.bandwidth.%s", extensionId)

		rate := rate_limit.Rate{
			TimeUnit: time.Minute,
			Limit:    browserExtensionApiBandwidthAbuseThreshold,
		}

		limitReached := rateLimiter.Test(context.Background(), key, rate)

		if limitReached {
			logging.WithFields(logging.Fields{
				"type":                 "security",
				"uri":                  c.Request.URL.String(),
				"browser_extension_id": extensionId,
				"ip":                   c.ClientIP(),
			}).Warning("API potentially abused at Browser Extension scope")
		}
	}
}
