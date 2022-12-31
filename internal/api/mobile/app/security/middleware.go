package security

import (
	"context"
	"fmt"
	"github.com/2fas/api/internal/common/logging"
	"github.com/2fas/api/internal/common/rate_limit"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

var mobileApiBandwidthAbuseThreshold = 100

func MobileIpAbuseAuditMiddleware(rateLimiter rate_limit.RateLimiter) gin.HandlerFunc {
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

		rate := rate_limit.Rate{
			TimeUnit: time.Minute,
			Limit:    mobileApiBandwidthAbuseThreshold,
		}

		limitReached := rateLimiter.Test(context.Background(), key, rate)

		if limitReached {
			logging.WithFields(logging.Fields{
				"type":                 "security",
				"uri":                  c.Request.URL.String(),
				"device_id":            deviceId,
				"browser_extension_id": extensionId,
				"ip":                   c.ClientIP(),
			}).Warning("API potentially abused at mobile scope")
		}
	}
}
