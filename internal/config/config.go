// Package config loads application settings from .env and OS environment,
// applies defaults, and validates required fields.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment.
type Config struct {
	HTTPPort       string
	LogFormat      string
	BaseURL        string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPass         string
	DBName         string
	DBSSLMode      string
	MigrationDir   string
	DBMaxOpenConns int
	DBMaxIdleConns int
}

const (
	// LogFormatText is the `text` value of LogFormat
	LogFormatText = "text"
	// LogFormatJSON is the `json` value of LogFormat
	LogFormatJSON = "json"
	// DefaultHTTPPort is the default listen port for the HTTP server.
	DefaultHTTPPort = "8080"
	// DefaultDBPort is the default PostgreSQL port.
	DefaultDBPort = "5432"
	// DefaultSSLMode is the default PostgreSQL SSL mode.
	DefaultSSLMode = "disable"
	// DefaultMigrationDir is the default path to SQL migration files.
	DefaultMigrationDir = "migrations"
	// DefaultDBMaxOpenConns is the default maximum number of open database connections.
	DefaultDBMaxOpenConns = 25
	// DefaultDBMaxIdleConns is the default maximum number of idle database connections.
	DefaultDBMaxIdleConns = 10
)

func (cfg *Config) applyDefaults() {
	if cfg.HTTPPort == "" {
		cfg.HTTPPort = DefaultHTTPPort
	}
	if cfg.LogFormat != LogFormatJSON && cfg.LogFormat != LogFormatText {
		cfg.LogFormat = LogFormatText
	}
	if cfg.DBPort == "" {
		cfg.DBPort = DefaultDBPort
	}
	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = DefaultSSLMode
	}
	if cfg.MigrationDir == "" {
		cfg.MigrationDir = DefaultMigrationDir
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

// DSN returns the PostgreSQL connection string in URI format.
func (cfg *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
	)
}

// Load reads configuration from the .env file at path and applies defaults.
func Load(path string) (*Config, error) {
	if err := godotenv.Load(path); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load .env: %w", err)
		}
		slog.Warn(".env file not found, using OS environment variables")
	}

	cfg := &Config{
		HTTPPort:     os.Getenv("HTTP_PORT"),
		LogFormat:    os.Getenv("LOG_FORMAT"),
		BaseURL:      os.Getenv("BASE_URL"),
		DBHost:       os.Getenv("DB_HOST"),
		DBPort:       os.Getenv("DB_PORT"),
		DBUser:       os.Getenv("DB_USER"),
		DBPass:       os.Getenv("DB_PASSWORD"),
		DBName:       os.Getenv("DB_NAME"),
		DBSSLMode:    os.Getenv("DB_SSL_MODE"),
		MigrationDir: os.Getenv("MIGRATION_DIR"),
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
