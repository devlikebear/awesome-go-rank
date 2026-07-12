package awesomego

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
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

// retryWithBackoff retries a function with exponential backoff
func retryWithBackoff(ctx context.Context, fn func() error, maxRetries int) error {
	attempts := maxRetries
	if attempts < 1 {
		attempts = 1
	}
	var err error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err = fn()
		if err == nil {
			return nil
		}
		if i < attempts-1 {
			base := time.Duration(1<<i) * time.Second
			waitTime := jitteredBackoff(base)
			timer := time.NewTimer(waitTime)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	return fmt.Errorf("max attempts (%d) exceeded: %w", attempts, err)
}

func jitteredBackoff(base time.Duration) time.Duration {
	jitter, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(base)))
	if err != nil {
		return base
	}
	return base/2 + time.Duration(jitter.Int64())
}

// Repository is a struct that represents a repository.
type Repository struct {
	Name           string
	URL            string
	Stars          int
	Forks          int
	LastUpdated    time.Time
	Description    string
	Archived       bool
	StarsDelta7d   *int
	StarsDelta30d  *int
	StarsGrowth7d  *float64
	StarsGrowth30d *float64
	IsNew          bool
}

// Section is a struct that represents a section in the Awesome Go list.
type Section struct {
	Name        string
	Description string
}

// AwesomeGo is the main struct for the awesome-go-ranking package.
type AwesomeGo struct {
	client     IGithubClient
	repos      map[string][]Repository
	sections   map[string]Section
	config     *config.Config
	maxRetries int
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
	}
}

// fetchRepoInfo fetches the repository info from GitHub.
func (ag *AwesomeGo) fetchRepoInfo(ctx context.Context, owner, repo string) (*Repository, error) {
	return ag.client.FetchRepository(ctx, owner, repo)
}

// FetchAndRankRepositories fetches the repositories in the Awesome Go list
func (ag *AwesomeGo) FetchAndRankRepositories(specificSection string, limit int) error {
	return ag.FetchAndRankRepositoriesContext(context.Background(), specificSection, limit)
}

type repositoryJob struct {
	section string
	index   int
	repo    Repository
}

type repositoryResult struct {
	repositoryJob
	err error
}

// FetchAndRankRepositoriesContext fetches repositories with cancellation support.
func (ag *AwesomeGo) FetchAndRankRepositoriesContext(ctx context.Context, specificSection string, limit int) error {
	owner := ag.config.GitHub.Owner
	repo := ag.config.GitHub.Repository
	readmeMarkdown, err := ag.client.FetchReadmeMarkdown(ctx, owner, repo)
	if err != nil {
		return err
	}

	ag.parseMarkdown(readmeMarkdown)

	jobs := ag.collectionJobs(specificSection, limit)
	progressBar := pb.StartNew(len(jobs))
	jobChannel := make(chan repositoryJob, len(jobs))
	resultChannel := make(chan repositoryResult, len(jobs))
	workers := ag.config.Collection.Workers
	if workers > len(jobs) {
		workers = len(jobs)
	}
	if workers < 1 && len(jobs) > 0 {
		workers = 1
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChannel {
				resultChannel <- ag.collectRepository(ctx, job)
			}
		}()
	}
	for _, job := range jobs {
		jobChannel <- job
	}
	close(jobChannel)
	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	successful := make(map[string]map[int]Repository)
	failed := 0
	for result := range resultChannel {
		progressBar.Increment()
		if result.err != nil {
			failed++
			owner, name := extractRepoURLs(result.repo.URL)
			logger.Error("Failed to fetch repository after retries",
				zap.String("owner", owner),
				zap.String("repo", name),
				zap.Error(result.err))
			continue
		}
		if successful[result.section] == nil {
			successful[result.section] = make(map[int]Repository)
		}
		successful[result.section][result.index] = result.repo
	}
	progressBar.Finish()
	ag.applyCollectionResults(specificSection, successful)

	attempted := len(jobs)
	rateLimit := ag.client.GetRateLimitInfo()
	logger.Info("Repository collection summary",
		zap.Int("collected", attempted-failed),
		zap.Int("failed", failed),
		zap.Int("rate_limit_remaining", rateLimit.Remaining))

	if err := ctx.Err(); err != nil {
		return err
	}
	if attempted > 0 {
		failureRate := float64(failed) / float64(attempted)
		if failureRate > ag.config.Collection.FailureThreshold {
			return fmt.Errorf("repository collection failure rate %.2f%% exceeds threshold %.2f%% (%d/%d)",
				failureRate*100, ag.config.Collection.FailureThreshold*100, failed, attempted)
		}
	}

	return nil
}

func (ag *AwesomeGo) collectionJobs(specificSection string, limit int) []repositoryJob {
	sections := make([]string, 0, len(ag.repos))
	for section := range ag.repos {
		sections = append(sections, section)
	}
	sort.Strings(sections)
	jobs := make([]repositoryJob, 0)
	for _, section := range sections {
		if !MatchesSection(specificSection, section) {
			continue
		}
		for i, repo := range ag.repos[section] {
			if limit > 0 && len(jobs) >= limit {
				return jobs
			}
			jobs = append(jobs, repositoryJob{section: section, index: i, repo: repo})
		}
	}
	return jobs
}

func (ag *AwesomeGo) collectRepository(ctx context.Context, job repositoryJob) repositoryResult {
	owner, name := extractRepoURLs(job.repo.URL)
	if owner == "" || name == "" {
		return repositoryResult{repositoryJob: job, err: fmt.Errorf("invalid repository URL %q", job.repo.URL)}
	}
	err := retryWithBackoff(ctx, func() error {
		repoInfo, err := ag.fetchRepoInfo(ctx, owner, name)
		if err != nil {
			return err
		}
		job.repo.Stars = repoInfo.Stars
		job.repo.Forks = repoInfo.Forks
		job.repo.LastUpdated = repoInfo.LastUpdated
		job.repo.Archived = repoInfo.Archived
		return nil
	}, ag.maxRetries)
	return repositoryResult{repositoryJob: job, err: err}
}

func (ag *AwesomeGo) applyCollectionResults(specificSection string, successful map[string]map[int]Repository) {
	for section := range ag.repos {
		if !MatchesSection(specificSection, section) {
			if specificSection != "" {
				ag.repos[section] = nil
			}
			continue
		}
		byIndex := successful[section]
		indices := make([]int, 0, len(byIndex))
		for index := range byIndex {
			indices = append(indices, index)
		}
		sort.Ints(indices)
		collected := make([]Repository, 0, len(indices))
		for _, index := range indices {
			collected = append(collected, byIndex[index])
		}
		ag.repos[section] = collected
	}
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
