package infra

import (
	"database/sql"
	"fmt"
	"go-auth-service/internal/config"

	_ "github.com/lib/pq"
)

func NewPostgresDB(envs *config.Env) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		envs.Database.Host,
		envs.Database.Port,
		envs.Database.Username,
		envs.Database.Password,
		envs.Database.Name)

	db, err := sql.Open(envs.Database.Type, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify the connection is actually working
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
