package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/venexene/nango/internal/service"
)

// ShortenHandle creates a short link from a JSON body with a "url" field.
func (h *Handler) ShortenHandle(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to bind json",
		})
		return
	}

	shortenResult, err := service.ShortenURL(c.Request.Context(), req.URL, h.cfg.BaseURL, h.repo)
	if err != nil {
		h.logger.Error("failed to shorten url", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to shorten url",
		})
		return
	}

	if shortenResult.IsNew {
		c.JSON(http.StatusCreated, gin.H{
			"short_url": shortenResult.ShortURL,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"short_url": shortenResult.ShortURL,
		})
	}
}
