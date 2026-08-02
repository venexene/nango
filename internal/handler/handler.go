package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/venexene/nango/internal/config"
	"github.com/venexene/nango/internal/repository"
)


// Handler holds HTTP handler methods for the Shortener API.
type Handler struct {
	repo   repository.Interface
	logger *slog.Logger
	cfg    *config.Config
}

// NewHandler creates a Handler with the given dependencies.
func NewHandler(repo repository.Interface, logger *slog.Logger, cfg *config.Config) *Handler {
	return &Handler{
		repo:   repo,
		logger: logger,
		cfg:    cfg,
	}
}

// LiveCheckHandle returns the server health status.
func (h *Handler) LiveCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}
