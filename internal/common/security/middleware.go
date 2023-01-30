package security

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/common/rate_limit"
	"time"
)

var apiBandwidthAbuseThreshold = 100

func IPAbuseAuditMiddleware(rateLimiter rate_limit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIp := c.ClientIP()

		key := fmt.Sprintf("security.api.ip_bandwidth_audit.%s", clientIp)

		rate := rate_limit.Rate{
			TimeUnit: time.Minute,
			Limit:    apiBandwidthAbuseThreshold,
		}

		limitReached := rateLimiter.Test(context.Background(), key, rate)

		if limitReached {
			logging.WithFields(logging.Fields{
				"type": "security",
				"uri":  c.Request.URL.String(),
				"ip":   c.ClientIP(),
			}).Warning("API potentially abused by Client IP")
		}
	}
}
