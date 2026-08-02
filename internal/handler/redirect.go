package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/venexene/nango/internal/service"
)

// RedirectHandle looks up the short code from the URL path and redirects to the original URL.
func (h *Handler) RedirectHandle(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		h.logger.Warn("short code is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "short code is required",
		})
		return
	}

	originalURL, id, err := service.Redirect(c.Request.Context(), shortCode, h.repo)
	if err != nil {
		h.logger.Warn("failed to find link", "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to find link",
		})
		return
	}

	go func() {
		if err := service.RecordClick(context.Background(), id, c.GetHeader("User-Agent"), c.ClientIP(), h.repo); err != nil {
			h.logger.Error("failed to record link", "error", err)
		}
	}()

	c.Redirect(http.StatusMovedPermanently, originalURL)
}
