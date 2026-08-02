package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/venexene/nango/internal/config"
	"github.com/venexene/nango/internal/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestHandler(repo repository.Interface) *Handler {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	return NewHandler(repo, logger, &config.Config{BaseURL: "http://localhost:8080"})
}

func TestShortenHandle_Success(t *testing.T) {
	t.Parallel()

	repo := &handlerMockRepo{
		getLinkByOriginalURLFn: func(_ string) (repository.Link, error) {
			return repository.Link{}, errNoRows
		},
		getLinkByShortCodeFn: func(_ string) (repository.Link, error) {
			return repository.Link{}, errNoRows
		},
		createLinkFn: func(arg repository.CreateLinkParams) (repository.Link, error) {
			return repository.Link{
				ID:          1,
				ShortCode:   arg.ShortCode,
				OriginalUrl: arg.OriginalUrl,
			}, nil
		},
	}

	h := newTestHandler(repo)
	router := gin.New()
	router.POST("/shorten", h.ShortenHandle)

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp map[string]any
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if _, ok := resp["short_url"]; !ok {
		t.Error("response missing short_url")
	}
	shortURL := resp["short_url"].(string)
	if !strings.Contains(shortURL, "/s/") {
		t.Errorf("short_url = %q, want contains /s/", shortURL)
	}
}

func TestShortenHandle_Duplicate(t *testing.T) {
	t.Parallel()

	repo := &handlerMockRepo{
		getLinkByOriginalURLFn: func(_ string) (repository.Link, error) {
			return repository.Link{
				ID:          5,
				ShortCode:   "abc1234",
				OriginalUrl: "https://example.com",
			}, nil
		},
	}

	h := newTestHandler(repo)
	router := gin.New()
	router.POST("/shorten", h.ShortenHandle)

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (duplicate should return 200)", w.Code, http.StatusOK)
	}
}

func TestShortenHandle_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&handlerMockRepo{})
	router := gin.New()
	router.POST("/shorten", h.ShortenHandle)

	body := `{"url":`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestShortenHandle_InvalidURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{"empty url", `{}`},
		{"url not a url", `{"url":"hello"}`},
		{"missing url field", `{"other":"value"}`},
		{"empty string url", `{"url":""}`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newTestHandler(&handlerMockRepo{})
			router := gin.New()
			router.POST("/shorten", h.ShortenHandle)

			req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d for body %q", w.Code, http.StatusBadRequest, tt.body)
			}
		})
	}
}

func TestRedirectHandle_Success(t *testing.T) {
	t.Parallel()

	repo := &handlerMockRepo{
		getLinkByShortCodeFn: func(_ string) (repository.Link, error) {
			return repository.Link{
				ID:          5,
				ShortCode:   "abc1234",
				OriginalUrl: "https://example.com",
			}, nil
		},
		recordClickFn: func(_ repository.RecordClickParams) error {
			return nil
		},
	}

	h := newTestHandler(repo)
	router := gin.New()
	router.GET("/s/:shortCode", h.RedirectHandle)

	req := httptest.NewRequest(http.MethodGet, "/s/abc1234", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMovedPermanently)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com" {
		t.Errorf("Location = %q, want %q", loc, "https://example.com")
	}
}

func TestRedirectHandle_NotFound(t *testing.T) {
	t.Parallel()

	repo := &handlerMockRepo{
		getLinkByShortCodeFn: func(_ string) (repository.Link, error) {
			return repository.Link{}, errNoRows
		},
	}

	h := newTestHandler(repo)
	router := gin.New()
	router.GET("/s/:shortCode", h.RedirectHandle)

	req := httptest.NewRequest(http.MethodGet, "/s/nonexist", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRedirectHandle_EmptyCode(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&handlerMockRepo{})
	router := gin.New()
	router.GET("/s/", h.RedirectHandle)

	req := httptest.NewRequest(http.MethodGet, "/s/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAnalyticsHandle_Success(t *testing.T) {
	t.Parallel()

	repo := &handlerMockRepo{
		getLinkByShortCodeFn: func(_ string) (repository.Link, error) {
			return repository.Link{ID: 5, ShortCode: "abc1234"}, nil
		},
		getTotalClicksFn: func(_ int32) (int32, error) {
			return 42, nil
		},
		getClicksByDayFn: func(_ int32) ([]repository.GetClicksByDayRow, error) {
			return []repository.GetClicksByDayRow{
				{Day: "2026-08-01", Count: 30},
				{Day: "2026-08-02", Count: 12},
			}, nil
		},
		getClicksByMonthFn: func(_ int32) ([]repository.GetClicksByMonthRow, error) {
			return []repository.GetClicksByMonthRow{
				{Month: "2026-08-01 00:00:00", Count: 42},
			}, nil
		},
		getClicksByUserAgentFn: func(_ int32) ([]repository.GetClicksByUserAgentRow, error) {
			return []repository.GetClicksByUserAgentRow{
				{UserAgent: "Chrome", Count: 30},
				{UserAgent: "Safari", Count: 12},
			}, nil
		},
	}

	h := newTestHandler(repo)
	router := gin.New()
	router.GET("/analytics/:shortCode", h.AnalyticsHandle)

	req := httptest.NewRequest(http.MethodGet, "/analytics/abc1234", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["total_clicks"] != float64(42) {
		t.Errorf("total_clicks = %v, want 42", resp["total_clicks"])
	}
}

func TestAnalyticsHandle_NotFound(t *testing.T) {
	t.Parallel()

	repo := &handlerMockRepo{
		getLinkByShortCodeFn: func(_ string) (repository.Link, error) {
			return repository.Link{}, errNoRows
		},
	}

	h := newTestHandler(repo)
	router := gin.New()
	router.GET("/analytics/:shortCode", h.AnalyticsHandle)

	req := httptest.NewRequest(http.MethodGet, "/analytics/nonexist", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// ── GET /health/live ──

func TestLiveCheckHandle(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&handlerMockRepo{})
	router := gin.New()
	router.GET("/health/live", h.LiveCheckHandle)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "UP" {
		t.Errorf("status = %q, want UP", resp["status"])
	}
}

func BenchmarkShortenHandle(b *testing.B) {
	gin.SetMode(gin.TestMode)

	repo := &handlerMockRepo{
		getLinkByOriginalURLFn: func(_ string) (repository.Link, error) {
			return repository.Link{}, errNoRows
		},
		getLinkByShortCodeFn: func(_ string) (repository.Link, error) {
			return repository.Link{}, errNoRows
		},
		createLinkFn: func(arg repository.CreateLinkParams) (repository.Link, error) {
			return repository.Link{ID: 1, ShortCode: arg.ShortCode}, nil
		},
	}

	h := newTestHandler(repo)
	router := gin.New()
	router.POST("/shorten", h.ShortenHandle)

	body := `{"url":"https://example.com"}`
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
