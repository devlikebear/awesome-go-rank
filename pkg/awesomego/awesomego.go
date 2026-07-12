package awesomego

import (
	"context"
	"fmt"
	"math"
	"net/url"
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
)

var githubReservedPaths = map[string]struct{}{
	"about": {}, "apps": {}, "collections": {}, "events": {}, "explore": {},
	"features": {}, "join": {}, "login": {}, "marketplace": {}, "new": {},
	"notifications": {}, "orgs": {}, "search": {}, "settings": {}, "sponsors": {},
	"topics": {}, "trending": {}, "users": {},
}

var githubNonRepositoryPages = map[string]struct{}{
	"actions": {}, "branches": {}, "commits": {}, "discussions": {}, "forks": {},
	"graphs": {}, "issues": {}, "network": {}, "projects": {}, "pulls": {},
	"releases": {}, "security": {}, "settings": {}, "stargazers": {}, "tags": {},
	"watchers": {}, "wiki": {},
}

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
	Name        string
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
		if !MatchesSection(specificSection, section) {
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
	failedRepos := make(map[string]map[int]struct{})
	var failedCount int32

	for section, repos := range ag.repos {
		// Skip section if specificSection is set and the section is not contain the specificSection
		if !MatchesSection(specificSection, section) {
			continue
		}

		for i := range repos {
			// Stop if we reach the limit
			if limit > 0 && atomic.LoadInt32(&cnt) >= int32(limit) {
				break
			}
			atomic.AddInt32(&cnt, 1)

			wg.Add(1)
			go func(section string, i int, repos []Repository) {
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
						reposMutex.Lock()
						if failedRepos[section] == nil {
							failedRepos[section] = make(map[int]struct{})
						}
						failedRepos[section][i] = struct{}{}
						reposMutex.Unlock()
						atomic.AddInt32(&failedCount, 1)
						logger.Error("Failed to fetch repository after retries",
							zap.String("owner", owner),
							zap.String("repo", name),
							zap.Error(err))
					}
				}
				progressBarMutex.Lock()
				progressBar.Increment() // Update progress bar
				progressBarMutex.Unlock()
			}(section, i, repos)

		}
	}

	wg.Wait()

	progressBar.Finish() // Complete progress bar

	for section, failed := range failedRepos {
		repos := ag.repos[section]
		collected := make([]Repository, 0, len(repos)-len(failed))
		for i, repo := range repos {
			if _, wasFailed := failed[i]; !wasFailed {
				collected = append(collected, repo)
			}
		}
		ag.repos[section] = collected
	}

	attempted := atomic.LoadInt32(&cnt)
	failed := atomic.LoadInt32(&failedCount)
	rateLimit := ag.client.GetRateLimitInfo()
	logger.Info("Repository collection summary",
		zap.Int32("collected", attempted-failed),
		zap.Int32("failed", failed),
		zap.Int("rate_limit_remaining", rateLimit.Remaining))

	if attempted > 0 {
		failureRate := float64(failed) / float64(attempted)
		if failureRate > ag.config.Collection.FailureThreshold {
			return fmt.Errorf("repository collection failure rate %.2f%% exceeds threshold %.2f%% (%d/%d)",
				failureRate*100, ag.config.Collection.FailureThreshold*100, failed, attempted)
		}
	}

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
			// Initialize section with name (description will be set later if available)
			if _, exists := ag.sections[currentSection]; !exists {
				ag.sections[currentSection] = Section{
					Name:        currentSection,
					Description: "",
				}
			}
		} else if len(repoMatches) >= 3 {
			url := repoMatches[2]
			owner, name := extractRepoURLs(url)
			// Only add repository if owner and name are valid
			if currentSection != "" && owner != "" && name != "" {
				ag.repos[currentSection] = append(ag.repos[currentSection],
					Repository{
						Name:        owner + "/" + name,
						URL:         url,
						Stars:       0,
						Forks:       0,
						Description: strings.TrimSpace(repoMatches[3]),
					})
			}
		} else {
			// Check if the line is a section description
			if currentSection != "" && strings.HasPrefix(line, "_") {
				ag.sections[currentSection] = Section{
					Name:        currentSection,
					Description: strings.Trim(line, "_"),
				}
			}
		}
	}
}

// MatchesSection reports whether an optional requested section exactly matches a section name.
func MatchesSection(requested, actual string) bool {
	return requested == "" || strings.EqualFold(strings.TrimSpace(requested), strings.TrimSpace(actual))
}

// extractRepoURLs extracts the owner and repository name from a direct GitHub repository URL.
func extractRepoURLs(input string) (owner, repo string) {
	parsed, err := url.Parse(input)
	if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Host, "github.com") {
		return "", ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", ""
	}
	if _, reserved := githubReservedPaths[strings.ToLower(parts[0])]; reserved {
		return "", ""
	}
	if len(parts) > 2 {
		if _, nonRepositoryPage := githubNonRepositoryPages[strings.ToLower(parts[2])]; nonRepositoryPage {
			return "", ""
		}
	}
	return parts[0], parts[1]
}

// Repositories returns the repositories in the Awesome Go list.
func (ag *AwesomeGo) Repositories() map[string][]Repository {
	return ag.repos
}

// Sections returns the sections in the Awesome Go list.
func (ag *AwesomeGo) Sections() map[string]Section {
	return ag.sections
}
