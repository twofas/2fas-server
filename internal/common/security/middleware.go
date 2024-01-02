package security

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/rate_limit"
)

const defaultAPIBandwidthAbuseThreshold = 100

func IPAbuseAuditMiddleware(rateLimiter rate_limit.RateLimiter, rateLimitValue int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIp := c.ClientIP()

		key := fmt.Sprintf("security.api.ip_bandwidth_audit.%s", clientIp)

		limitValue := rateLimitValue
		if limitValue == 0 {
			limitValue = defaultAPIBandwidthAbuseThreshold
		}
		rate := rate_limit.Rate{
			TimeUnit: time.Minute,
			Limit:    limitValue,
		}

		limitReached := rateLimiter.Test(c, key, rate)

		if limitReached {
			logging.WithFields(logging.Fields{
				"type": "security",
				"uri":  c.Request.URL.String(),
				"ip":   c.ClientIP(),
			}).Warning("API potentially abused by Client IP, blocking")
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}
