package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"go-auth-service/pkg/constants"
	"go-auth-service/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

// RefreshTokenData represents the data stored in Redis for refresh tokens
type RefreshTokenData struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TokenRepository handles token storage in Redis
type TokenRepository struct {
	redis *redis.Client
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(redisClient *redis.Client) *TokenRepository {
	return &TokenRepository{
		redis: redisClient,
	}
}

// StoreRefreshToken stores a refresh token in Redis with TTL
func (r *TokenRepository) StoreRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	key := fmt.Sprintf("%s%d", constants.RedisKeyPrefixToken, userID)

	data := RefreshTokenData{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("Failed to marshal refresh token data for user %d: %v", userID, err)
		return err
	}

	// Calculate TTL based on expiration time
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}

	if err := r.redis.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		logger.Errorf("Failed to store refresh token in Redis for user %d: %v", userID, err)
		return err
	}

	logger.Infof("Stored refresh token for user %d with TTL %v", userID, ttl)
	return nil
}

// GetRefreshToken retrieves a refresh token from Redis
func (r *TokenRepository) GetRefreshToken(ctx context.Context, userID int64) (*RefreshTokenData, error) {
	key := fmt.Sprintf("%s%d", constants.RedisKeyPrefixToken, userID)

	val, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("refresh token not found for user %d", userID)
	}
	if err != nil {
		logger.Errorf("Failed to get refresh token from Redis for user %d: %v", userID, err)
		return nil, err
	}

	var data RefreshTokenData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		logger.Errorf("Failed to unmarshal refresh token data for user %d: %v", userID, err)
		return nil, err
	}

	return &data, nil
}

// DeleteRefreshToken removes a refresh token from Redis (used for logout)
func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("%s%d", constants.RedisKeyPrefixToken, userID)

	if err := r.redis.Del(ctx, key).Err(); err != nil {
		logger.Errorf("Failed to delete refresh token from Redis for user %d: %v", userID, err)
		return err
	}

	logger.Infof("Deleted refresh token for user %d", userID)
	return nil
}

// ValidateRefreshToken checks if the stored token matches the provided token
func (r *TokenRepository) ValidateRefreshToken(ctx context.Context, userID int64, token string) (bool, error) {
	data, err := r.GetRefreshToken(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if token matches
	if data.Token != token {
		logger.Warnf("Refresh token mismatch for user %d", userID)
		return false, nil
	}

	// Check if token is expired
	if time.Now().After(data.ExpiresAt) {
		logger.Warnf("Refresh token expired for user %d", userID)
		return false, nil
	}

	return true, nil
}
