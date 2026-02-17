package awesomego

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/devlikebear/awesome-go-rank/pkg/logger"
	"github.com/google/go-github/v68/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// RateLimitInfo holds GitHub API rate limit information
type RateLimitInfo struct {
	Remaining int
	Limit     int
	ResetTime time.Time
	mu        sync.RWMutex
}

// IGithubClient is an interface for the Github API
type IGithubClient interface {
	// FetchReadmeMarkdown fetches the README.md file of a given repository.
	FetchReadmeMarkdown(ctx context.Context, owner, repo string) (string, error)

	// FetchRepository fetches the repositories from the Github API
	FetchRepository(ctx context.Context, owner, repo string) (*Repository, error)

	// GetRateLimitInfo returns the current rate limit information
	GetRateLimitInfo() RateLimitInfo
}

// GithubClient is a struct that represents a Github client.
type GithubClient struct {
	client        *github.Client
	rateLimitInfo *RateLimitInfo
}

// NewGithubClient creates a new GithubClient instance.
func NewGithubClient(token string) *GithubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &GithubClient{
		client: client,
		rateLimitInfo: &RateLimitInfo{
			Remaining: 5000, // Default GitHub API limit
			Limit:     5000,
			ResetTime: time.Now().Add(1 * time.Hour),
		},
	}
}

// FetchRepository fetches the repositories from the Github API
func (gc *GithubClient) FetchRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	// Check rate limit before making request
	gc.waitIfNeeded()

	repoInfo, resp, err := gc.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	// Update rate limit info from response
	gc.updateRateLimitInfo(resp)

	return &Repository{
		Name:        repoInfo.GetFullName(),
		Stars:       repoInfo.GetStargazersCount(),
		Forks:       repoInfo.GetForksCount(),
		LastUpdated: repoInfo.GetUpdatedAt().Time,
	}, nil
}

// GetRateLimitInfo returns the current rate limit information
func (gc *GithubClient) GetRateLimitInfo() RateLimitInfo {
	gc.rateLimitInfo.mu.RLock()
	defer gc.rateLimitInfo.mu.RUnlock()

	return RateLimitInfo{
		Remaining: gc.rateLimitInfo.Remaining,
		Limit:     gc.rateLimitInfo.Limit,
		ResetTime: gc.rateLimitInfo.ResetTime,
	}
}

// updateRateLimitInfo updates the rate limit information from the API response
func (gc *GithubClient) updateRateLimitInfo(resp *github.Response) {
	if resp == nil || resp.Rate.Remaining == 0 {
		return
	}

	gc.rateLimitInfo.mu.Lock()
	defer gc.rateLimitInfo.mu.Unlock()

	gc.rateLimitInfo.Remaining = resp.Rate.Remaining
	gc.rateLimitInfo.Limit = resp.Rate.Limit
	gc.rateLimitInfo.ResetTime = resp.Rate.Reset.Time
}

// waitIfNeeded waits if the rate limit is close to being exceeded
func (gc *GithubClient) waitIfNeeded() {
	gc.rateLimitInfo.mu.RLock()
	remaining := gc.rateLimitInfo.Remaining
	resetTime := gc.rateLimitInfo.ResetTime
	gc.rateLimitInfo.mu.RUnlock()

	// If we have less than 10 requests remaining, wait until reset
	if remaining < 10 {
		waitTime := time.Until(resetTime)
		if waitTime > 0 {
			logger.Warn("Rate limit low, waiting for reset",
				zap.Int("remaining", remaining),
				zap.Duration("wait_time", waitTime),
				zap.Time("reset_time", resetTime))
			time.Sleep(waitTime)
		}
	}
}

// FetchReadmeMarkdown fetches the README.md file of a given repository.
func (ag *GithubClient) FetchReadmeMarkdown(ctx context.Context, owner, repo string) (string, error) {
	readmeURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/README.md", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, readmeURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("failed to fetch README.md: %s", resp.Status)
	}

	readmeMarkdown, _ := io.ReadAll(resp.Body)
	return string(readmeMarkdown), nil
}
