package redis

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	DefaultPassword = ""
	DefaultDb       = 0
)

func New(host string, port int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: DefaultPassword,
		DB:       DefaultDb,
	})
}
