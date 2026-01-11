package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadServiceEnvironmentVariables_WithValidYAML(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  port: ":9000"
  read_header_timeout_duration: "10s"
  read_timeout_duration: "30s"
  write_timeout_duration: "30s"
  idle_timeout_duration: "60s"

database:
  type: postgres
  host: localhost
  port: "5432"
  name: testdb
  username: testuser
  password: testpass
  max_idle_connections: "10"
  max_open_connections: "100"
  max_idle_connection_lifetime: "5m"
  max_connection_lifetime: "30m"

redis:
  host: localhost
  port: "6379"
  password: ""

jwt:
  secret_key: test-secret-key-at-least-32-characters-long
  access_token_expiry: 15m
  refresh_token_expiry: 168h
  issuer: test-service
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Load config
	cfg := LoadServiceEnvironmentVariables()

	// Verify Server Config
	if cfg.Server.Port != ":9000" {
		t.Errorf("Expected server port :9000, got %s", cfg.Server.Port)
	}
	if cfg.Server.ReadHeaderTimeout != 10*time.Second {
		t.Errorf("Expected ReadHeaderTimeout 10s, got %v", cfg.Server.ReadHeaderTimeout)
	}

	// Verify Database Config
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected database host localhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.Name != "testdb" {
		t.Errorf("Expected database name testdb, got %s", cfg.Database.Name)
	}
	if cfg.Database.MaxConnectionLifetime != 30*time.Minute {
		t.Errorf("Expected MaxConnectionLifetime 30m, got %v", cfg.Database.MaxConnectionLifetime)
	}

	// Verify Redis Config
	if cfg.Redis.Host != "localhost" {
		t.Errorf("Expected redis host localhost, got %s", cfg.Redis.Host)
	}
	if cfg.Redis.Port != "6379" {
		t.Errorf("Expected redis port 6379, got %s", cfg.Redis.Port)
	}

	// Verify JWT Config
	if cfg.JWT.SecretKey != "test-secret-key-at-least-32-characters-long" {
		t.Errorf("Expected JWT secret key, got %s", cfg.JWT.SecretKey)
	}
	if cfg.JWT.AccessTokenExpiry != 15*time.Minute {
		t.Errorf("Expected AccessTokenExpiry 15m, got %v", cfg.JWT.AccessTokenExpiry)
	}
	if cfg.JWT.RefreshTokenExpiry != 168*time.Hour {
		t.Errorf("Expected RefreshTokenExpiry 168h, got %v", cfg.JWT.RefreshTokenExpiry)
	}
	if cfg.JWT.Issuer != "test-service" {
		t.Errorf("Expected JWT issuer test-service, got %s", cfg.JWT.Issuer)
	}
}

func TestLoadServiceEnvironmentVariables_MissingConfigFile_LoadsFromEnv(t *testing.T) {
	// Set required environment variables
	os.Setenv("DB_HOST", "envhost")
	os.Setenv("DB_NAME", "envdb")
	os.Setenv("DB_USERNAME", "envuser")
	os.Setenv("DB_PASSWORD", "envpass")
	os.Setenv("REDIS_HOST", "redishost")
	os.Setenv("REDIS_PASSWORD", "redispass")
	os.Setenv("JWT_SECRET_KEY", "env-jwt-secret-key-at-least-32-characters-long")
	os.Setenv("JWT_ISSUER", "env-issuer")

	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_USERNAME")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("JWT_SECRET_KEY")
		os.Unsetenv("JWT_ISSUER")
	}()

	// Change to temp directory (no config.yaml)
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Load config (should fallback to env vars)
	cfg := LoadServiceEnvironmentVariables()

	// Verify values from environment
	if cfg.Database.Host != "envhost" {
		t.Errorf("Expected DB host from env, got %s", cfg.Database.Host)
	}
	if cfg.Redis.Host != "redishost" {
		t.Errorf("Expected Redis host from env, got %s", cfg.Redis.Host)
	}
	if cfg.JWT.SecretKey != "env-jwt-secret-key-at-least-32-characters-long" {
		t.Errorf("Expected JWT secret from env, got %s", cfg.JWT.SecretKey)
	}
}

func TestInitializeJWTConfig_EmptySecretKey_ShouldFail(t *testing.T) {
	// This test expects logger.Fatal to be called
	// In real scenario, you might want to use a test logger or mock
	// For now, we skip this test as it would terminate the test
	t.Skip("Skipping fatal test - would terminate test suite")

	cfg := &Env{
		JWT: JWTConfig{
			SecretKey:             "",
			AccessTokenExpiryStr:  "15m",
			RefreshTokenExpiryStr: "168h",
		},
	}

	// This would call logger.Fatal
	initializeJWTConfig(cfg)
}

