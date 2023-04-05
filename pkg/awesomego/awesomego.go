package awesomego

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
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
	client *github.Client
	repos  map[string][]Repository
}

// NewAwesomeGo creates a new AwesomeGo instance.
func NewAwesomeGo(token string) *AwesomeGo {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &AwesomeGo{client: client, repos: make(map[string][]Repository)}
}

// fetchReadmeMarkdown fetches the awesome-go README.md file.
func (ag *AwesomeGo) fetchReadmeMarkdown(owner, repo string) (string, error) {
	readmeURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/README.md", owner, repo)
	resp, err := http.Get(readmeURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	readmeMarkdown, _ := ioutil.ReadAll(resp.Body)
	return string(readmeMarkdown), nil
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
	repoInfo, _, err := ag.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return &Repository{
		Name:        repoInfo.GetFullName(),
		Stars:       repoInfo.GetStargazersCount(),
		Forks:       repoInfo.GetForksCount(),
		LastUpdated: repoInfo.GetUpdatedAt().Time,
	}, nil
}

// FetchAndRankRepositories fetches the repositories in the Awesome Go list
func (ag *AwesomeGo) FetchAndRankRepositories(specificSection string, limit int) error {
	owner, repo := "avelino", "awesome-go"
	readmeMarkdown, err := ag.fetchReadmeMarkdown(owner, repo)
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

	// Fetch repository info
	// Until we reach the limit or run out of repositories
	cnt := 0
	for section, repos := range ag.repos {
		// Skip section if specificSection is set and the section is not contain the specificSection
		if specificSection != "" && !strings.Contains(specificSection, section) {
			continue
		}

		for i := range repos {
			// Stop if we reach the limit
			if limit > 0 && cnt >= limit {
				break
			}
			cnt++

			owner, name := extractRepoURLs(repos[i].URL)
			if owner != "" && name != "" {
				repoInfo, err := ag.fetchRepoInfo(owner, name)
				if err == nil {
					repos[i].Stars = repoInfo.Stars
					repos[i].Forks = repoInfo.Forks
					repos[i].LastUpdated = repoInfo.LastUpdated
				}
			}
			progressBar.Increment() // Update progress bar
		}
	}

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
