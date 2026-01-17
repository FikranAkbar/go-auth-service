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

	err := initializeJWTConfig(cfg)
	if err != nil {
		t.Fatalf("Expected initializeJWTConfig to succeed, got error: %v", err)
	}

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

	err := initializeJWTConfig(cfg)
	if err != nil {
		t.Fatalf("Expected initializeJWTConfig to succeed, got error: %v", err)
	}

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

	err := initializeServerConfig(cfg)
	if err != nil {
		t.Fatalf("Expected initializeServerConfig to succeed, got error: %v", err)
	}

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

	err := initializeDatabaseConfig(cfg)
	if err != nil {
		t.Fatalf("Expected initializeDatabaseConfig to succeed, got error: %v", err)
	}

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

// ============================================================================
// ERROR CASE TESTS - To achieve 100% coverage
// ============================================================================

func TestLoadFromYAML_WithEnvVarSubstitution(t *testing.T) {
	// Set environment variable for substitution
	os.Setenv("TEST_JWT_SECRET", "env-substituted-secret-key-32-chars")
	defer os.Unsetenv("TEST_JWT_SECRET")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config with env var placeholder
	configContent := `
server:
  port: ":8080"
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
  secret_key: ${TEST_JWT_SECRET}
  access_token_expiry: 15m
  refresh_token_expiry: 168h
  issuer: test-service
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, err := loadFromYAML(configPath)
	if err != nil {
		t.Fatalf("Expected loadFromYAML to succeed, got error: %v", err)
	}

	if cfg.JWT.SecretKey != "env-substituted-secret-key-32-chars" {
		t.Errorf("Expected env var substitution, got %s", cfg.JWT.SecretKey)
	}
}

func TestInitializeServerConfig_InvalidDurations(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Env
		shouldErr bool
	}{
		{
			name: "Invalid ReadHeaderTimeout",
			cfg: &Env{
				Server: ServerConfig{
					ReadHeaderTimeoutDuration: "invalid",
					ReadTimeoutDuration:       "30s",
					WriteTimeoutDuration:      "30s",
					IdleTimeoutDuration:       "60s",
				},
			},
			shouldErr: true,
		},
		{
			name: "Invalid ReadTimeout",
			cfg: &Env{
				Server: ServerConfig{
					ReadHeaderTimeoutDuration: "10s",
					ReadTimeoutDuration:       "invalid",
					WriteTimeoutDuration:      "30s",
					IdleTimeoutDuration:       "60s",
				},
			},
			shouldErr: true,
		},
		{
			name: "Invalid WriteTimeout",
			cfg: &Env{
				Server: ServerConfig{
					ReadHeaderTimeoutDuration: "10s",
					ReadTimeoutDuration:       "30s",
					WriteTimeoutDuration:      "invalid",
					IdleTimeoutDuration:       "60s",
				},
			},
			shouldErr: true,
		},
		{
			name: "Invalid IdleTimeout",
			cfg: &Env{
				Server: ServerConfig{
					ReadHeaderTimeoutDuration: "10s",
					ReadTimeoutDuration:       "30s",
					WriteTimeoutDuration:      "30s",
					IdleTimeoutDuration:       "invalid",
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These tests would call logger.Fatalf in real scenario
			// We skip them as they would terminate the test
			t.Skip("Skipping fatal test - would terminate test suite")
			initializeServerConfig(tt.cfg)
		})
	}
}

func TestInitializeDatabaseConfig_InvalidDurations(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Env
	}{
		{
			name: "Invalid MaxConnectionLifetime",
			cfg: &Env{
				Database: DatabaseConfig{
					MaxConnectionLifetimeDuration:     "invalid",
					MaxIdleConnectionLifetimeDuration: "5m",
				},
			},
		},
		{
			name: "Invalid MaxIdleConnectionLifetime",
			cfg: &Env{
				Database: DatabaseConfig{
					MaxConnectionLifetimeDuration:     "30m",
					MaxIdleConnectionLifetimeDuration: "invalid",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These tests would call logger.Fatalf in real scenario
			t.Skip("Skipping fatal test - would terminate test suite")
			initializeDatabaseConfig(tt.cfg)
		})
	}
}

func TestInitializeJWTConfig_InvalidDurations(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Env
	}{
		{
			name: "Invalid AccessTokenExpiry",
			cfg: &Env{
				JWT: JWTConfig{
					SecretKey:             "valid-secret-key-at-least-32-characters-long",
					AccessTokenExpiryStr:  "invalid",
					RefreshTokenExpiryStr: "168h",
					Issuer:                "test",
				},
			},
		},
		{
			name: "Invalid RefreshTokenExpiry",
			cfg: &Env{
				JWT: JWTConfig{
					SecretKey:             "valid-secret-key-at-least-32-characters-long",
					AccessTokenExpiryStr:  "15m",
					RefreshTokenExpiryStr: "invalid",
					Issuer:                "test",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These tests would call logger.Fatalf in real scenario
			t.Skip("Skipping fatal test - would terminate test suite")
			initializeJWTConfig(tt.cfg)
		})
	}
}

func TestInitializeJWTConfig_ShortSecretKeyWarning(t *testing.T) {
	// This test is to cover the warning case for short secret key
	// It doesn't fail, just warns
	cfg := &Env{
		JWT: JWTConfig{
			SecretKey:             "short", // Less than 32 characters
			AccessTokenExpiryStr:  "15m",
			RefreshTokenExpiryStr: "168h",
			Issuer:                "test",
		},
	}

	// This should log a warning but not fail
	err := initializeJWTConfig(cfg)
	if err != nil {
		t.Fatalf("Expected initializeJWTConfig to succeed despite warning, got error: %v", err)
	}

	// Verify it still parses correctly
	if cfg.JWT.AccessTokenExpiry != 15*time.Minute {
		t.Errorf("Expected AccessTokenExpiry 15m despite warning, got %v", cfg.JWT.AccessTokenExpiry)
	}
}

func TestLoadFromEnv_AllOptionalDefaults(t *testing.T) {
	// Set only required env vars, check if defaults are applied
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USERNAME", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("REDIS_HOST", "redishost")
	os.Setenv("REDIS_PASSWORD", "redispass")
	os.Setenv("JWT_SECRET_KEY", "test-secret-key-at-least-32-characters-long")
	os.Setenv("JWT_ISSUER", "test-issuer")

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

	cfg, err := loadFromEnv()
	if err != nil {
		t.Fatalf("Expected loadFromEnv to succeed, got error: %v", err)
	}

	// Check defaults
	if cfg.Database.Type != "postgres" {
		t.Errorf("Expected default DB type postgres, got %s", cfg.Database.Type)
	}
	if cfg.Database.Port != "5432" {
		t.Errorf("Expected default DB port 5432, got %s", cfg.Database.Port)
	}
	if cfg.Redis.Port != "6379" {
		t.Errorf("Expected default Redis port 6379, got %s", cfg.Redis.Port)
	}
	if cfg.Server.Port != ":8080" {
		t.Errorf("Expected default server port :8080, got %s", cfg.Server.Port)
	}
	if cfg.JWT.AccessTokenExpiryStr != "15m" {
		t.Errorf("Expected default access token expiry 15m, got %s", cfg.JWT.AccessTokenExpiryStr)
	}
	if cfg.JWT.RefreshTokenExpiryStr != "7d" {
		t.Errorf("Expected default refresh token expiry 7d, got %s", cfg.JWT.RefreshTokenExpiryStr)
	}
	if cfg.JWT.BcryptCost != 4 {
		t.Errorf("Expected default bcrypt cost 4, got %d", cfg.JWT.BcryptCost)
	}
	if cfg.Email.SMTPPort != "587" {
		t.Errorf("Expected default SMTP port 587, got %s", cfg.Email.SMTPPort)
	}
	if cfg.Email.FromEmail != "noreply@example.com" {
		t.Errorf("Expected default from email, got %s", cfg.Email.FromEmail)
	}
	if cfg.Email.FromName != "Go Auth Service" {
		t.Errorf("Expected default from name, got %s", cfg.Email.FromName)
	}
	if cfg.App.URL != "http://localhost:8080" {
		t.Errorf("Expected default app URL, got %s", cfg.App.URL)
	}
}

func TestLoadFromEnv_WithCustomOptionalValues(t *testing.T) {
	// Set all env vars including optional ones
	os.Setenv("DB_TYPE", "mysql")
	os.Setenv("DB_HOST", "customhost")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_NAME", "customdb")
	os.Setenv("DB_USERNAME", "customuser")
	os.Setenv("DB_PASSWORD", "custompass")
	os.Setenv("REDIS_HOST", "customredis")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_PASSWORD", "customredispass")
	os.Setenv("SERVER_ADDR", ":9000")
	os.Setenv("JWT_SECRET_KEY", "custom-secret-key-at-least-32-characters-long")
	os.Setenv("JWT_ACCESS_TOKEN_EXPIRY", "30m")
	os.Setenv("JWT_REFRESH_TOKEN_EXPIRY", "14d")
	os.Setenv("JWT_ISSUER", "custom-issuer")
	os.Setenv("SMTP_HOST", "smtp.custom.com")
	os.Setenv("SMTP_PORT", "465")
	os.Setenv("SMTP_USERNAME", "customsmtp")
	os.Setenv("SMTP_PASSWORD", "customsmtppass")
	os.Setenv("FROM_EMAIL", "custom@example.com")
	os.Setenv("FROM_NAME", "Custom Service")
	os.Setenv("APP_URL", "https://custom.example.com")

	defer func() {
		os.Unsetenv("DB_TYPE")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_USERNAME")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("SERVER_ADDR")
		os.Unsetenv("JWT_SECRET_KEY")
		os.Unsetenv("JWT_ACCESS_TOKEN_EXPIRY")
		os.Unsetenv("JWT_REFRESH_TOKEN_EXPIRY")
		os.Unsetenv("JWT_ISSUER")
		os.Unsetenv("SMTP_HOST")
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("SMTP_USERNAME")
		os.Unsetenv("SMTP_PASSWORD")
		os.Unsetenv("FROM_EMAIL")
		os.Unsetenv("FROM_NAME")
		os.Unsetenv("APP_URL")
	}()

	cfg, err := loadFromEnv()
	if err != nil {
		t.Fatalf("Expected loadFromEnv to succeed, got error: %v", err)
	}

	// Verify custom values
	if cfg.Database.Type != "mysql" {
		t.Errorf("Expected custom DB type mysql, got %s", cfg.Database.Type)
	}
	if cfg.Database.Port != "3306" {
		t.Errorf("Expected custom DB port 3306, got %s", cfg.Database.Port)
	}
	if cfg.Redis.Port != "6380" {
		t.Errorf("Expected custom Redis port 6380, got %s", cfg.Redis.Port)
	}
	if cfg.Server.Port != ":9000" {
		t.Errorf("Expected custom server port :9000, got %s", cfg.Server.Port)
	}
	if cfg.JWT.AccessTokenExpiryStr != "30m" {
		t.Errorf("Expected custom access token expiry 30m, got %s", cfg.JWT.AccessTokenExpiryStr)
	}
	if cfg.Email.SMTPHost != "smtp.custom.com" {
		t.Errorf("Expected custom SMTP host, got %s", cfg.Email.SMTPHost)
	}
	if cfg.App.URL != "https://custom.example.com" {
		t.Errorf("Expected custom app URL, got %s", cfg.App.URL)
	}
}
func TestLoadFromEnv_MissingRequiredVars(t *testing.T) {
	// Clear all env vars to test missing required vars
	envVars := []string{
		"DB_HOST", "DB_NAME", "DB_USERNAME", "DB_PASSWORD",
		"REDIS_HOST", "REDIS_PASSWORD", "JWT_SECRET_KEY", "JWT_ISSUER",
	}
	for _, v := range envVars {
		_ = os.Unsetenv(v)
	}
	cfg, err := loadFromEnv()
	if err == nil {
		t.Fatal("Expected error for missing required env vars, got nil")
	}
	if cfg != nil {
		t.Error("Expected nil config when env vars are missing")
	}
	if err.Error() == "" {
		t.Error("Expected error message about missing env vars")
	}
}
func TestLoadServiceEnvironmentVariables_MissingEnvVars(t *testing.T) {
	// Clear all env vars
	envVars := []string{
		"DB_HOST", "DB_NAME", "DB_USERNAME", "DB_PASSWORD",
		"REDIS_HOST", "REDIS_PASSWORD", "JWT_SECRET_KEY", "JWT_ISSUER",
	}
	for _, v := range envVars {
		_ = os.Unsetenv(v)
	}
	// Change to temp directory (no config.yaml)
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)
	// Should return nil due to missing env vars
	cfg := LoadServiceEnvironmentVariables()
	if cfg != nil {
		t.Error("Expected nil config when required env vars are missing")
	}
}
func TestLoadServiceEnvironmentVariables_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	// Invalid YAML content
	invalidYAML := `
server:
  port: ":8080"
  invalid yaml syntax here: [
`
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)
	cfg := LoadServiceEnvironmentVariables()
	if cfg != nil {
		t.Error("Expected nil config for invalid YAML")
	}
}
func TestLoadServiceEnvironmentVariables_InvalidServerConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
server:
  port: ":8080"
  read_header_timeout_duration: "invalid"
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
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)
	cfg := LoadServiceEnvironmentVariables()
	if cfg != nil {
		t.Error("Expected nil config for invalid server timeout duration")
	}
}
func TestLoadServiceEnvironmentVariables_InvalidDatabaseConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
server:
  port: ":8080"
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
  max_idle_connection_lifetime: "invalid"
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
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)
	cfg := LoadServiceEnvironmentVariables()
	if cfg != nil {
		t.Error("Expected nil config for invalid database connection lifetime")
	}
}
func TestLoadServiceEnvironmentVariables_InvalidJWTConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
server:
  port: ":8080"
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
  secret_key: ""
  access_token_expiry: 15m
  refresh_token_expiry: 168h
  issuer: test-service
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)
	cfg := LoadServiceEnvironmentVariables()
	if cfg != nil {
		t.Error("Expected nil config for empty JWT secret key")
	}
}
func TestLoadFromYAML_FileNotFound(t *testing.T) {
	cfg, err := loadFromYAML("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
	if cfg != nil {
		t.Error("Expected nil config for non-existent file")
	}
}
func TestLoadFromYAML_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	// Invalid YAML syntax
	invalidYAML := `
server:
  port: ":8080"
  invalid: [unclosed bracket
`
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	cfg, err := loadFromYAML(configPath)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
	if cfg != nil {
		t.Error("Expected nil config for invalid YAML")
	}
}
func TestInitializeServerConfig_AllInvalidDurations(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Env
		expectErr bool
		errMsg    string
	}{
		{
			name: "Invalid ReadHeaderTimeout",
			cfg: &Env{
				Server: ServerConfig{
					ReadHeaderTimeoutDuration: "invalid",
					ReadTimeoutDuration:       "30s",
					WriteTimeoutDuration:      "30s",
					IdleTimeoutDuration:       "60s",
				},
			},
			expectErr: true,
			errMsg:    "read header timeout",
		},
		{
			name: "Invalid ReadTimeout",
			cfg: &Env{
				Server: ServerConfig{
					ReadHeaderTimeoutDuration: "10s",
					ReadTimeoutDuration:       "not-a-duration",
					WriteTimeoutDuration:      "30s",
					IdleTimeoutDuration:       "60s",
				},
			},
			expectErr: true,
			errMsg:    "read timeout",
		},
		{
			name: "Invalid WriteTimeout",
			cfg: &Env{
				Server: ServerConfig{
					ReadHeaderTimeoutDuration: "10s",
					ReadTimeoutDuration:       "30s",
					WriteTimeoutDuration:      "xyz",
					IdleTimeoutDuration:       "60s",
				},
			},
			expectErr: true,
			errMsg:    "write timeout",
		},
		{
			name: "Invalid IdleTimeout",
			cfg: &Env{
				Server: ServerConfig{
					ReadHeaderTimeoutDuration: "10s",
					ReadTimeoutDuration:       "30s",
					WriteTimeoutDuration:      "30s",
					IdleTimeoutDuration:       "bad-duration",
				},
			},
			expectErr: true,
			errMsg:    "idle timeout",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := initializeServerConfig(tt.cfg)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
func TestInitializeDatabaseConfig_AllInvalidDurations(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Env
		expectErr bool
		errMsg    string
	}{
		{
			name: "Invalid MaxConnectionLifetime",
			cfg: &Env{
				Database: DatabaseConfig{
					MaxConnectionLifetimeDuration:     "not-valid",
					MaxIdleConnectionLifetimeDuration: "5m",
				},
			},
			expectErr: true,
			errMsg:    "max connection lifetime",
		},
		{
			name: "Invalid MaxIdleConnectionLifetime",
			cfg: &Env{
				Database: DatabaseConfig{
					MaxConnectionLifetimeDuration:     "30m",
					MaxIdleConnectionLifetimeDuration: "bad",
				},
			},
			expectErr: true,
			errMsg:    "max idle connection lifetime",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := initializeDatabaseConfig(tt.cfg)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
func TestInitializeJWTConfig_AllInvalidCases(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Env
		expectErr bool
		errMsg    string
	}{
		{
			name: "Empty SecretKey",
			cfg: &Env{
				JWT: JWTConfig{
					SecretKey:             "",
					AccessTokenExpiryStr:  "15m",
					RefreshTokenExpiryStr: "168h",
				},
			},
			expectErr: true,
			errMsg:    "secret key is required",
		},
		{
			name: "Invalid AccessTokenExpiry",
			cfg: &Env{
				JWT: JWTConfig{
					SecretKey:             "valid-secret-key-at-least-32-characters-long",
					AccessTokenExpiryStr:  "not-a-duration",
					RefreshTokenExpiryStr: "168h",
				},
			},
			expectErr: true,
			errMsg:    "access token expiry",
		},
		{
			name: "Invalid RefreshTokenExpiry",
			cfg: &Env{
				JWT: JWTConfig{
					SecretKey:             "valid-secret-key-at-least-32-characters-long",
					AccessTokenExpiryStr:  "15m",
					RefreshTokenExpiryStr: "invalid",
				},
			},
			expectErr: true,
			errMsg:    "refresh token expiry",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := initializeJWTConfig(tt.cfg)
			if tt.expectErr && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
