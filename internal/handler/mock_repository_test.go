package handler

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/venexene/nango/internal/repository"
)

var errNoRows = pgx.ErrNoRows

type handlerMockRepo struct {
	createLinkFn           func(arg repository.CreateLinkParams) (repository.Link, error)
	getLinkByShortCodeFn   func(shortCode string) (repository.Link, error)
	getLinkByOriginalURLFn func(originalUrl string) (repository.Link, error)
	recordClickFn          func(arg repository.RecordClickParams) error
	getClicksByDayFn       func(linkID int32) ([]repository.GetClicksByDayRow, error)
	getClicksByMonthFn     func(linkID int32) ([]repository.GetClicksByMonthRow, error)
	getClicksByUserAgentFn func(linkID int32) ([]repository.GetClicksByUserAgentRow, error)
	getTotalClicksFn       func(linkID int32) (int32, error)
}

func (m *handlerMockRepo) CreateLink(_ context.Context, arg repository.CreateLinkParams) (repository.Link, error) {
	if m.createLinkFn == nil {
		return repository.Link{}, errors.New("CreateLink not mocked")
	}
	return m.createLinkFn(arg)
}

func (m *handlerMockRepo) GetLinkByShortCode(_ context.Context, shortCode string) (repository.Link, error) {
	if m.getLinkByShortCodeFn == nil {
		return repository.Link{}, errors.New("GetLinkByShortCode not mocked")
	}
	return m.getLinkByShortCodeFn(shortCode)
}

func (m *handlerMockRepo) GetLinkByOriginalURL(_ context.Context, originalURL string) (repository.Link, error) {
	if m.getLinkByOriginalURLFn == nil {
		return repository.Link{}, errors.New("GetLinkByOriginalURL not mocked")
	}
	return m.getLinkByOriginalURLFn(originalURL)
}

func (m *handlerMockRepo) RecordClick(_ context.Context, arg repository.RecordClickParams) error {
	if m.recordClickFn == nil {
		return errors.New("RecordClick not mocked")
	}
	return m.recordClickFn(arg)
}

func (m *handlerMockRepo) GetClicksByDay(_ context.Context, linkID int32) ([]repository.GetClicksByDayRow, error) {
	if m.getClicksByDayFn == nil {
		return nil, errors.New("GetClicksByDay not mocked")
	}
	return m.getClicksByDayFn(linkID)
}

func (m *handlerMockRepo) GetClicksByMonth(_ context.Context, linkID int32) ([]repository.GetClicksByMonthRow, error) {
	if m.getClicksByMonthFn == nil {
		return nil, errors.New("GetClicksByMonth not mocked")
	}
	return m.getClicksByMonthFn(linkID)
}

func (m *handlerMockRepo) GetClicksByUserAgent(_ context.Context, linkID int32) ([]repository.GetClicksByUserAgentRow, error) {
	if m.getClicksByUserAgentFn == nil {
		return nil, errors.New("GetClicksByUserAgent not mocked")
	}
	return m.getClicksByUserAgentFn(linkID)
}

func (m *handlerMockRepo) GetTotalClicks(_ context.Context, linkID int32) (int32, error) {
	if m.getTotalClicksFn == nil {
		return 0, errors.New("GetTotalClicks not mocked")
	}
	return m.getTotalClicksFn(linkID)
}

func (m *handlerMockRepo) RunMigrations() error { return nil }
func (m *handlerMockRepo) Close()               {}
