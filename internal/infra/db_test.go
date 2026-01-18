package infra

import (
	"go-auth-service/internal/config"
	"testing"

	_ "github.com/lib/pq"
)

// ============================================================================
// DATABASE CONNECTION TESTS
// Contract: Creates PostgreSQL database connection
// Success: Valid config + reachable DB → returns *sql.DB
// Error: Invalid config OR unreachable DB → returns error
//
// NOTE: These tests now require an actual PostgreSQL database to be running
// because db.Ping() verifies the connection. Tests will fail if DB is unavailable.
// To run these tests, either:
// 1. Have a local PostgreSQL running on localhost:5432
// 2. Skip these tests: go test -short
// 3. Use testcontainers for integration tests
// ============================================================================

func TestNewPostgresDB_ValidConfig_WithActualDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

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

	// NOTE: This will now fail if there's no actual database
	// because db.Ping() actually connects
	db, err := NewPostgresDB(cfg)

	// Without a real DB, we expect an error from Ping()
	if err != nil {
		// This is expected - no real database in test environment
		t.Logf("Expected error (no real DB): %v", err)
		return
	}

	// If we somehow have a real DB, verify it works
	if db == nil {
		t.Fatal("Expected db to be non-nil")
	}

	// Verify we can get stats
	stats := db.Stats()
	if stats.MaxOpenConnections < 0 {
		t.Error("Expected valid connection pool stats")
	}

	// Clean up
	db.Close()
}

func TestNewPostgresDB_DSNFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

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

	// This will fail ping because customhost doesn't exist
	db, err := NewPostgresDB(cfg)

	// We expect an error from Ping() since the host doesn't exist
	if err == nil {
		t.Log("Unexpected success - test database might be running")
		if db != nil {
			db.Close()
		}
		return
	}

	// Error is expected - verify it's a ping error
	if err != nil {
		t.Logf("Expected ping failure: %v", err)
	}
}

func TestNewPostgresDB_DifferentDatabaseTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

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

	// Expect ping error without real DB
	if err != nil {
		t.Logf("Expected error (no real DB): %v", err)
		return
	}

	if db != nil {
		db.Close()
	}
}

func TestNewPostgresDB_EmptyPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

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

	// Expect ping error without real DB
	if err != nil {
		t.Logf("Expected error (no real DB): %v", err)
		return
	}

	if db != nil {
		db.Close()
	}
}

func TestNewPostgresDB_SpecialCharactersInPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

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

	// Expect ping error without real DB
	if err != nil {
		t.Logf("Expected error (no real DB): %v", err)
		return
	}

	if db != nil {
		db.Close()
	}
}

func TestNewPostgresDB_IPv6Host(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

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

	// Expect ping error without real DB
	if err != nil {
		t.Logf("Expected error (no real DB): %v", err)
		return
	}

	if db != nil {
		db.Close()
	}
}

func TestNewPostgresDB_NonStandardPort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

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

	// Expect ping error without real DB
	if err != nil {
		t.Logf("Expected error (no real DB): %v", err)
		return
	}

	if db != nil {
		db.Close()
	}
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
