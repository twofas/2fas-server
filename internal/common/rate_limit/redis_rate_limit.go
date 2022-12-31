package rate_limit

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
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
	Client *redis.Client
}

func New(client *redis.Client) RateLimiter {
	return &RedisRateLimit{Client: client}
}

func (r *RedisRateLimit) Test(ctx context.Context, key string, rate Rate) bool {
	counter, err := r.Client.Get(context.Background(), key).Int()

	if err == redis.Nil {
		r.Client.Set(ctx, key, 1, rate.TimeUnit)

		return false
	}

	if err != nil {
		return false
	}

	if counter >= rate.Limit {
		r.Client.Del(context.Background(), key)

		return true
	} else {
		r.Client.Incr(context.Background(), key)
	}

	return false
}
