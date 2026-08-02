package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/venexene/nango/internal/repository"
)

func TestRedirect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		shortCode   string
		setupMock   func() *mockRepo
		wantURL     string
		wantLinkID  int32
		wantErr     bool
		errContains string
	}{
		{
			name:      "existing short code returns original URL",
			shortCode: "abc1234",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByShortCodeFn: func(ctx context.Context, code string) (repository.Link, error) {
						return repository.Link{
							ID:          5,
							ShortCode:   "abc1234",
							OriginalUrl: "https://example.com",
						}, nil
					},
				}
			},
			wantURL:    "https://example.com",
			wantLinkID: 5,
			wantErr:    false,
		},
		{
			name:      "non-existent short code returns error",
			shortCode: "nonex01",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByShortCodeFn: func(ctx context.Context, code string) (repository.Link, error) {
						return repository.Link{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:     true,
			errContains: "failed to find short code",
		},
		{
			name:      "DB error returns error",
			shortCode: "abc1234",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByShortCodeFn: func(ctx context.Context, code string) (repository.Link, error) {
						return repository.Link{}, fmt.Errorf("connection refused")
					},
				}
			},
			wantErr:     true,
			errContains: "connection refused",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := tt.setupMock()
			url, linkID, err := Redirect(context.Background(), tt.shortCode, repo)

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
			if url != tt.wantURL {
				t.Errorf("url = %q, want %q", url, tt.wantURL)
			}
			if linkID != tt.wantLinkID {
				t.Errorf("linkID = %d, want %d", linkID, tt.wantLinkID)
			}
		})
	}
}

func TestRecordClick(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupMock   func() *mockRepo
		linkID      int32
		userAgent   string
		ip          string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful click recording",
			setupMock: func() *mockRepo {
				return &mockRepo{
					recordClickFn: func(ctx context.Context, arg repository.RecordClickParams) error {
						if arg.LinkID != 5 {
							t.Errorf("LinkID = %d, want 5", arg.LinkID)
						}
						if arg.UserAgent != "Mozilla/5.0" {
							t.Errorf("UserAgent = %q, want Mozilla/5.0", arg.UserAgent)
						}
						if arg.Ip != "127.0.0.1" {
							t.Errorf("Ip = %q, want 127.0.0.1", arg.Ip)
						}
						return nil
					},
				}
			},
			linkID:    5,
			userAgent: "Mozilla/5.0",
			ip:        "127.0.0.1",
			wantErr:   false,
		},
		{
			name: "DB error on record",
			setupMock: func() *mockRepo {
				return &mockRepo{
					recordClickFn: func(ctx context.Context, arg repository.RecordClickParams) error {
						return fmt.Errorf("disk full")
					},
				}
			},
			linkID:      1,
			userAgent:   "",
			ip:          "",
			wantErr:     true,
			errContains: "disk full",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := tt.setupMock()
			err := RecordClick(context.Background(), tt.linkID, tt.userAgent, tt.ip, repo)

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
		})
	}
}

var _ repository.Interface = (*mockRepo)(nil)

func TestMockRepoImplementsInterface(t *testing.T) {
	_ = repository.Interface(&mockRepo{
		createLinkFn: func(ctx context.Context, arg repository.CreateLinkParams) (repository.Link, error) {
			return repository.Link{}, nil
		},
		getLinkByShortCodeFn:   func(ctx context.Context, sc string) (repository.Link, error) { return repository.Link{}, nil },
		getLinkByOriginalURLFn: func(ctx context.Context, url string) (repository.Link, error) { return repository.Link{}, nil },
		recordClickFn:          func(ctx context.Context, arg repository.RecordClickParams) error { return nil },
		getClicksByDayFn:       func(ctx context.Context, lid int32) ([]repository.GetClicksByDayRow, error) { return nil, nil },
		getClicksByMonthFn:     func(ctx context.Context, lid int32) ([]repository.GetClicksByMonthRow, error) { return nil, nil },
		getClicksByUserAgentFn: func(ctx context.Context, lid int32) ([]repository.GetClicksByUserAgentRow, error) { return nil, nil },
		getTotalClicksFn:       func(ctx context.Context, lid int32) (int32, error) { return 0, nil },
	})
}
