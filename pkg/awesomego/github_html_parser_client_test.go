package awesomego

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type GithubHtmlParserClientTestSuite struct {
	suite.Suite
	client *GithubHtmlParserClient
}

func (s *GithubHtmlParserClientTestSuite) SetupSuite() {
	// Create a new Github client
	s.client = NewGithubHtmlParserClient()
}

func (s *GithubHtmlParserClientTestSuite) TestGithubHtmlParserClient_FetchRepository_ValidOwnerAndRepo() {
	// Fetch the repositories
	repo, err := s.client.FetchRepository(context.Background(), "avelino", "awesome-go")
	s.NoError(err)
	s.Equal("avelino/awesome-go", repo.Name)
	s.T().Log(repo)
}

func TestGithubHtmlParserClientSuite(t *testing.T) {
	suite.Run(t, new(GithubHtmlParserClientTestSuite))
}
