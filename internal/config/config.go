package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort       string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPass         string
	DBName         string
	DBMaxOpenConns int
	DBMaxIdleConns int
	BaseURL        string
}

const (
	DefaultHTTPPort       = "8080"
	DefaultDBPort         = "5432"
	DefaultDBMaxOpenConns = 25
	DefaultDBMaxIdleConns = 10
)

func (cfg *Config) applyDefaults() {
	if cfg.HTTPPort == "" {
		cfg.HTTPPort = DefaultHTTPPort
	}
	if cfg.DBPort == "" {
		cfg.DBPort = DefaultDBPort
	}
	if cfg.DBMaxOpenConns == 0 {
		cfg.DBMaxOpenConns = DefaultDBMaxOpenConns
	}
	if cfg.DBMaxIdleConns == 0 {
		cfg.DBMaxIdleConns = DefaultDBMaxIdleConns
	}
}

func (cfg *Config) validate() error {
	if cfg.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if cfg.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if cfg.DBPass == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if cfg.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if cfg.BaseURL == "" {
		return fmt.Errorf("BASE_URL is required")
	}
	if cfg.DBMaxOpenConns <= 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNECTIONS must be more than 0")
	}
	if cfg.DBMaxIdleConns <= 0 {
		return fmt.Errorf("DB_MAX_IDLE_CONNECTIONS must be more than 0")
	}
	return nil
}

func (cfg *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
}

func Load(path string) (*Config, error) {
	if err := godotenv.Load(path); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load .env: %w", err)
		}
		slog.Warn(".env file not found, using OS environment variables")
	}

	cfg := &Config{
		HTTPPort: os.Getenv("HTTP_PORT"),
		DBHost:   os.Getenv("DB_HOST"),
		DBPort:   os.Getenv("DB_PORT"),
		DBUser:   os.Getenv("DB_USER"),
		DBPass:   os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		BaseURL:  os.Getenv("BASE_URL"),
	}

	var err error

	if raw := os.Getenv("DB_MAX_OPEN_CONNECTIONS"); raw != "" {
		cfg.DBMaxOpenConns, err = strconv.Atoi(raw)
		if err != nil {
			slog.Warn("DB_MAX_OPEN_CONNECTIONS must be a number", "error", err)
		}
	}

	if raw := os.Getenv("DB_MAX_IDLE_CONNECTIONS"); raw != "" {
		cfg.DBMaxIdleConns, err = strconv.Atoi(raw)
		if err != nil {
			slog.Warn("DB_MAX_IDLE_CONNECTIONS must be a number", "error", err)
		}
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configs: %w", err)
	}

	return cfg, nil
}
