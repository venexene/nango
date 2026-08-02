package repository

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/venexene/nango/internal/config"
)

type Repository struct {
	Querier
	pool          *pgxpool.Pool
	migrationPath string
}

type Interface interface {
	Querier
	RunMigrations() error
	Close()
}

func NewRepository(ctx context.Context, cfg *config.Config) (*Repository, error) {
	pool, err := CreatePool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	querier := New(pool)

	return &Repository{
		Querier:       querier,
		pool:          pool,
		migrationPath: cfg.MigrationDir,
	}, nil
}

func CreatePool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.DBMaxOpenConns)
	poolConfig.MinConns = int32(cfg.DBMaxIdleConns)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func (r *Repository) RunMigrations() error {
	connStr := r.pool.Config().ConnConfig.ConnString()

	absPath, err := filepath.Abs(r.migrationPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute migration path: %w", err)
	}

	m, err := migrate.New(fmt.Sprintf("file://%s", absPath), connStr)
	if err != nil {
		return fmt.Errorf("failed to init migrate: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to up migrate: %w", err)
	}

	return nil
}

func (r *Repository) Close() {
	r.pool.Close()
}
