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
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
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
	}

	if len(l.errors) > 0 {
		return nil, fmt.Errorf("errors required env vars: %s", strings.Join(l.errors, ", "))
	}

	return cfg, nil
}

func LoadServiceEnvironmentVariables() (*Env, error) {
	const (
		configPath = `config.yaml`
	)

	if !fileExists(configPath) {
		return loadFromEnv()
	}

	cfg, err := loadFromYAML(configPath)
	if err != nil {
		return nil, fmt.Errorf("load from yaml: %w", err)
	}

	err = initializeServerConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize server config: %w", err)
	}

	return cfg, nil
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
		return nil, fmt.Errorf("read yaml: %w", err)
	}

	var cfg Env
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	return &cfg, nil
}

func initializeServerConfig(cfg *Env) error {
	var (
		err error
	)

	cfg.Server.ReadHeaderTimeout, err = time.ParseDuration(cfg.Server.ReadHeaderTimeoutDuration)
	if err != nil {
		logger.Fatalf("Failed to parse read header timeout duration: %v", err)
		return err
	}

	cfg.Server.ReadTimeout, err = time.ParseDuration(cfg.Server.ReadTimeoutDuration)
	if err != nil {
		logger.Fatalf("Failed to parse read timeout duration: %v", err)
		return err
	}

	cfg.Server.WriteTimeout, err = time.ParseDuration(cfg.Server.WriteTimeoutDuration)
	if err != nil {
		logger.Fatalf("Failed to parse write timeout duration: %v", err)
		return err
	}

	cfg.Server.IdleTimeout, err = time.ParseDuration(cfg.Server.IdleTimeoutDuration)
	if err != nil {
		logger.Fatalf("Failed to parse idle timeout duration: %v", err)
		return err
	}

	return nil
}
