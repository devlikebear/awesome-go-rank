package awesomego

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestFetchHTMLRejectsNonGitHubURL(t *testing.T) {
	if _, err := fetchHTML("https://example.com/owner/repo"); err == nil {
		t.Fatal("fetchHTML accepted a non-GitHub URL")
	}
}

type GithubHtmlParserClientTestSuite struct {
	suite.Suite
	client *GithubHtmlParserClient
}

func (s *GithubHtmlParserClientTestSuite) SetupSuite() {
	// Create a new Github client
	s.client = NewGithubHtmlParserClient()
}

func (s *GithubHtmlParserClientTestSuite) TestGithubHtmlParserClient_FetchRepository_ValidOwnerAndRepo() {
	// Scrapes live github.com markup, which changes without notice; opt in explicitly.
	if os.Getenv("AWESOME_GO_RANK_LIVE_TESTS") == "" {
		s.T().Skip("set AWESOME_GO_RANK_LIVE_TESTS=1 to run tests against live github.com")
	}
	// Fetch the repositories
	repo, err := s.client.FetchRepository(context.Background(), "avelino", "awesome-go")
	s.NoError(err)
	s.Equal("avelino/awesome-go", repo.Name)
	s.T().Log(repo)
}

func TestGithubHtmlParserClientSuite(t *testing.T) {
	suite.Run(t, new(GithubHtmlParserClientTestSuite))
}
