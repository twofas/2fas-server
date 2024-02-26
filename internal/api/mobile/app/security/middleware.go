package security

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/rate_limit"
)

const defaultMobileApiBandwidthAbuseThreshold = 100

func MobileIpAbuseAuditMiddleware(rateLimiter rate_limit.RateLimiter, rateLimitValue int) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceId := c.Param("device_id")
		extensionId := c.Param("extension_id")

		if deviceId == "" && extensionId == "" {
			return
		}

		key := strings.TrimSuffix(
			fmt.Sprintf("security.api.mobile.bandwidth.%s.%s", deviceId, extensionId),
			".",
		)
		limitValue := rateLimitValue
		if limitValue == 0 {
			limitValue = defaultMobileApiBandwidthAbuseThreshold
		}
		rate := rate_limit.Rate{
			TimeUnit: time.Minute,
			Limit:    limitValue,
		}

		limitReached := rateLimiter.Test(c, key, rate)

		if limitReached {
			logging.FromContext(c.Request.Context()).WithFields(logging.Fields{
				"type":                 "security",
				"uri":                  c.Request.URL.String(),
				"device_id":            deviceId,
				"browser_extension_id": extensionId,
				"ip":                   c.ClientIP(),
			}).Warning("API potentially abused at mobile scope, blocking")
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}
