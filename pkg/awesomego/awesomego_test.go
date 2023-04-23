package awesomego_test

import (
	"testing"

	"github.com/devlikebear/awesome-go-rank/pkg/awesomego"
	"github.com/stretchr/testify/suite"
)

// AwesomeGoTestSuite is the test suite for the AwesomeGo package
type AwesomeGoTestSuite struct {
	suite.Suite
	ag *awesomego.AwesomeGo
}

// SetupSuite sets up the test suite
func (s *AwesomeGoTestSuite) SetupSuite() {
	// Create a new AwesomeGo instance with a mock Github client
	mockClient := awesomego.NewMockGithubClient(&awesomego.Repository{
		Name:  "avelino/awesome-go",
		URL:   "https://github.com/avelino/awesome-go",
		Stars: 0,
		Forks: 0,
	}, cReadme)

	// Create a new AwesomeGo instance
	s.ag = awesomego.NewAwesomeGo("", mockClient)
}

// TestAwesomeGo_FetchAndRankRepositories_ValidSpecificSectionAndLimit tests the FetchAndRankRepositories method with a valid specific section and a limit
func (s *AwesomeGoTestSuite) TestAwesomeGo_FetchAndRankRepositories_ValidSpecificSectionAndLimit() {
	err := s.ag.FetchAndRankRepositories("Audio and Music", 10)
	s.NoError(err)

	repos := s.ag.Repositories()
	s.Equal(2, len(repos)) // Contents and Audio and Music
	s.Equal(len(repos["Audio and Music"]), 9)
	s.Equal("mewkiz/flac", repos["Audio and Music"][0].Name)
}

// TestAwesomeGoTestSuite runs the test suite
func TestAwesomeGoTestSuite(t *testing.T) {
	suite.Run(t, new(AwesomeGoTestSuite))
}
