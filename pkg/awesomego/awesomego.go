package awesomego

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cheggaaa/pb/v3"
)

// Repository is a struct that represents a repository.
type Repository struct {
	Name        string
	URL         string
	Stars       int
	Forks       int
	LastUpdated time.Time
}

// AwesomeGo is the main struct for the awesome-go-ranking package.
type AwesomeGo struct {
	client IGithubClient
	repos  map[string][]Repository
}

// NewAwesomeGo creates a new AwesomeGo instance.
func NewAwesomeGo(client IGithubClient) *AwesomeGo {
	return &AwesomeGo{client: client, repos: make(map[string][]Repository)}
}

// parseMarkdown parses the awesome-go README.md file and returns a map of
func parseMarkdown(input string, sections map[string][]Repository) {
	sectionRe := regexp.MustCompile(`^## (.+)$`)
	repoRe := regexp.MustCompile(`- \[(.+)\]\((https:\/\/github\.com\/[^)]+)\)`)

	var currentSection string

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		sectionMatches := sectionRe.FindStringSubmatch(line)
		repoMatches := repoRe.FindStringSubmatch(line)

		if len(sectionMatches) >= 2 {
			currentSection = sectionMatches[1]
			sections[currentSection] = []Repository{}
		} else if len(repoMatches) >= 3 {
			url := repoMatches[2]
			owner, name := extractRepoURLs(url)
			if currentSection != "" {
				sections[currentSection] = append(sections[currentSection],
					Repository{
						Name:  owner + "/" + name,
						URL:   url,
						Stars: 0,
						Forks: 0})
			}
		}
	}
}

// fetchRepoInfo fetches the repository info from GitHub.
func (ag *AwesomeGo) fetchRepoInfo(owner, repo string) (*Repository, error) {
	ctx := context.Background()
	return ag.client.FetchRepository(ctx, owner, repo)
}

// FetchAndRankRepositories fetches the repositories in the Awesome Go list
func (ag *AwesomeGo) FetchAndRankRepositories(specificSection string, limit int) error {
	owner, repo := "avelino", "awesome-go"
	readmeMarkdown, err := ag.client.FetchReadmeMarkdown(context.Background(), owner, repo)
	if err != nil {
		return err
	}

	//repoURLs := extractRepoURLs(readmeMarkdown)
	parseMarkdown(readmeMarkdown, ag.repos)

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
					repoInfo, err := ag.fetchRepoInfo(owner, name)
					if err == nil {
						reposMutex.Lock()
						repos[i].Stars = repoInfo.Stars
						repos[i].Forks = repoInfo.Forks
						repos[i].LastUpdated = repoInfo.LastUpdated
						reposMutex.Unlock()
					} else {
						fmt.Println(err)
					}
					// Sleep to avoid rate limit
					time.Sleep(100 * time.Microsecond)
				}
				progressBarMutex.Lock()
				progressBar.Increment() // Update progress bar
				progressBarMutex.Unlock()
			}(i, repos)
								// Sleep to avoid rate limit
			time.Sleep(1000 * time.Microsecond)

		}
	}

	wg.Wait()

	progressBar.Finish() // Complete progress bar

	return nil
}

// extractRepoURLs extracts the owner and repository name from a GitHub repository URL.
func extractRepoURLs(input string) (owner, repo string) {
	repoRegex := regexp.MustCompile(`^https://github.com/([^/]+)/([^/]+)$`)
	matches := repoRegex.FindStringSubmatch(input)
	if len(matches) == 3 {
		owner, repo = matches[1], matches[2]
	}

	return owner, repo
}

// Repositories returns the repositories in the Awesome Go list.
func (ag *AwesomeGo) Repositories() map[string][]Repository {
	return ag.repos
}
