package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/logger"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func init() {
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     config.GetRedisAddr(),
		Password: config.GlobalConfig.Redis.Password,
		DB:       0,
	})

	// Bound startup probing and close the client if the probe fails.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		logger.Error(fmt.Sprintf("failed to connect to Redis: %v", err))
		return
	}

	Client = client
	logger.Info("Redis connected successfully")
	return
}

func Close() error {
	if Client != nil {
		if err := Client.Close(); err != nil {
			return fmt.Errorf("failed to close Redis connection: %v", err)
		}
	}
	return nil
}

func SendPluginLogToRedis(key string, msg string) error {
	_, err := Client.RPush(context.Background(), key, msg).Result()
	if err != nil {
		return fmt.Errorf("failed to add values to list: %v", err)
	}
	return nil
}
