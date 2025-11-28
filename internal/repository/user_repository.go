package repository

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
)

type UserRepository struct {
	Db          *sql.DB
	RedisClient *redis.Client
}

func NewUserRepository(db *sql.DB, redisClient *redis.Client) *UserRepository {
	return &UserRepository{Db: db, RedisClient: redisClient}
}
