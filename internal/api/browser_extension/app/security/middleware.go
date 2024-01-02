package security

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/rate_limit"
)

const defaultBrowserExtensionApiBandwidthAbuseThreshold = 100

func BrowserExtensionBandwidthAuditMiddleware(rateLimiter rate_limit.RateLimiter, rateLimitValue int) gin.HandlerFunc {
	return func(c *gin.Context) {
		extensionId := c.Param("extension_id")

		if extensionId == "" {
			return
		}

		key := fmt.Sprintf("security.api.browser_extension.bandwidth.%s", extensionId)

		limitValue := rateLimitValue
		if limitValue == 0 {
			limitValue = defaultBrowserExtensionApiBandwidthAbuseThreshold
		}
		rate := rate_limit.Rate{
			TimeUnit: time.Minute,
			Limit:    limitValue,
		}
		limitReached := rateLimiter.Test(c, key, rate)

		if limitReached {
			logging.WithFields(logging.Fields{
				"type":                 "security",
				"uri":                  c.Request.URL.String(),
				"browser_extension_id": extensionId,
				"ip":                   c.ClientIP(),
			}).Warning("API potentially abused at Browser Extension scope, blocking")
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}
