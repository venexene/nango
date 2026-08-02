package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/venexene/nango/internal/config"
	"github.com/venexene/nango/internal/handler"
	"github.com/venexene/nango/internal/repository"
)

type Dependencies struct {
	Config     *config.Config
	Logger     *slog.Logger
	Repository repository.Interface
	Router     *gin.Engine
	Server     *http.Server
	Handler    *handler.Handler
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

	dep.Router, err = createRouter(dep)
	if err != nil {
		dep.Logger.Error("failed to create router", "error", err)
		return fmt.Errorf("failed to create router: %w", err)
	}
	dep.Logger.Info("created router")

	dep.Server = &http.Server{
		Addr:         ":" + dep.Config.HTTPPort,
		Handler:      dep.Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	dep.Logger.Info("created server")

	errCh := make(chan error, 1)
	go func() {
		if err := dep.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	dep.Logger.Info("started HTTP server on port", "port", dep.Config.HTTPPort)

	select {
	case <-ctx.Done():
		dep.Logger.Info("shutting down server...")

		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := dep.Server.Shutdown(ctxShutdown); err != nil {
			dep.Logger.Error("failed to shutdown server", "error", err)
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
		dep.Logger.Info("shutdown server")
	case err := <-errCh:
		dep.Logger.Error("HTTP server error", "error", err)
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

func createRouter(dep *Dependencies) (*gin.Engine, error) {
	router := gin.Default()

	dep.Handler = handler.NewHandler(dep.Repository, dep.Logger, dep.Config)

	router.GET("/health/live", dep.Handler.LiveCheckHandle)

	router.POST("/shorten", dep.Handler.ShortenHandle)

	router.GET("/s/:shortCode", dep.Handler.RedirectHandle)

	return router, nil
}