func TestInitializeJWTConfig_ValidConfig(t *testing.T) {
	cfg := &Env{
		JWT: JWTConfig{
			SecretKey:             "valid-secret-key-at-least-32-characters-long",
			AccessTokenExpiryStr:  "15m",
			RefreshTokenExpiryStr: "168h",
			Issuer:                "test-issuer",
		},
	}

	initializeJWTConfig(cfg)

	if cfg.JWT.AccessTokenExpiry != 15*time.Minute {
		t.Errorf("Expected AccessTokenExpiry 15m, got %v", cfg.JWT.AccessTokenExpiry)
	}
	if cfg.JWT.RefreshTokenExpiry != 168*time.Hour {
		t.Errorf("Expected RefreshTokenExpiry 168h, got %v", cfg.JWT.RefreshTokenExpiry)
	}
	if cfg.JWT.Issuer != "test-issuer" {
		t.Errorf("Expected issuer test-issuer, got %s", cfg.JWT.Issuer)
	}
}

func TestInitializeJWTConfig_DefaultIssuer(t *testing.T) {
	cfg := &Env{
		JWT: JWTConfig{
			SecretKey:             "valid-secret-key-at-least-32-characters-long",
			AccessTokenExpiryStr:  "15m",
			RefreshTokenExpiryStr: "168h",
			Issuer:                "", // Empty issuer
		},
	}

	initializeJWTConfig(cfg)

	if cfg.JWT.Issuer != "go-auth-service" {
		t.Errorf("Expected default issuer go-auth-service, got %s", cfg.JWT.Issuer)
	}
}

func TestInitializeServerConfig(t *testing.T) {
	cfg := &Env{
		Server: ServerConfig{
			ReadHeaderTimeoutDuration: "10s",
			ReadTimeoutDuration:       "30s",
			WriteTimeoutDuration:      "30s",
			IdleTimeoutDuration:       "60s",
		},
	}

	initializeServerConfig(cfg)

	if cfg.Server.ReadHeaderTimeout != 10*time.Second {
		t.Errorf("Expected ReadHeaderTimeout 10s, got %v", cfg.Server.ReadHeaderTimeout)
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Expected ReadTimeout 30s, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 30*time.Second {
		t.Errorf("Expected WriteTimeout 30s, got %v", cfg.Server.WriteTimeout)
	}
	if cfg.Server.IdleTimeout != 60*time.Second {
		t.Errorf("Expected IdleTimeout 60s, got %v", cfg.Server.IdleTimeout)
	}
}

func TestInitializeDatabaseConfig(t *testing.T) {
	cfg := &Env{
		Database: DatabaseConfig{
			MaxConnectionLifetimeDuration:     "30m",
			MaxIdleConnectionLifetimeDuration: "5m",
		},
	}

	initializeDatabaseConfig(cfg)

	if cfg.Database.MaxConnectionLifetime != 30*time.Minute {
		t.Errorf("Expected MaxConnectionLifetime 30m, got %v", cfg.Database.MaxConnectionLifetime)
	}
	if cfg.Database.MaxIdleConnectionLifetime != 5*time.Minute {
		t.Errorf("Expected MaxIdleConnectionLifetime 5m, got %v", cfg.Database.MaxIdleConnectionLifetime)
	}
}

func TestFileExists(t *testing.T) {
	// Test existing file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	if !fileExists(tmpFile) {
		t.Error("Expected file to exist")
	}

	// Test non-existing file
	if fileExists(filepath.Join(tmpDir, "nonexistent.txt")) {
		t.Error("Expected file to not exist")
	}

	// Test directory (should return false)
	if fileExists(tmpDir) {
		t.Error("Expected directory to return false")
	}
}

func TestEnvLoader_Required(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	loader := &envLoader{}
	value := loader.req("TEST_KEY")

	if value != "test_value" {
		t.Errorf("Expected test_value, got %s", value)
	}
	if len(loader.errors) != 0 {
		t.Errorf("Expected no errors, got %v", loader.errors)
	}
}

func TestEnvLoader_Required_Missing(t *testing.T) {
	loader := &envLoader{}
	value := loader.req("MISSING_KEY")

	if value != "" {
		t.Errorf("Expected empty string, got %s", value)
	}
	if len(loader.errors) != 1 {
		t.Errorf("Expected 1 error, got %v", loader.errors)
	}
	if loader.errors[0] != "MISSING_KEY" {
		t.Errorf("Expected MISSING_KEY in errors, got %v", loader.errors)
	}
}

func TestEnvLoader_Optional(t *testing.T) {
	os.Setenv("TEST_OPT", "opt_value")
	defer os.Unsetenv("TEST_OPT")

	loader := &envLoader{}
	value := loader.opt("TEST_OPT", "default")

	if value != "opt_value" {
		t.Errorf("Expected opt_value, got %s", value)
	}
}

func TestEnvLoader_Optional_Default(t *testing.T) {
	loader := &envLoader{}
	value := loader.opt("MISSING_OPT", "default_value")

	if value != "default_value" {
		t.Errorf("Expected default_value, got %s", value)
	}
}
