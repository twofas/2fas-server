package rate_limit

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/twofas/2fas-server/internal/common/logging"
)

type Rate struct {
	TimeUnit time.Duration
	Limit    int
}

type RateLimiter interface {
	Test(ctx context.Context, key string, rate Rate) bool
}

type LimitHandler func()

type RedisRateLimit struct {
	limiter *redis_rate.Limiter
}

func New(client *redis.Client) RateLimiter {
	return &RedisRateLimit{
		limiter: redis_rate.NewLimiter(client),
	}
}

// Test returns information if limit has been reached.
func (r *RedisRateLimit) Test(ctx context.Context, key string, rate Rate) bool {
	res, err := r.limiter.Allow(ctx, key, redis_rate.Limit{
		Rate:   rate.Limit,
		Burst:  rate.Limit,
		Period: rate.TimeUnit,
	})
	if err != nil {
		logging.WithFields(logging.Fields{
			"type": "security",
		}).Warnf("Could not check rate limit: %v", err)

		// for now we return that limit has not been reached.
		return false
	}
	if res.Allowed <= 0 {
		// limit has been reached.
		return true
	}
	return false
}
