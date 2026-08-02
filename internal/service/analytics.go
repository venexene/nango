package service

import (
	"context"
	"fmt"

	"github.com/venexene/nango/internal/repository"
)

type AnalyticsRecord struct {
	Label string `json:"label"`
	Count int32  `json:"count"`
}

type AnalyticsResult struct {
	TotalClicks int32             `json:"total_clicks"`
	Days        []AnalyticsRecord `json:"days"`
	Months      []AnalyticsRecord `json:"months"`
	UserAgents  []AnalyticsRecord `json:"user_agents"`
}

func Analytics(ctx context.Context, shortCode string, repo repository.Interface) (*AnalyticsResult, error) {
	result := &AnalyticsResult{}

	link, err := repo.GetLinkByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get link by short code: %w", err)
	}

	result.TotalClicks, err = repo.GetTotalClicks(ctx, link.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total clicks: %w", err)
	}

	clicksDays, err := repo.GetClicksByDay(ctx, link.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get clicks by day: %w", err)
	}
	for _, day := range clicksDays {
		result.Days = append(result.Days, AnalyticsRecord{day.Day, day.Count})
	}

	clicksMonths, err := repo.GetClicksByMonth(ctx, link.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get clicks by month: %w", err)
	}
	for _, month := range clicksMonths {
		result.Months = append(result.Months, AnalyticsRecord{month.Month, month.Count})
	}

	clicksUserAgents, err := repo.GetClicksByUserAgent(ctx, link.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get clicks by user agent: %w", err)
	}
	for _, ug := range clicksUserAgents {
		result.UserAgents = append(result.UserAgents, AnalyticsRecord{ug.UserAgent, ug.Count})
	}

	return result, nil
}
