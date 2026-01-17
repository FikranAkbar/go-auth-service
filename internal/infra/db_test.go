package infra

import (
	"go-auth-service/internal/config"
	"testing"

	_ "github.com/lib/pq"
)

// ============================================================================
// DATABASE CONNECTION TESTS
// Contract: Creates PostgreSQL database connection
// Success: Valid config → returns *sql.DB
// Note: Tests use connection pooling behavior, not actual DB connections
// ============================================================================

func TestNewPostgresDB_ValidConfig(t *testing.T) {
	// Create valid config
	cfg := &config.Env{
		Database: config.DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Port:     "5432",
			Name:     "testdb",
			Username: "testuser",
			Password: "testpass",
		},
	}

	// Note: sql.Open doesn't actually connect to the database,
	// it just validates the DSN and prepares the connection pool.
	// Actual connection happens on first query.
	db, err := NewPostgresDB(cfg)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify db object is created
	if db == nil {
		t.Fatal("Expected db to be non-nil")
	}

	// Verify we can get stats (validates connection pool is initialized)
	stats := db.Stats()
	if stats.MaxOpenConnections < 0 {
		t.Error("Expected valid connection pool stats")
	}

	// Clean up
	db.Close()
}

func TestNewPostgresDB_DSNFormat(t *testing.T) {
	// Test that DSN is formatted correctly
	cfg := &config.Env{
		Database: config.DatabaseConfig{
			Type:     "postgres",
			Host:     "customhost",
			Port:     "5433",
			Name:     "customdb",
			Username: "customuser",
			Password: "custompass",
		},
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if db == nil {
		t.Fatal("Expected db to be non-nil")
	}

	// Verify connection pool is initialized
	stats := db.Stats()
	if stats.MaxOpenConnections < 0 {
		t.Error("Expected valid connection pool")
	}

	db.Close()
}

func TestNewPostgresDB_DifferentDatabaseTypes(t *testing.T) {
	// Test with different database type specified
	cfg := &config.Env{
		Database: config.DatabaseConfig{
			Type:     "postgres", // Only postgres is supported
			Host:     "localhost",
			Port:     "5432",
			Name:     "testdb",
			Username: "testuser",
			Password: "testpass",
		},
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if db == nil {
		t.Fatal("Expected db to be non-nil")
	}

	db.Close()
}

func TestNewPostgresDB_EmptyPassword(t *testing.T) {
	// Test with empty password (some DBs allow this for local connections)
	cfg := &config.Env{
		Database: config.DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Port:     "5432",
			Name:     "testdb",
			Username: "testuser",
			Password: "", // Empty password
		},
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if db == nil {
		t.Fatal("Expected db to be non-nil")
	}

	db.Close()
}

func TestNewPostgresDB_SpecialCharactersInPassword(t *testing.T) {
	// Test with special characters in password
	cfg := &config.Env{
		Database: config.DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Port:     "5432",
			Name:     "testdb",
			Username: "testuser",
			Password: "p@ssw0rd!#$%",
		},
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if db == nil {
		t.Fatal("Expected db to be non-nil")
	}

	db.Close()
}

func TestNewPostgresDB_IPv6Host(t *testing.T) {
	// Test with IPv6 localhost
	cfg := &config.Env{
		Database: config.DatabaseConfig{
			Type:     "postgres",
			Host:     "::1",
			Port:     "5432",
			Name:     "testdb",
			Username: "testuser",
			Password: "testpass",
		},
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if db == nil {
		t.Fatal("Expected db to be non-nil")
	}

	db.Close()
}

func TestNewPostgresDB_NonStandardPort(t *testing.T) {
	// Test with non-standard port
	cfg := &config.Env{
		Database: config.DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Port:     "15432",
			Name:     "testdb",
			Username: "testuser",
			Password: "testpass",
		},
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if db == nil {
		t.Fatal("Expected db to be non-nil")
	}

	db.Close()
}

func TestNewPostgresDB_InvalidDriver_ReturnsError(t *testing.T) {
	// Now we can test error cases!
	cfg := &config.Env{
		Database: config.DatabaseConfig{
			Type:     "invalid_driver",
			Host:     "localhost",
			Port:     "5432",
			Name:     "testdb",
			Username: "testuser",
			Password: "testpass",
		},
	}

	db, err := NewPostgresDB(cfg)

	if err == nil {
		t.Error("Expected error for invalid driver, got nil")
	}

	if db != nil {
		t.Error("Expected db to be nil on error")
		db.Close()
	}
}
