package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"github.com/venexene/nango/internal/repository"
)

const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	length = 7
)


func GenerateShortCode() (string, error) {
	bytes := make([]byte, length)

	for i := 0; i < length; i++ {
		randomByte := make([]byte, 1)
		if _, err := rand.Read(randomByte); err != nil {
			return "", fmt.Errorf("failed to generate random byte: %w", err)
		}
		bytes[i] = alphabet[int(randomByte[0])%len(alphabet)]
	}

	return string(bytes), nil
}

func GenerateUniqueShortCode(ctx context.Context, repo repository.Interface) (string, error) {
    for attempt := 0; attempt < 5; attempt++ {
        code, err := GenerateShortCode()
        if err != nil {
            return "", err
        }

        _, err = repo.GetLinkByShortCode(ctx, code)
        if err == sql.ErrNoRows {
            return code, nil
        }
        if err != nil {
            return "", fmt.Errorf("failed to check short code: %w", err)
        }
    }
    return "", fmt.Errorf("failed to generate unique short code after %d attempts", 5)
}

type ShortenResult struct {
	ShortCode string
	ShortURL  string
	OriginalUrl string
	CreatedAt time.Time
	IsNew       bool 
}

func ShortenURL(ctx context.Context, url string, baseURL string, repo repository.Interface) (*ShortenResult, error) {
	result := &ShortenResult{OriginalUrl: url}
	var err error

	if link, err := repo.GetLinkByOriginalURL(ctx, result.OriginalUrl); err == nil {
		result.ShortCode = link.ShortCode
		result.ShortURL = fmt.Sprintf("%s/s/%s", baseURL, result.ShortCode)
		result.CreatedAt = link.CreatedAt.Time
		result.IsNew = false
		return result, nil
	} 

	result.ShortCode, err = GenerateUniqueShortCode(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate short code: %w", err)
	}
	
	params := repository.CreateLinkParams{
		ShortCode: result.ShortCode,
		OriginalUrl: result.OriginalUrl,
	}

	link, err := repo.CreateLink(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create link in repository: %w", err)
	}

	result.ShortURL = fmt.Sprintf("%s/s/%s", baseURL, result.ShortCode)

	result.CreatedAt = link.CreatedAt.Time

	result.IsNew = true

	return result, nil
}