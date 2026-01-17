package infra

import (
	"context"
	"fmt"
	"go-auth-service/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, envs config.Env) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     envs.Redis.Host + ":" + envs.Redis.Port,
		Password: envs.Redis.Password,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return rdb, nil
}
