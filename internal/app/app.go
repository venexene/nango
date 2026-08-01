package app

import (
	"os"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"context"

	"github.com/venexene/nango/internal/config"
	"github.com/venexene/nango/internal/repository"
)

type Dependencies struct {
	Config     *config.Config
	Logger     *slog.Logger
	Repository *repository.Repository
}

func Run() error {
	dep := &Dependencies{}
	var err error

	dep.Config, err = config.Load(".env")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return fmt.Errorf("failed to load config: %w", err)
	}
	slog.Info("loaded config")

	var logHandler slog.Handler
	if dep.Config.LogFormat == "json" {
		logHandler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, nil)
	}
	dep.Logger = slog.New(logHandler)
	dep.Logger.Info("created logger")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dep.Repository, err = repository.NewRepository(ctx, dep.Config)
	if err != nil {
		dep.Logger.Error("failed to connect database", "error", err)
		return fmt.Errorf("failed to connect database: %w", err)
	}
	defer dep.Repository.Close()
	dep.Logger.Info("created repository")

	if err := dep.Repository.RunMigrations(); err != nil {
		dep.Logger.Error("failed to migrate database", "error", err)
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	dep.Logger.Info("migrated database")

	// work in progress
	return nil
}
