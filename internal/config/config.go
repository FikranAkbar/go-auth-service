package config

import (
	"fmt"
	"go-auth-service/pkg/logger"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Env struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Email    EmailConfig    `yaml:"email"`
	App      AppConfig      `yaml:"app"`
}

type ServerConfig struct {
	Port              string        `yaml:"port"`
	ReadHeaderTimeout time.Duration `yaml:"-"`
	ReadTimeout       time.Duration `yaml:"-"`
	WriteTimeout      time.Duration `yaml:"-"`
	IdleTimeout       time.Duration `yaml:"-"`

	ReadHeaderTimeoutDuration string `yaml:"read_header_timeout_duration"`
	ReadTimeoutDuration       string `yaml:"read_timeout_duration"`
	WriteTimeoutDuration      string `yaml:"write_timeout_duration"`
	IdleTimeoutDuration       string `yaml:"idle_timeout_duration"`
}

type DatabaseConfig struct {
	Type                              string        `yaml:"type"`
	Host                              string        `yaml:"host"`
	Port                              string        `yaml:"port"`
	Name                              string        `yaml:"name"`
	Username                          string        `yaml:"username"`
	Password                          string        `yaml:"password"`
	MaxIdleConnections                string        `yaml:"max_idle_connections"`
	MaxOpenConnections                string        `yaml:"max_open_connections"`
	MaxIdleConnectionLifetimeDuration string        `yaml:"max_idle_connection_lifetime"`
	MaxIdleConnectionLifetime         time.Duration `yaml:"-"`
	MaxConnectionLifetimeDuration     string        `yaml:"max_connection_lifetime"`
	MaxConnectionLifetime             time.Duration `yaml:"-"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
}

type JWTConfig struct {
	SecretKey             string        `yaml:"secret_key"`
	AccessTokenExpiry     time.Duration `yaml:"-"`
	RefreshTokenExpiry    time.Duration `yaml:"-"`
	AccessTokenExpiryStr  string        `yaml:"access_token_expiry"`
	RefreshTokenExpiryStr string        `yaml:"refresh_token_expiry"`
	Issuer                string        `yaml:"issuer"`
	BcryptCost            int           `yaml:"bcrypt_cost"` // Password hashing cost (4-31, recommended: 4 for dev, 10 for prod)
}

type EmailConfig struct {
	SMTPHost     string `yaml:"smtp_host"`
	SMTPPort     string `yaml:"smtp_port"`
	SMTPUsername string `yaml:"smtp_username"`
	SMTPPassword string `yaml:"smtp_password"`
	FromEmail    string `yaml:"from_email"`
	FromName     string `yaml:"from_name"`
}

type AppConfig struct {
	URL string `yaml:"url"`
}

type envLoader struct {
	errors []string
}

func (l *envLoader) req(key string) string {
	v := os.Getenv(key)
	if v == "" {
		l.errors = append(l.errors, key)
	}
	return v
}

func (l *envLoader) opt(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func loadFromEnv() (*Env, error) {
	l := &envLoader{}

	cfg := &Env{
		Database: DatabaseConfig{
			Type:     l.opt("DB_TYPE", "postgres"),
			Host:     l.req("DB_HOST"),
			Port:     l.opt("DB_PORT", "5432"),
			Name:     l.req("DB_NAME"),
			Username: l.req("DB_USERNAME"),
			Password: l.req("DB_PASSWORD"),
		},
		Redis: RedisConfig{
			Host:     l.req("REDIS_HOST"),
			Port:     l.opt("REDIS_PORT", "6379"),
			Password: l.req("REDIS_PASSWORD"),
		},
		Server: ServerConfig{
			Port: l.opt("SERVER_ADDR", ":8080"),
		},
		JWT: JWTConfig{
			SecretKey:             l.req("JWT_SECRET_KEY"),
			AccessTokenExpiryStr:  l.opt("JWT_ACCESS_TOKEN_EXPIRY", "15m"),
			RefreshTokenExpiryStr: l.opt("JWT_REFRESH_TOKEN_EXPIRY", "7d"),
			Issuer:                l.req("JWT_ISSUER"),
			BcryptCost:            4, // Default to 4 for fast development
		},
		Email: EmailConfig{
			SMTPHost:     l.opt("SMTP_HOST", ""),
			SMTPPort:     l.opt("SMTP_PORT", "587"),
			SMTPUsername: l.opt("SMTP_USERNAME", ""),
			SMTPPassword: l.opt("SMTP_PASSWORD", ""),
			FromEmail:    l.opt("FROM_EMAIL", "noreply@example.com"),
			FromName:     l.opt("FROM_NAME", "Go Auth Service"),
		},
		App: AppConfig{
			URL: l.opt("APP_URL", "http://localhost:8080"),
		},
	}

	if len(l.errors) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(l.errors, ", "))
	}

	return cfg, nil
}

func LoadServiceEnvironmentVariables() *Env {
	const (
		configPath = `config.yaml`
	)

	if !fileExists(configPath) {
		cfg, err := loadFromEnv()
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		return cfg
	}

	cfg, err := loadFromYAML(configPath)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	if err := initializeServerConfig(cfg); err != nil {
		logger.Error(err.Error())
		return nil
	}

	if err := initializeDatabaseConfig(cfg); err != nil {
		logger.Error(err.Error())
		return nil
	}

	if err := initializeJWTConfig(cfg); err != nil {
		logger.Error(err.Error())
		return nil
	}

	return cfg
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func loadFromYAML(path string) (*Env, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error loading config file %s: %s", path, err)
	}

	// Expand environment variables if present (e.g., ${JWT_SECRET_KEY})
	// This allows flexibility: use env vars OR hardcode values in config.yaml
	dataStr := os.ExpandEnv(string(data))

	var cfg Env
	if err := yaml.Unmarshal([]byte(dataStr), &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file %s: %s", path, err)
	}

	return &cfg, nil
}

func initializeServerConfig(cfg *Env) error {
	var (
		err error
	)

	cfg.Server.ReadHeaderTimeout, err = time.ParseDuration(cfg.Server.ReadHeaderTimeoutDuration)
	if err != nil {
		return fmt.Errorf("failed to parse read header timeout duration: %v", err)
	}

	cfg.Server.ReadTimeout, err = time.ParseDuration(cfg.Server.ReadTimeoutDuration)
	if err != nil {
		return fmt.Errorf("failed to parse read timeout duration: %v", err)
	}

	cfg.Server.WriteTimeout, err = time.ParseDuration(cfg.Server.WriteTimeoutDuration)
	if err != nil {
		return fmt.Errorf("failed to parse write timeout duration: %v", err)
	}

	cfg.Server.IdleTimeout, err = time.ParseDuration(cfg.Server.IdleTimeoutDuration)
	if err != nil {
		return fmt.Errorf("failed to parse idle timeout duration: %v", err)
	}

	return nil
}

func initializeDatabaseConfig(cfg *Env) error {
	var (
		err error
	)

	cfg.Database.MaxConnectionLifetime, err = time.ParseDuration(cfg.Database.MaxConnectionLifetimeDuration)
	if err != nil {
		return fmt.Errorf("failed to parse max connection lifetime duration: %v", err)
	}

	cfg.Database.MaxIdleConnectionLifetime, err = time.ParseDuration(cfg.Database.MaxIdleConnectionLifetimeDuration)
	if err != nil {
		return fmt.Errorf("failed to parse max idle connection lifetime duration: %v", err)
	}

	return nil
}

func initializeJWTConfig(cfg *Env) error {
	var err error

	// Validate secret key
	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key is required. Set it in config.yaml")
	}

	if len(cfg.JWT.SecretKey) < 32 {
		logger.Warn("JWT secret key should be at least 32 characters for better security")
	}

	// Parse access token expiry
	cfg.JWT.AccessTokenExpiry, err = time.ParseDuration(cfg.JWT.AccessTokenExpiryStr)
	if err != nil {
		return fmt.Errorf("failed to parse JWT access token expiry duration: %v", err)
	}

	// Parse refresh token expiry
	cfg.JWT.RefreshTokenExpiry, err = time.ParseDuration(cfg.JWT.RefreshTokenExpiryStr)
	if err != nil {
		return fmt.Errorf("failed to parse JWT refresh token expiry duration: %v", err)
	}

	// Set default issuer if not provided
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "go-auth-service"
	}

	return nil
}
