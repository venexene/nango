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

func TestAnalytics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupMock   func() *mockRepo
		shortCode   string
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, r *AnalyticsResult)
	}{
		{
			name:      "full analytics with all data",
			shortCode: "abc1234",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByShortCodeFn: func(ctx context.Context, code string) (repository.Link, error) {
						return repository.Link{
							ID:          5,
							ShortCode:   "abc1234",
							OriginalUrl: "https://example.com",
							CreatedAt:   pgtype.Timestamp{Time: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), Valid: true},
						}, nil
					},
					getTotalClicksFn: func(ctx context.Context, linkID int32) (int32, error) {
						return 100, nil
					},
					getClicksByDayFn: func(ctx context.Context, linkID int32) ([]repository.GetClicksByDayRow, error) {
						return []repository.GetClicksByDayRow{
							{Day: "2026-08-01", Count: 60},
							{Day: "2026-08-02", Count: 40},
						}, nil
					},
					getClicksByMonthFn: func(ctx context.Context, linkID int32) ([]repository.GetClicksByMonthRow, error) {
						return []repository.GetClicksByMonthRow{
							{Month: "2026-08-01 00:00:00", Count: 100},
						}, nil
					},
					getClicksByUserAgentFn: func(ctx context.Context, linkID int32) ([]repository.GetClicksByUserAgentRow, error) {
						return []repository.GetClicksByUserAgentRow{
							{UserAgent: "Mozilla/5.0 (iPhone)", Count: 70},
							{UserAgent: "Mozilla/5.0 (Windows NT)", Count: 30},
						}, nil
					},
				}
			},
			wantErr: false,
			checkResult: func(t *testing.T, r *AnalyticsResult) {
				if r.TotalClicks != 100 {
					t.Errorf("TotalClicks = %d, want 100", r.TotalClicks)
				}
				if len(r.Days) != 2 {
					t.Errorf("len(Days) = %d, want 2", len(r.Days))
				}
				if r.Days[0].Label != "2026-08-01" || r.Days[0].Count != 60 {
					t.Errorf("Days[0] = {%s, %d}, want {2026-08-01, 60}", r.Days[0].Label, r.Days[0].Count)
				}
				if r.Days[1].Label != "2026-08-02" || r.Days[1].Count != 40 {
					t.Errorf("Days[1] = {%s, %d}, want {2026-08-02, 40}", r.Days[1].Label, r.Days[1].Count)
				}
				if len(r.Months) != 1 {
					t.Errorf("len(Months) = %d, want 1", len(r.Months))
				}
				if r.Months[0].Count != 100 {
					t.Errorf("Months[0].Count = %d, want 100", r.Months[0].Count)
				}
				if len(r.UserAgents) != 2 {
					t.Errorf("len(UserAgents) = %d, want 2", len(r.UserAgents))
				}
				if r.UserAgents[0].Count != 70 {
					t.Errorf("UserAgents[0].Count = %d, want 70", r.UserAgents[0].Count)
				}
				if r.UserAgents[1].Label != "Mozilla/5.0 (Windows NT)" {
					t.Errorf("UserAgents[1].Label = %q", r.UserAgents[1].Label)
				}
			},
		},
		{
			name:      "empty analytics for new link",
			shortCode: "new0001",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByShortCodeFn: func(ctx context.Context, code string) (repository.Link, error) {
						return repository.Link{ID: 10, ShortCode: "new0001"}, nil
					},
					getTotalClicksFn: func(ctx context.Context, linkID int32) (int32, error) {
						return 0, nil
					},
					getClicksByDayFn: func(ctx context.Context, linkID int32) ([]repository.GetClicksByDayRow, error) {
						return []repository.GetClicksByDayRow{}, nil
					},
					getClicksByMonthFn: func(ctx context.Context, linkID int32) ([]repository.GetClicksByMonthRow, error) {
						return []repository.GetClicksByMonthRow{}, nil
					},
					getClicksByUserAgentFn: func(ctx context.Context, linkID int32) ([]repository.GetClicksByUserAgentRow, error) {
						return []repository.GetClicksByUserAgentRow{}, nil
					},
				}
			},
			wantErr: false,
			checkResult: func(t *testing.T, r *AnalyticsResult) {
				if r.TotalClicks != 0 {
					t.Errorf("TotalClicks = %d, want 0", r.TotalClicks)
				}
				if len(r.Days) != 0 {
					t.Errorf("len(Days) = %d, want 0", len(r.Days))
				}
				if len(r.Months) != 0 {
					t.Errorf("len(Months) = %d, want 0", len(r.Months))
				}
				if len(r.UserAgents) != 0 {
					t.Errorf("len(UserAgents) = %d, want 0", len(r.UserAgents))
				}
			},
		},
		{
			name:      "link not found returns error",
			shortCode: "nonex01",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByShortCodeFn: func(ctx context.Context, code string) (repository.Link, error) {
						return repository.Link{}, pgx.ErrNoRows
					},
				}
			},
			wantErr:     true,
			errContains: "failed to get link by short code",
		},
		{
			name:      "DB error on total clicks returns error",
			shortCode: "abc1234",
			setupMock: func() *mockRepo {
				return &mockRepo{
					getLinkByShortCodeFn: func(ctx context.Context, code string) (repository.Link, error) {
						return repository.Link{ID: 5, ShortCode: "abc1234"}, nil
					},
					getTotalClicksFn: func(ctx context.Context, linkID int32) (int32, error) {
						return 0, fmt.Errorf("connection lost")
					},
				}
			},
			wantErr:     true,
			errContains: "connection lost",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := tt.setupMock()
			result, err := Analytics(context.Background(), tt.shortCode, repo)

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
			if result == nil {
				t.Fatal("expected result, got nil")
			}
			tt.checkResult(t, result)
		})
	}
}
