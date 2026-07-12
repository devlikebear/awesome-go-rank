package awesomego

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devlikebear/awesome-go-rank/pkg/config"
)

type concurrencyGithubClient struct {
	readme   string
	active   atomic.Int32
	max      atomic.Int32
	finished atomic.Int32
}

func (c *concurrencyGithubClient) FetchReadmeMarkdown(context.Context, string, string) (string, error) {
	return c.readme, nil
}

func (c *concurrencyGithubClient) FetchRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	active := c.active.Add(1)
	defer c.active.Add(-1)
	for {
		max := c.max.Load()
		if active <= max || c.max.CompareAndSwap(max, active) {
			break
		}
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Millisecond):
	}
	c.finished.Add(1)
	return &Repository{Name: owner + "/" + repo, Stars: 1}, nil
}

func (c *concurrencyGithubClient) GetRateLimitInfo() RateLimitInfo {
	return RateLimitInfo{Remaining: 4999, Limit: 5000}
}

func TestFetchAndRankRepositoriesUsesBoundedWorkerPool(t *testing.T) {
	var readme strings.Builder
	readme.WriteString("## Database\n")
	for i := 0; i < 24; i++ {
		fmt.Fprintf(&readme, "- [repo-%d](https://github.com/example/repo-%d) - Test.\n", i, i)
	}
	client := &concurrencyGithubClient{readme: readme.String()}
	cfg := config.Default()
	cfg.Collection.Workers = 3
	cfg.RateLimit.MaxRetries = 1
	ag := NewAwesomeGo(client, cfg)

	if err := ag.FetchAndRankRepositories("", 0); err != nil {
		t.Fatal(err)
	}
	if got := client.max.Load(); got < 2 || got > 3 {
		t.Fatalf("peak concurrency = %d, want between 2 and 3", got)
	}
	if got := client.finished.Load(); got != 24 {
		t.Fatalf("finished repositories = %d, want 24", got)
	}
}

func TestRetryWithBackoffStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	calls := 0

	err := retryWithBackoff(ctx, func() error {
		calls++
		return errors.New("temporary")
	}, 3)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("retry error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("cancellation took too long: %s", elapsed)
	}
	if calls != 1 {
		t.Fatalf("retry calls = %d, want 1 before cancellation", calls)
	}
}
