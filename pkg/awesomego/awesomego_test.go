package awesomego_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/devlikebear/awesome-go-rank/pkg/awesomego"
	"github.com/devlikebear/awesome-go-rank/pkg/config"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// Create test config
	cfg := config.Default()
	cfg.GitHub.Token = "test-token"

	// Create a new AwesomeGo instance
	s.ag = awesomego.NewAwesomeGo(mockClient, cfg)
}

// TestAwesomeGo_FetchAndRankRepositories_ValidSpecificSectionAndLimit tests the FetchAndRankRepositories method with a valid specific section and a limit
func (s *AwesomeGoTestSuite) TestAwesomeGo_FetchAndRankRepositories_ValidSpecificSectionAndLimit() {
	err := s.ag.FetchAndRankRepositories("Audio and Music", 10)
	s.NoError(err)

	repos := s.ag.Repositories()
	s.Equal(2, len(repos)) // Contents and Audio and Music
	s.Equal(len(repos["Audio and Music"]), 9)
	s.Equal("mewkiz/flac", repos["Audio and Music"][0].Name)
	s.Equal("- Native Go FLAC encoder/decoder with support for FLAC streams.", repos["Audio and Music"][0].Description)

	sections := s.ag.Sections()
	s.Equal(2, len(sections)) // Contents (no description) and Audio and Music
	s.Equal("Audio and Music", sections["Audio and Music"].Name)
	s.Equal("Libraries for manipulating audio.", sections["Audio and Music"].Description)
	s.Equal("Contents", sections["Contents"].Name)
	s.Equal("", sections["Contents"].Description) // Contents has no description
}

func (s *AwesomeGoTestSuite) TestAwesomeGo_FetchAndRankRepositories_ExactSectionMatch() {
	readme := `# Awesome Go

## Data
- [data](https://github.com/example/data) - Data package.

## Database
- [database](https://github.com/example/database) - Database package.`
	mockClient := awesomego.NewMockGithubClient(&awesomego.Repository{Stars: 1}, readme)
	cfg := config.Default()
	cfg.RateLimit.MinInterval = 0
	ag := awesomego.NewAwesomeGo(mockClient, cfg)

	s.NoError(ag.FetchAndRankRepositories("Database", 0))
	s.Len(ag.Repositories()["Database"], 1)
	s.Equal(1, ag.Repositories()["Database"][0].Stars)
	s.Empty(ag.Repositories()["Data"])
}

func TestFetchAndRankRepositoriesFailureThreshold(t *testing.T) {
	tests := []struct {
		name          string
		failures      int
		wantErr       bool
		wantCollected int
	}{
		{name: "nine percent succeeds", failures: 9, wantCollected: 91},
		{name: "eleven percent fails", failures: 11, wantErr: true, wantCollected: 89},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			var lines strings.Builder
			lines.WriteString("## Database\n")
			for i := 0; i < 100; i++ {
				fmt.Fprintf(&lines, "- [repo-%d](https://github.com/example/repo-%d) - Test.\n", i, i)
			}

			readmeResponse := map[string]any{
				"name":     "README.md",
				"path":     "README.md",
				"encoding": "base64",
				"content":  base64.StdEncoding.EncodeToString([]byte(lines.String())),
			}
			httpmock.RegisterResponder(http.MethodGet,
				"https://api.github.com/repos/source/list/readme",
				httpmock.NewJsonResponderOrPanic(http.StatusOK, readmeResponse))

			for i := 0; i < 100; i++ {
				url := fmt.Sprintf("https://api.github.com/repos/example/repo-%d", i)
				if i < tt.failures {
					httpmock.RegisterResponder(http.MethodGet, url,
						httpmock.NewJsonResponderOrPanic(http.StatusInternalServerError, map[string]string{"message": "failed"}))
					continue
				}
				httpmock.RegisterResponder(http.MethodGet, url,
					httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
						"full_name":        fmt.Sprintf("example/repo-%d", i),
						"stargazers_count": i + 1,
						"forks_count":      i,
						"updated_at":       "2026-07-12T00:00:00Z",
					}))
			}

			cfg := config.Default()
			cfg.GitHub.Owner = "source"
			cfg.GitHub.Repository = "list"
			cfg.RateLimit.MaxRetries = 1
			cfg.RateLimit.MinInterval = 0
			client := awesomego.NewGithubClient("")
			ag := awesomego.NewAwesomeGo(client, cfg)

			err := ag.FetchAndRankRepositories("Database", 0)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Len(t, ag.Repositories()["Database"], tt.wantCollected)
		})
	}
}

func TestGithubClientReadmeRequestUsesRepositoryAPI(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	content := "# README from the default branch"
	httpmock.RegisterResponder(http.MethodGet,
		"https://api.github.com/repos/example/project/readme",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, map[string]any{
			"encoding": "base64",
			"content":  base64.StdEncoding.EncodeToString([]byte(content)),
		}))

	client := awesomego.NewGithubClient("")
	got, err := client.FetchReadmeMarkdown(context.Background(), "example", "project")
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

// TestAwesomeGoTestSuite runs the test suite
func TestAwesomeGoTestSuite(t *testing.T) {
	suite.Run(t, new(AwesomeGoTestSuite))
}
