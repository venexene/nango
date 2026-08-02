package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/venexene/nango/internal/repository"
)

func TestGenerateShortCode(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		code, err := GenerateShortCode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != 7 {
			t.Errorf("expected length 7, got %d: %q", len(code), code)
		}
		for _, c := range code {
			if !strings.ContainsRune(alphabet, c) {
				t.Errorf("code %q contains invalid character %q", code, c)
			}
		}
	}

	// Проверка, что коды не повторяются (вероятностная)
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, _ := GenerateShortCode()
		if seen[code] {
			t.Errorf("duplicate code generated: %q", code)
		}
		seen[code] = true
	}
}

func TestShortenURL(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	baseURL := "http://localhost:8080"
	originalURL := "https://example.com"

	tests := []struct {
		name        string
		setupMock   func() *mockRepo
		wantIsNew   bool
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, r *ShortenResult)
	}{
		{
			name: "new URL creates short link",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByOriginalURLFn: func(_ context.Context, url string) (repository.Link, error) {
						return repository.Link{}, pgx.ErrNoRows
					},
					getLinkByShortCodeFn: func(_ context.Context, code string) (repository.Link, error) {
						return repository.Link{}, pgx.ErrNoRows
					},
					createLinkFn: func(_ context.Context, arg repository.CreateLinkParams) (repository.Link, error) {
						return repository.Link{
							ID:          1,
							ShortCode:   arg.ShortCode,
							OriginalUrl: arg.OriginalUrl,
							CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
						}, nil
					},
				}
			},
			wantIsNew: true,
			wantErr:   false,
			checkResult: func(t *testing.T, r *ShortenResult) {
				if r.OriginalURL != originalURL {
					t.Errorf("OriginalUrl = %q, want %q", r.OriginalURL, originalURL)
				}
				if len(r.ShortCode) != 7 {
					t.Errorf("ShortCode length = %d, want 7", len(r.ShortCode))
				}
				if !strings.HasPrefix(r.ShortURL, baseURL+"/s/") {
					t.Errorf("ShortURL = %q, want prefix %q", r.ShortURL, baseURL+"/s/")
				}
			},
		},
		{
			name: "existing URL returns existing link",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByOriginalURLFn: func(_ context.Context, url string) (repository.Link, error) {
						return repository.Link{
							ID:          5,
							ShortCode:   "abc1234",
							OriginalUrl: originalURL,
							CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
						}, nil
					},
				}
			},
			wantIsNew: false,
			wantErr:   false,
			checkResult: func(t *testing.T, r *ShortenResult) {
				if r.ShortCode != "abc1234" {
					t.Errorf("ShortCode = %q, want %q", r.ShortCode, "abc1234")
				}
				if r.ShortURL != baseURL+"/s/abc1234" {
					t.Errorf("ShortURL = %q, want %q", r.ShortURL, baseURL+"/s/abc1234")
				}
				if r.IsNew {
					t.Error("IsNew should be false for existing URL")
				}
			},
		},
		{
			name: "DB error on create returns error",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByOriginalURLFn: func(_ context.Context, url string) (repository.Link, error) {
						return repository.Link{}, pgx.ErrNoRows
					},
					getLinkByShortCodeFn: func(_ context.Context, code string) (repository.Link, error) {
						return repository.Link{}, pgx.ErrNoRows
					},
					createLinkFn: func(_ context.Context, arg repository.CreateLinkParams) (repository.Link, error) {
						return repository.Link{}, fmt.Errorf("connection refused")
					},
				}
			},
			wantIsNew:   false,
			wantErr:     true,
			errContains: "connection refused",
		},
		{
			name: "all generated codes taken returns error",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByOriginalURLFn: func(_ context.Context, url string) (repository.Link, error) {
						return repository.Link{}, pgx.ErrNoRows
					},
					getLinkByShortCodeFn: func(_ context.Context, code string) (repository.Link, error) {
						return repository.Link{
							ShortCode: code,
						}, nil // все коды заняты
					},
				}
			},
			wantIsNew:   false,
			wantErr:     true,
			errContains: "failed to generate unique short code after 5 attempts",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := tt.setupMock()
			result, err := ShortenURL(context.Background(), originalURL, baseURL, repo)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want contains %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.IsNew != tt.wantIsNew {
				t.Errorf("IsNew = %v, want %v", result.IsNew, tt.wantIsNew)
			}
			tt.checkResult(t, result)
		})
	}
}

func TestGenerateUniqueShortCode_Collision(t *testing.T) {
	t.Parallel()

	callCount := 0
	repo := &mockRepo{
		getLinkByShortCodeFn: func(_ context.Context, code string) (repository.Link, error) {
			callCount++
			if callCount < 3 {
				return repository.Link{ShortCode: code}, nil // занят
			}
			return repository.Link{}, pgx.ErrNoRows // свободен
		},
	}

	code, err := GenerateUniqueShortCode(context.Background(), repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != 7 {
		t.Errorf("code length = %d, want 7", len(code))
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}
