package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Env struct {
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
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
	missing []string
}

func (l *envLoader) req(key string) string {
	v := os.Getenv(key)
	if v == "" {
		l.missing = append(l.missing, key)
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
	}

	if len(l.missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %s", strings.Join(l.missing, ", "))
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
