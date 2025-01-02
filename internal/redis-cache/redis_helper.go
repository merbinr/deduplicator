package rediscache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/merbinr/deduplicator/pkg/logger"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func LoadRedisClient() error {
	HOST_ENV := fmt.Sprintf("%s_DEDUPLICATOR_REDIS_CACHE_HOST", config.Config.StageName)
	host := os.Getenv(HOST_ENV)
	if host == "" {
		return fmt.Errorf("%s env variable not set", HOST_ENV)
	}

	address := fmt.Sprintf("%s:%d", host, config.Config.Services.RedisCache.Port)
	rdb = redis.NewClient(&redis.Options{
		Addr: address,
		// Password: password,
		DB: config.Config.Services.RedisCache.Db,
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}
	return nil
}

func SetValue(key string, value string) error {
	expiry := config.Config.Services.RedisCache.Expiry
	expiry_in_duration := time.Duration(expiry) * time.Second
	// expiry is already of type time.Duration
	err := rdb.Set(ctx, key, value, expiry_in_duration).Err()
	if err != nil {
		return fmt.Errorf("unable to set key: %s in the cache, error: %s", key, err)
	}
	return nil
}

func GetValue(key string) (string, error) {
	logger := logger.GetLogger()
	val, err := rdb.Get(ctx, key).Result()

	if err == redis.Nil {
		logger.Debug("Key does not exist in redis cache")
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("unable to get value for the key %s from cache, error: %s", key, err)
	}
	return val, nil
}
