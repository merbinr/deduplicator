package rediscache

import (
	"context"
	"fmt"
	"os"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func LoadRedisClient() error {
	address := fmt.Sprintf("%s:%d", config.Config.RedisCache.Host, config.Config.IncommingQueue.Port)
	password := os.Getenv("DEDUPLICATOR_REDIS_PASSWORD")
	if password == "" {
		return fmt.Errorf("DEDUPLICATOR_REDIS_PASSWORD env variable not set")
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       config.Config.RedisCache.DB,
	})
	return nil
}

func SetValue(key string, value string, expiry int) error {
	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("unable to set key: %s in the cache, error: %s", key, err)
	}
	return nil
}

func GetValue(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("unable to get value for the key %s from cache, error: %s", key, err)
	}
	return val, nil
}
