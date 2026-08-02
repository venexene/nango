package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDSN(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		DBUser:    "testuser",
		DBPass:    "testpass",
		DBHost:    "localhost",
		DBPort:    "5432",
		DBName:    "testdb",
		DBSSLMode: "disable",
	}

	dsn := cfg.DSN()
	expected := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"

	if dsn != expected {
		t.Errorf("DSN = %q, want %q", dsn, expected)
	}
}

func TestDSN_DefaultSSLMode(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		DBUser:    "u",
		DBPass:    "p",
		DBHost:    "h",
		DBPort:    "1234",
		DBName:    "d",
		DBSSLMode: "require",
	}

	dsn := cfg.DSN()
	if !strings.Contains(dsn, "sslmode=require") {
		t.Errorf("DSN %q should contain sslmode=require", dsn)
	}
}

func TestApplyDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      Config
		wantFunc func(cfg Config) bool
	}{
		{
			name: "empty HTTPPort gets default",
			cfg:  Config{},
			wantFunc: func(cfg Config) bool {
				return cfg.HTTPPort == DefaultHTTPPort
			},
		},
		{
			name: "custom HTTPPort kept",
			cfg:  Config{HTTPPort: "9090"},
			wantFunc: func(cfg Config) bool {
				return cfg.HTTPPort == "9090"
			},
		},
		{
			name: "empty DBPort gets default",
			cfg:  Config{},
			wantFunc: func(cfg Config) bool {
				return cfg.DBPort == DefaultDBPort
			},
		},
		{
			name: "empty DBSSLMode gets default",
			cfg:  Config{},
			wantFunc: func(cfg Config) bool {
				return cfg.DBSSLMode == DefaultSSLMode
			},
		},
		{
			name: "empty MigrationDir gets default",
			cfg:  Config{},
			wantFunc: func(cfg Config) bool {
				return cfg.MigrationDir == DefaultMigrationDir
			},
		},
		{
			name: "zero DBMaxOpenConns gets default",
			cfg:  Config{},
			wantFunc: func(cfg Config) bool {
				return cfg.DBMaxOpenConns == DefaultDBMaxOpenConns
			},
		},
		{
			name: "zero DBMaxIdleConns gets default",
			cfg:  Config{},
			wantFunc: func(cfg Config) bool {
				return cfg.DBMaxIdleConns == DefaultDBMaxIdleConns
			},
		},
		{
			name: "LogFormat empty gets default",
			cfg:  Config{},
			wantFunc: func(cfg Config) bool {
				return cfg.LogFormat == DefaultLogFormat
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := tt.cfg
			cfg.applyDefaults()
			if !tt.wantFunc(cfg) {
				t.Errorf("applyDefaults produced unexpected value for %s", tt.name)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: Config{
				DBHost:         "localhost",
				DBUser:         "user",
				DBPass:         "pass",
				DBName:         "db",
				BaseURL:        "http://localhost",
				DBMaxOpenConns: 10,
				DBMaxIdleConns: 5,
			},
			wantErr: false,
		},
		{
			name:    "missing DBHost",
			cfg:     Config{DBUser: "u", DBPass: "p", DBName: "d", BaseURL: "http://h", DBMaxOpenConns: 1, DBMaxIdleConns: 1},
			wantErr: true,
			errMsg:  "DB_HOST is required",
		},
		{
			name:    "missing DBUser",
			cfg:     Config{DBHost: "h", DBPass: "p", DBName: "d", BaseURL: "http://h", DBMaxOpenConns: 1, DBMaxIdleConns: 1},
			wantErr: true,
			errMsg:  "DB_USER is required",
		},
		{
			name:    "missing DBPass",
			cfg:     Config{DBHost: "h", DBUser: "u", DBName: "d", BaseURL: "http://h", DBMaxOpenConns: 1, DBMaxIdleConns: 1},
			wantErr: true,
			errMsg:  "DB_PASSWORD is required",
		},
		{
			name:    "missing DBName",
			cfg:     Config{DBHost: "h", DBUser: "u", DBPass: "p", BaseURL: "http://h", DBMaxOpenConns: 1, DBMaxIdleConns: 1},
			wantErr: true,
			errMsg:  "DB_NAME is required",
		},
		{
			name:    "missing BaseURL",
			cfg:     Config{DBHost: "h", DBUser: "u", DBPass: "p", DBName: "d", DBMaxOpenConns: 1, DBMaxIdleConns: 1},
			wantErr: true,
			errMsg:  "BASE_URL is required",
		},
		{
			name:    "zero max open conns",
			cfg:     Config{DBHost: "h", DBUser: "u", DBPass: "p", DBName: "d", BaseURL: "http://h", DBMaxOpenConns: 0, DBMaxIdleConns: 1},
			wantErr: true,
			errMsg:  "DB_MAX_OPEN_CONNECTIONS must be more than 0",
		},
		{
			name:    "zero max idle conns",
			cfg:     Config{DBHost: "h", DBUser: "u", DBPass: "p", DBName: "d", BaseURL: "http://h", DBMaxOpenConns: 1, DBMaxIdleConns: 0},
			wantErr: true,
			errMsg:  "DB_MAX_IDLE_CONNECTIONS must be more than 0",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %v, want contains %q", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env.test")

	content := `HTTP_PORT=9090
DB_HOST=testhost
DB_PORT=5432
DB_USER=testuser
DB_PASSWORD=testpass
DB_NAME=testdb
BASE_URL=http://test.example.com
DB_MAX_OPEN_CONNECTIONS=20
DB_MAX_IDLE_CONNECTIONS=8
`
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp .env: %v", err)
	}

	// Очищаем переменные окружения, чтобы тест был изолированным
	for _, k := range []string{"HTTP_PORT", "DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "BASE_URL", "DB_MAX_OPEN_CONNECTIONS", "DB_MAX_IDLE_CONNECTIONS", "DB_SSL_MODE", "MIGRATION_DIR", "LOG_FORMAT"} {
		_ = os.Unsetenv(k)
	}

	cfg, err := Load(envPath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.HTTPPort != "9090" {
		t.Errorf("HTTPPort = %q, want 9090", cfg.HTTPPort)
	}
	if cfg.DBHost != "testhost" {
		t.Errorf("DBHost = %q, want testhost", cfg.DBHost)
	}
	if cfg.DBMaxOpenConns != 20 {
		t.Errorf("DBMaxOpenConns = %d, want 20", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 8 {
		t.Errorf("DBMaxIdleConns = %d, want 8", cfg.DBMaxIdleConns)
	}
	if cfg.BaseURL != "http://test.example.com" {
		t.Errorf("BaseURL = %q, want http://test.example.com", cfg.BaseURL)
	}
}

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env.defaults")

	content := `DB_HOST=testhost
DB_USER=testuser
DB_PASSWORD=testpass
DB_NAME=testdb
BASE_URL=http://test.example.com
`
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp .env: %v", err)
	}

	for _, k := range []string{"HTTP_PORT", "DB_PORT", "DB_MAX_OPEN_CONNECTIONS", "DB_MAX_IDLE_CONNECTIONS", "DB_SSL_MODE", "MIGRATION_DIR", "LOG_FORMAT"} {
		_ = os.Unsetenv(k)
	}

	cfg, err := Load(envPath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.DBPort != DefaultDBPort {
		t.Errorf("DBPort = %q, want %q", cfg.DBPort, DefaultDBPort)
	}
	if cfg.DBSSLMode != DefaultSSLMode {
		t.Errorf("DBSSLMode = %q, want %q", cfg.DBSSLMode, DefaultSSLMode)
	}
	if cfg.DBMaxOpenConns != DefaultDBMaxOpenConns {
		t.Errorf("DBMaxOpenConns = %d, want %d", cfg.DBMaxOpenConns, DefaultDBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != DefaultDBMaxIdleConns {
		t.Errorf("DBMaxIdleConns = %d, want %d", cfg.DBMaxIdleConns, DefaultDBMaxIdleConns)
	}
}
