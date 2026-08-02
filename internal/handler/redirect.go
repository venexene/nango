package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/venexene/nango/internal/service"
)

func (h *Handler) RedirectHandle(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		h.logger.Error("short code is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"short code is required",
		})
		return
	}

	originalUrl, id, err := service.Redirect(c.Request.Context(), shortCode, h.repo)
	if err != nil {
		h.logger.Warn("failed to find link", "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":"failed to find link",
		})
		return
	}

	go func() {
		if err := service.RecordClick(context.Background(), id, c.GetHeader("User-Agent"), c.ClientIP(), h.repo); err != nil {
			h.logger.Error("failed to record link", "error", err)
		}
	}()

	c.Redirect(http.StatusMovedPermanently, originalUrl)
}
