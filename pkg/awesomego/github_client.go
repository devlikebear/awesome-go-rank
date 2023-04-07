package awesomego

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

// IGithubClient is an interface for the Github API
type IGithubClient interface {
	// FetchReadmeMarkdown fetches the README.md file of a given repository.
	FetchReadmeMarkdown(ctx context.Context, owner, repo string) (string, error)

	// FetchRepository fetches the repositories from the Github API
	FetchRepository(ctx context.Context, owner, repo string) (*Repository, error)
}

// GithubClient is a struct that represents a Github client.
type GithubClient struct {
	client *github.Client
}

// NewGithubClient creates a new GithubClient instance.
func NewGithubClient(token string) *GithubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &GithubClient{client: client}
}

// FetchRepository fetches the repositories from the Github API
func (gc *GithubClient) FetchRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	repoInfo, _, err := gc.client.Repositories.Get(ctx, owner, repo)
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

	readmeMarkdown, _ := ioutil.ReadAll(resp.Body)
	return string(readmeMarkdown), nil
}
