package awesomego

import "context"

// MockGithubClient is a mock Github client
type MockGithubClient struct {
	result *Repository
	readme string
}

// NewMockGithubClient creates a new MockGithubClient instance
func NewMockGithubClient(result *Repository, readme string) *MockGithubClient {
	return &MockGithubClient{result: result, readme: readme}
}

// FetchRepository fetches the repositories from the Github API
func (gc *MockGithubClient) FetchRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	return gc.result, nil
}

// FetchReadmeMarkdown fetches the README.md file of a given repository.
func (gc *MockGithubClient) FetchReadmeMarkdown(ctx context.Context, owner, repo string) (string, error) {
	return gc.readme, nil
}
