package service

import (
	"context"

	"github.com/venexene/nango/internal/repository"
)

type mockRepo struct {
	createLinkFn           func(ctx context.Context, arg repository.CreateLinkParams) (repository.Link, error)
	getLinkByShortCodeFn   func(ctx context.Context, shortCode string) (repository.Link, error)
	getLinkByOriginalURLFn func(ctx context.Context, originalUrl string) (repository.Link, error)
	recordClickFn          func(ctx context.Context, arg repository.RecordClickParams) error
	getClicksByDayFn       func(ctx context.Context, linkID int32) ([]repository.GetClicksByDayRow, error)
	getClicksByMonthFn     func(ctx context.Context, linkID int32) ([]repository.GetClicksByMonthRow, error)
	getClicksByUserAgentFn func(ctx context.Context, linkID int32) ([]repository.GetClicksByUserAgentRow, error)
	getTotalClicksFn       func(ctx context.Context, linkID int32) (int32, error)
}

func (m *mockRepo) CreateLink(ctx context.Context, arg repository.CreateLinkParams) (repository.Link, error) {
	return m.createLinkFn(ctx, arg)
}

func (m *mockRepo) GetLinkByShortCode(ctx context.Context, shortCode string) (repository.Link, error) {
	return m.getLinkByShortCodeFn(ctx, shortCode)
}

func (m *mockRepo) GetLinkByOriginalURL(ctx context.Context, originalUrl string) (repository.Link, error) {
	return m.getLinkByOriginalURLFn(ctx, originalUrl)
}

func (m *mockRepo) RecordClick(ctx context.Context, arg repository.RecordClickParams) error {
	return m.recordClickFn(ctx, arg)
}

func (m *mockRepo) GetClicksByDay(ctx context.Context, linkID int32) ([]repository.GetClicksByDayRow, error) {
	return m.getClicksByDayFn(ctx, linkID)
}

func (m *mockRepo) GetClicksByMonth(ctx context.Context, linkID int32) ([]repository.GetClicksByMonthRow, error) {
	return m.getClicksByMonthFn(ctx, linkID)
}

func (m *mockRepo) GetClicksByUserAgent(ctx context.Context, linkID int32) ([]repository.GetClicksByUserAgentRow, error) {
	return m.getClicksByUserAgentFn(ctx, linkID)
}

func (m *mockRepo) GetTotalClicks(ctx context.Context, linkID int32) (int32, error) {
	return m.getTotalClicksFn(ctx, linkID)
}

func (m *mockRepo) RunMigrations() error { return nil }
func (m *mockRepo) Close()               {}
