package infra

import (
	"context"
	"go-auth-service/internal/config"
	"go-auth-service/pkg/logger"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, envs config.Env) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     envs.Redis.Host + ":" + envs.Redis.Port,
		Password: envs.Redis.Password,
	})

	if statusCmd := rdb.Ping(ctx); statusCmd.Err() != nil {
		logger.Fatalf("Failed to connect to Redis: %v", statusCmd.Err())
	}

	return rdb
}
