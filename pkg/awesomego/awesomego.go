package awesomego

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/devlikebear/awesome-go-rank/pkg/config"
	"github.com/devlikebear/awesome-go-rank/pkg/logger"
	"go.uber.org/zap"
)

// Package-level compiled regex patterns to avoid recompilation
var (
	sectionRe = regexp.MustCompile(`^## (.+)$`)
	repoRe    = regexp.MustCompile(`- \[(.+)\]\((https:\/\/github\.com\/[^)]+)\)(.*$)`)
	repoURLRe = regexp.MustCompile(`^https://github.com/([^/]+)/([^/]+)$`)
)

// rateLimiter implements a simple rate limiting mechanism
type rateLimiter struct {
	lastCall    time.Time
	mu          sync.Mutex
	minInterval time.Duration
}

// Wait blocks until enough time has passed since the last call
func (rl *rateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	elapsed := time.Since(rl.lastCall)
	if elapsed < rl.minInterval {
		time.Sleep(rl.minInterval - elapsed)
	}
	rl.lastCall = time.Now()
}

// retryWithBackoff retries a function with exponential backoff
func retryWithBackoff(fn func() error, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if i < maxRetries-1 {
			waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
			time.Sleep(waitTime)
		}
	}
	return fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, err)
}

// Repository is a struct that represents a repository.
type Repository struct {
	Name        string
	URL         string
	Stars       int
	Forks       int
	LastUpdated time.Time
	Description string
}

// Section is a struct that represents a section in the Awesome Go list.
type Section struct {
	Name string
	Description string
}

// AwesomeGo is the main struct for the awesome-go-ranking package.
type AwesomeGo struct {
	client      IGithubClient
	repos       map[string][]Repository
	sections    map[string]Section
	rateLimiter *rateLimiter
	config      *config.Config
	maxRetries  int
}

// NewAwesomeGo creates a new AwesomeGo instance.
func NewAwesomeGo(client IGithubClient, cfg *config.Config) *AwesomeGo {
	if cfg == nil {
		cfg = config.Default()
	}
	return &AwesomeGo{
		client:     client,
		repos:      make(map[string][]Repository),
		sections:   make(map[string]Section),
		config:     cfg,
		maxRetries: cfg.RateLimit.MaxRetries,
		rateLimiter: &rateLimiter{
			minInterval: cfg.RateLimit.MinInterval,
			lastCall:    time.Now(),
		},
	}
}

// fetchRepoInfo fetches the repository info from GitHub.
func (ag *AwesomeGo) fetchRepoInfo(owner, repo string) (*Repository, error) {
	ctx := context.Background()
	return ag.client.FetchRepository(ctx, owner, repo)
}

// FetchAndRankRepositories fetches the repositories in the Awesome Go list
func (ag *AwesomeGo) FetchAndRankRepositories(specificSection string, limit int) error {
	owner := ag.config.GitHub.Owner
	repo := ag.config.GitHub.Repository
	readmeMarkdown, err := ag.client.FetchReadmeMarkdown(context.Background(), owner, repo)
	if err != nil {
		return err
	}

	//repoURLs := extractRepoURLs(readmeMarkdown)
	ag.parseMarkdown(readmeMarkdown)

	// Accumulate repositories
	reposCount := 0
	for section, repos := range ag.repos {
		// Skip section if specificSection is set and the section is not contain the specificSection
		if specificSection != "" && !strings.Contains(specificSection, section) {
			continue
		}
		reposCount += len(repos)
	}

	// Initialize progress bar
	progressBar := pb.StartNew(reposCount)

	// Parallelize fetching repository info
	var wg sync.WaitGroup
	var cnt int32 = 0
	var progressBarMutex sync.Mutex
	var reposMutex sync.Mutex

	for section, repos := range ag.repos {
		// Skip section if specificSection is set and the section is not contain the specificSection
		if specificSection != "" && !strings.Contains(specificSection, section) {
			continue
		}

		for i := range repos {
			// Stop if we reach the limit
			if limit > 0 && atomic.LoadInt32(&cnt) >= int32(limit) {
				break
			}
			atomic.AddInt32(&cnt, 1)

			wg.Add(1)
			go func(i int, repos []Repository) {
				defer wg.Done()

				owner, name := extractRepoURLs(repos[i].URL)
				if owner != "" && name != "" {
					// Rate limiting before API call
					ag.rateLimiter.Wait()

					// Retry logic with exponential backoff
					err := retryWithBackoff(func() error {
						repoInfo, err := ag.fetchRepoInfo(owner, name)
						if err != nil {
							return err
						}
						reposMutex.Lock()
						repos[i].Stars = repoInfo.Stars
						repos[i].Forks = repoInfo.Forks
						repos[i].LastUpdated = repoInfo.LastUpdated
						reposMutex.Unlock()
						return nil
					}, ag.maxRetries)

					if err != nil {
						logger.Error("Failed to fetch repository after retries",
							zap.String("owner", owner),
							zap.String("repo", name),
							zap.Error(err))
					}
				}
				progressBarMutex.Lock()
				progressBar.Increment() // Update progress bar
				progressBarMutex.Unlock()
			}(i, repos)

		}
	}

	wg.Wait()

	progressBar.Finish() // Complete progress bar

	return nil
}

// parseMarkdown parses the awesome-go README.md file and returns a map of
func (ag *AwesomeGo) parseMarkdown(input string) {
	var currentSection string

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		sectionMatches := sectionRe.FindStringSubmatch(line)
		repoMatches := repoRe.FindStringSubmatch(line)

		if len(sectionMatches) >= 2 {
			currentSection = sectionMatches[1]
			ag.repos[currentSection] = []Repository{}
		} else if len(repoMatches) >= 3 {
			url := repoMatches[2]
			owner, name := extractRepoURLs(url)
			if currentSection != "" {
				ag.repos[currentSection] = append(ag.repos[currentSection],
					Repository{
						Name:  owner + "/" + name,
						URL:   url,
						Stars: 0,
						Forks: 0,
						Description: strings.TrimSpace(repoMatches[3]),
					})
			}
		} else {
			// Check if the line is a section description
			if currentSection != "" && strings.HasPrefix(line, "_") {
				ag.sections[currentSection] = Section{
					Name: currentSection,
					Description: strings.Trim(line, "_"),
				}
			}
		}
	}
}


// extractRepoURLs extracts the owner and repository name from a GitHub repository URL.
func extractRepoURLs(input string) (owner, repo string) {
	matches := repoURLRe.FindStringSubmatch(input)
	if len(matches) == 3 {
		owner, repo = matches[1], matches[2]
	}

	return owner, repo
}

// Repositories returns the repositories in the Awesome Go list.
func (ag *AwesomeGo) Repositories() map[string][]Repository {
	return ag.repos
}

// Sections returns the sections in the Awesome Go list.
func (ag *AwesomeGo) Sections() map[string]Section {
	return ag.sections
}