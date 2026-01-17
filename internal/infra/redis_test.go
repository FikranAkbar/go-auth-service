package infra

import (
	"context"
	"go-auth-service/internal/config"
	"testing"
	"time"
)

// ============================================================================
// REDIS CLIENT TESTS
// Contract: Creates Redis client connection
// Success: Valid config → returns *redis.Client
// Note: Tests validate client creation, not actual Redis connection
// ============================================================================

func TestNewRedisClient_ValidConfig(t *testing.T) {
	// This test would require a running Redis instance
	// We skip it in unit tests and would run it in integration tests
	t.Skip("Skipping integration test - requires running Redis instance")

	ctx := context.Background()
	cfg := config.Env{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
		},
	}

	client, err := NewRedisClient(ctx, cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected Redis client to be non-nil")
	}

	// Clean up
	client.Close()
}

func TestNewRedisClient_WithPassword(t *testing.T) {
	// This test documents Redis client creation with password
	t.Skip("Skipping integration test - requires running Redis instance with password")

	ctx := context.Background()
	cfg := config.Env{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "redis_password",
		},
	}

	client, err := NewRedisClient(ctx, cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected Redis client to be non-nil")
	}

	client.Close()
}

func TestNewRedisClient_CustomPort(t *testing.T) {
	// This test documents Redis client creation with custom port
	t.Skip("Skipping integration test - requires running Redis instance")

	ctx := context.Background()
	cfg := config.Env{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6380",
			Password: "",
		},
	}

	client, err := NewRedisClient(ctx, cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected Redis client to be non-nil")
	}

	client.Close()
}

func TestNewRedisClient_WithTimeout(t *testing.T) {
	// This test documents Redis client creation with context timeout
	t.Skip("Skipping integration test - requires running Redis instance")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := config.Env{
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
		},
	}

	client, err := NewRedisClient(ctx, cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected Redis client to be non-nil")
	}

	client.Close()
}

func TestNewRedisClient_ConnectionFailure_ReturnsError(t *testing.T) {
	// Now we can test error cases without Fatal!
	ctx := context.Background()
	cfg := config.Env{
		Redis: config.RedisConfig{
			Host:     "invalid_host_that_does_not_exist",
			Port:     "6379",
			Password: "",
		},
	}

	client, err := NewRedisClient(ctx, cfg)

	if err == nil {
		t.Error("Expected error for invalid Redis host, got nil")
	}

	if client != nil {
		t.Error("Expected client to be nil on error")
		client.Close()
	}
}

// Note: Redis client tests are typically integration tests
// Unit tests would require mocking the Redis connection
// For production, consider:
// 1. Separate integration tests that require actual Redis
// 2. Mock Redis interface for unit tests
// 3. Use testcontainers for isolated Redis instances in CI/CD
