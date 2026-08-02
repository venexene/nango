package service

import (
	"context"
	"fmt"

	"github.com/venexene/nango/internal/repository"
)

// Redirect looks up a short code and returns the original URL and link ID.
func Redirect(ctx context.Context, shortCode string, repo repository.Interface) (string, int32, error) {
	link, err := repo.GetLinkByShortCode(ctx, shortCode)
	if err != nil {
		return "", 0, fmt.Errorf("failed to find short code in repository: %w", err)
	}

	return link.OriginalUrl, int32(link.ID), nil
}

// RecordClick persists a click event with user agent and IP information.
func RecordClick(ctx context.Context, linkId int32, userAgent string, ip string, repo repository.Interface) error {
	params := repository.RecordClickParams{
		LinkID:    linkId,
		UserAgent: userAgent,
		Ip:        ip,
	}

	if err := repo.RecordClick(ctx, params); err != nil {
		return fmt.Errorf("failed to record click: %w", err)
	}

	return nil
}
