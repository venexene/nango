package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/venexene/nango/internal/config"
	"github.com/venexene/nango/internal/repository"
)

const (
	statusDown = "DOWN"
	statusUp   = "UP"
)

type Handler struct {
	repo     repository.Interface
	logger   *slog.Logger
	cfg   *config.Config
}

func NewHandler(repo repository.Interface, logger *slog.Logger, cfg *config.Config) *Handler {
	return &Handler{
		repo:     repo,
		logger:   logger,
		cfg:   cfg,
	}
}

func (h *Handler) LiveCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": statusUp})
}