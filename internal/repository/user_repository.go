package repository

import (
	"database/sql"
	"go-auth-service/internal/domain/user"

	"github.com/redis/go-redis/v9"
)

type UserRepository struct {
	Db          *sql.DB
	RedisClient *redis.Client
}

func NewUserRepository(db *sql.DB, redisClient *redis.Client) *UserRepository {
	return &UserRepository{Db: db, RedisClient: redisClient}
}

// Compile-time check to ensure UserRepository implements user.RepositoryInterface
var _ user.RepositoryInterface = (*UserRepository)(nil)
