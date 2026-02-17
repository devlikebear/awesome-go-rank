package awesomego

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
)

// GithubClientTestSuite is the test suite for the GithubClient
type GithubClientTestSuite struct {
	suite.Suite
	client *GithubClient
}

// SetupSuite sets up the test suite
func (s *GithubClientTestSuite) SetupSuite() {
	// Create a new Github client with empty token for testing
	s.client = NewGithubClient("")

	// Activate httpmock
	httpmock.Activate()
}

// TearDownSuite cleans up after the test suite
func (s *GithubClientTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

// SetupTest runs before each test
func (s *GithubClientTestSuite) SetupTest() {
	// Reset mock state before each test
	httpmock.Reset()
}

// TestGithubClient_FetchRepository_ValidOwnerAndRepo tests the FetchRepository method with a valid owner and repo
func (s *GithubClientTestSuite) TestGithubClient_FetchRepository_ValidOwnerAndRepo() {
	// Mock the GitHub API response
	httpmock.RegisterResponder("GET",
		"https://api.github.com/repos/avelino/awesome-go",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"full_name":        "avelino/awesome-go",
			"stargazers_count": 12345,
			"forks_count":      1234,
			"updated_at":       "2024-01-01T00:00:00Z",
		}))

	// Fetch the repositories
	repo, err := s.client.FetchRepository(context.Background(), "avelino", "awesome-go")
	s.NoError(err)
	s.Equal("avelino/awesome-go", repo.Name)
	s.Equal(12345, repo.Stars)
	s.Equal(1234, repo.Forks)

	// Verify the expected number of calls
	s.Equal(1, httpmock.GetTotalCallCount())
}

// TestGithubClient_FetchRepository_InvalidOwner tests the FetchRepository method with an invalid owner
func (s *GithubClientTestSuite) TestGithubClient_FetchRepository_InvalidOwner() {
	// Mock 404 response for invalid owner
	httpmock.RegisterResponder("GET",
		"https://api.github.com/repos/invalid/awesome-go",
		httpmock.NewJsonResponderOrPanic(404, map[string]interface{}{
			"message": "Not Found",
		}))

	// Fetch the repositories
	repo, err := s.client.FetchRepository(context.Background(), "invalid", "awesome-go")
	s.Error(err)
	s.Nil(repo)
}

// TestGithubClient_FetchRepository_InvalidRepo tests the FetchRepository method with an invalid repo
func (s *GithubClientTestSuite) TestGithubClient_FetchRepository_InvalidRepo() {
	// Mock 404 response for invalid repo
	httpmock.RegisterResponder("GET",
		"https://api.github.com/repos/avelino/invalid",
		httpmock.NewJsonResponderOrPanic(404, map[string]interface{}{
			"message": "Not Found",
		}))

	// Fetch the repositories
	repo, err := s.client.FetchRepository(context.Background(), "avelino", "invalid")
	s.Error(err)
	s.Nil(repo)
}

// TestGithubClient_FetchReadmeMarkdown_ValidOwnerAndRepo tests the FetchReadmeMarkdown method with a valid owner and repo
func (s *GithubClientTestSuite) TestGithubClient_FetchReadmeMarkdown_ValidOwnerAndRepo() {
	mockMarkdown := `# Awesome Go

## Contents
- [Authentication](#authentication)

## Authentication
- [awesome-lib](https://github.com/user/awesome-lib) - Great library`

	// Mock the raw.githubusercontent.com response
	httpmock.RegisterResponder("GET",
		"https://raw.githubusercontent.com/avelino/awesome-go/main/README.md",
		httpmock.NewStringResponder(200, mockMarkdown))

	// Fetch the repositories
	markdown, err := s.client.FetchReadmeMarkdown(context.Background(), "avelino", "awesome-go")
	s.NoError(err)
	s.NotEmpty(markdown)
	s.Contains(markdown, "Awesome Go")
	s.Contains(markdown, "Authentication")
}

// TestGithubClient_FetchReadmeMarkdown_InvalidOwner tests the FetchReadmeMarkdown method with an invalid owner
func (s *GithubClientTestSuite) TestGithubClient_FetchReadmeMarkdown_InvalidOwner() {
	// Mock 404 response for invalid owner
	httpmock.RegisterResponder("GET",
		"https://raw.githubusercontent.com/invalid/awesome-go/main/README.md",
		httpmock.NewStringResponder(404, "404: Not Found"))

	// Fetch the repositories
	markdown, err := s.client.FetchReadmeMarkdown(context.Background(), "invalid", "awesome-go")
	s.Error(err)
	s.Empty(markdown)
	s.Contains(err.Error(), "failed to fetch README.md")
}

// TestGithubClient_FetchReadmeMarkdown_InvalidRepo tests the FetchReadmeMarkdown method with an invalid repo
func (s *GithubClientTestSuite) TestGithubClient_FetchReadmeMarkdown_InvalidRepo() {
	// Mock 404 response for invalid repo
	httpmock.RegisterResponder("GET",
		"https://raw.githubusercontent.com/avelino/invalid/main/README.md",
		httpmock.NewStringResponder(404, "404: Not Found"))

	// Fetch the repositories
	markdown, err := s.client.FetchReadmeMarkdown(context.Background(), "avelino", "invalid")
	s.Error(err)
	s.Empty(markdown)
}

// TestGithubClient_FetchRepository_NetworkError tests error handling for network failures
func (s *GithubClientTestSuite) TestGithubClient_FetchRepository_NetworkError() {
	// Mock network error
	httpmock.RegisterResponder("GET",
		"https://api.github.com/repos/test/test",
		httpmock.NewErrorResponder(http.ErrHandlerTimeout))

	repo, err := s.client.FetchRepository(context.Background(), "test", "test")
	s.Error(err)
	s.Nil(repo)
}

// TestGithubClient_ParseRepositoryResponse tests parsing of various GitHub API responses
func (s *GithubClientTestSuite) TestGithubClient_ParseRepositoryResponse() {
	updatedTime, _ := time.Parse(time.RFC3339, "2024-02-17T12:00:00Z")

	httpmock.RegisterResponder("GET",
		"https://api.github.com/repos/test/repo",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"full_name":        "test/repo",
			"stargazers_count": 999,
			"forks_count":      99,
			"updated_at":       "2024-02-17T12:00:00Z",
		}))

	repo, err := s.client.FetchRepository(context.Background(), "test", "repo")
	s.NoError(err)
	s.Equal("test/repo", repo.Name)
	s.Equal(999, repo.Stars)
	s.Equal(99, repo.Forks)
	s.Equal(updatedTime, repo.LastUpdated)
}

// TestGithubClientTestSuite runs the test suite
func TestGithubClientTestSuite(t *testing.T) {
	suite.Run(t, new(GithubClientTestSuite))
}
