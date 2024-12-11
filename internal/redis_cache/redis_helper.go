package rediscache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func LoadRedisClient() error {
	host := os.Getenv("DEDUPLICATOR_REDIS_CACHE_HOST")
	if host == "" {
		return fmt.Errorf("DEDUPLICATOR_REDIS_CACHE_HOST env variable not set")
	}

	address := fmt.Sprintf("%s:%d", host, config.Config.RedisCache.Port)
	// password := os.Getenv("DEDUPLICATOR_REDIS_PASSWORD")
	// if password == "" {
	// 	return fmt.Errorf("DEDUPLICATOR_REDIS_PASSWORD env variable not set")
	// }
	rdb = redis.NewClient(&redis.Options{
		Addr: address,
		// Password: password,
		DB: config.Config.RedisCache.DB,
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}
	return nil
}

func SetValue(key string, value string) error {
	expiry := config.Config.RedisCache.Expiry
	err := rdb.Set(ctx, key, value, time.Duration(expiry)).Err()
	if err != nil {
		return fmt.Errorf("unable to set key: %s in the cache, error: %s", key, err)
	}
	return nil
}

func GetValue(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()

	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("unable to get value for the key %s from cache, error: %s", key, err)
	}
	return val, nil
}
