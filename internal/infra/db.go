package infra

import (
	"database/sql"
	"fmt"
	"go-auth-service/internal/config"
	"go-auth-service/pkg/logger"

	_ "github.com/lib/pq"
)

func NewPostgresDB(envs *config.Env) *sql.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		envs.Database.Host,
		envs.Database.Port,
		envs.Database.Username,
		envs.Database.Password,
		envs.Database.Name)

	db, err := sql.Open(envs.Database.Type, dsn)
	if err != nil {
		logger.Fatalf("failed to open database: %v", err)
	}

	return db
}
