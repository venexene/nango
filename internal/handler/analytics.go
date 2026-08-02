package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/venexene/nango/internal/service"
)

func (h *Handler) AnalyticsHandle(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		h.logger.Warn("short code is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "short code is required",
		})
		return
	}

	result, err := service.Analytics(c.Request.Context(), shortCode, h.repo)
	if err != nil {
		h.logger.Warn("failed to get analytics", "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to get analytics",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
