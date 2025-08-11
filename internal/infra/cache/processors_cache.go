package cache

import (
	"context"

	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type ProcessorsCache struct {
	RedisClient *redis.Client
}

func NewProcessorsCache() *ProcessorsCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		DB:   0,
	})
	return &ProcessorsCache{RedisClient: rdb}
}

const (
	defaultProcessorFailingKey  = "processor:default:failing"
	fallbackProcessorFailingKey = "processor:fallback:failing"
)

func (c *ProcessorsCache) SetProcessorFailing(defaultProcessor bool, failing bool) error {
	key := ""
	if defaultProcessor {
		key = defaultProcessorFailingKey
	} else {
		key = fallbackProcessorFailingKey
	}
	val := "false"
	if failing {
		val = "true"
	}

	return c.RedisClient.Set(ctx, key, val, 0).Err()
}

func (c *ProcessorsCache) SetProcessorFailingTrue(defaultProcessor bool) error {
	key := ""
	if defaultProcessor {
		key = defaultProcessorFailingKey
	} else {
		key = fallbackProcessorFailingKey
	}
	return c.RedisClient.Set(ctx, key, "true", 1000*time.Millisecond).Err()
}

func (c *ProcessorsCache) GetProcessorFailing(defaultProcessor bool) (bool, error) {
	key := ""
	if defaultProcessor {
		key = defaultProcessorFailingKey
	} else {
		key = fallbackProcessorFailingKey
	}

	val, err := c.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return val == "true", nil
}
