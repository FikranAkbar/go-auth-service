package repository

import (
	"context"
	"time"
)

// RefreshTokenData represents the data stored in Redis for refresh tokens
type RefreshTokenData struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TokenRepositoryInterface defines the contract for token storage operations
// This allows mocking in tests
type TokenRepositoryInterface interface {
	StoreRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, userID int64) (*RefreshTokenData, error)
	DeleteRefreshToken(ctx context.Context, userID int64) error
	ValidateRefreshToken(ctx context.Context, userID int64, token string) (bool, error)
}
