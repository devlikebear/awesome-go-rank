package awesomego

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

// GithubClientTestSuite is the test suite for the GithubClient
type GithubClientTestSuite struct {
	suite.Suite
	client *GithubClient
}

// SetupSuite sets up the test suite
func (s *GithubClientTestSuite) SetupSuite() {
	// Get the Github token from the environment
	err := godotenv.Load()
	s.NoError(err)
	token := os.Getenv("GITHUB_TOKEN")

	// Create a new Github client
	s.client = NewGithubClient(token)
}

// TestGithubClient_FetchRepository_ValidOwnerAndRepo tests the FetchRepository method with a valid owner and repo
func (s *GithubClientTestSuite) TestGithubClient_FetchRepository_ValidOwnerAndRepo() {
	// Fetch the repositories
	repo, err := s.client.FetchRepository(context.Background(), "avelino", "awesome-go")
	s.NoError(err)
	s.Equal("avelino/awesome-go", repo.Name)
}

// TestGithubClient_FetchRepository_InvalidOwner tests the FetchRepository method with an invalid owner
func (s *GithubClientTestSuite) TestGithubClient_FetchRepository_InvalidOwner() {
	// Fetch the repositories
	repo, err := s.client.FetchRepository(context.Background(), "invalid", "awesome-go")
	s.Error(err)
	s.Nil(repo)
}

// TestGithubClient_FetchRepository_InvalidRepo tests the FetchRepository method with an invalid repo
func (s *GithubClientTestSuite) TestGithubClient_FetchRepository_InvalidRepo() {
	// Fetch the repositories
	repo, err := s.client.FetchRepository(context.Background(), "avelino", "invalid")
	s.Error(err)
	s.Nil(repo)
}

// TestGithubClient_FetchReadmeMarkdown_ValidOwnerAndRepo tests the FetchReadmeMarkdown method with a valid owner and repo
func (s *GithubClientTestSuite) TestGithubClient_FetchReadmeMarkdown_ValidOwnerAndRepo() {
	// Fetch the repositories
	markdown, err := s.client.FetchReadmeMarkdown(context.Background(), "avelino", "awesome-go")
	s.NoError(err)
	s.NotEmpty(markdown)
}

// TestGithubClient_FetchReadmeMarkdown_InvalidOwner tests the FetchReadmeMarkdown method with an invalid owner
func (s *GithubClientTestSuite) TestGithubClient_FetchReadmeMarkdown_InvalidOwner() {
	// Fetch the repositories
	markdown, err := s.client.FetchReadmeMarkdown(context.Background(), "invalid", "awesome-go")
	s.Error(err)
	s.Empty(markdown)
}

// TestGithubClient_FetchReadmeMarkdown_InvalidRepo tests the FetchReadmeMarkdown method with an invalid repo
func (s *GithubClientTestSuite) TestGithubClient_FetchReadmeMarkdown_InvalidRepo() {
	// Fetch the repositories
	markdown, err := s.client.FetchReadmeMarkdown(context.Background(), "avelino", "invalid")
	s.Error(err)
	s.Empty(markdown)
}

// TestGithubClientTestSuite runs the test suite
func TestGithubClientTestSuite(t *testing.T) {
	suite.Run(t, new(GithubClientTestSuite))
}
