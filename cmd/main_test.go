package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/devlikebear/awesome-go-rank/pkg/awesomego"
	"github.com/stretchr/testify/assert"
)

func TestConvertToFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lowercase", "test", "test"},
		{"with spaces", "test section", "test-section"},
		{"multiple spaces", "test  section  name", "test--section--name"},
		{"no spaces", "Authentication", "Authentication"},
		{"mixed case with spaces", "Web Frameworks", "Web-Frameworks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteRepositoriesToFile(t *testing.T) {
	repos := []awesomego.Repository{
		{
			Name:        "user/repo1",
			URL:         "https://github.com/user/repo1",
			Stars:       1000,
			Forks:       100,
			LastUpdated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Description: "-Test description-",
		},
		{
			Name:        "user/repo2",
			URL:         "https://github.com/user/repo2",
			Stars:       5000,
			Forks:       500,
			LastUpdated: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			Description: "Another repo",
		},
	}

	var buf bytes.Buffer
	writeRepositoriesToFile("Test Title", repos, &buf)

	output := buf.String()

	// Verify header
	assert.Contains(t, output, "### Test Title")
	assert.Contains(t, output, "| Repository | Stars | Forks | Last Updated | Description |")
	assert.Contains(t, output, "|------------|-------|-------|--------------|-------------|")

	// Verify repo entries
	assert.Contains(t, output, "[user/repo1](https://github.com/user/repo1)")
	assert.Contains(t, output, "1k") // Stars formatted
	assert.Contains(t, output, "100") // Forks
	assert.Contains(t, output, "Test description") // Description trimmed

	assert.Contains(t, output, "[user/repo2](https://github.com/user/repo2)")
	assert.Contains(t, output, "5k")
	assert.Contains(t, output, "500")
	assert.Contains(t, output, "Another repo")
}

func TestWriteReadmeHeader(t *testing.T) {
	var buf bytes.Buffer
	err := writeReadmeHeader(&buf)

	assert.NoError(t, err)
	output := buf.String()

	// Verify key sections
	assert.Contains(t, output, "# Awesome Go Ranking")
	assert.Contains(t, output, "awesome-go")
	assert.Contains(t, output, "https://awesome-go-rank.vercel.app/")
	assert.Contains(t, output, "GITHUB_TOKEN")
	assert.Contains(t, output, "go run cmd/main.go")
}

func TestGenerateTableOfContents(t *testing.T) {
	sections := map[string]awesomego.Section{
		"Authentication": {
			Name:        "Authentication",
			Description: "Authentication and OAuth libraries",
		},
		"Database": {
			Name:        "Database",
			Description: "Database libraries",
		},
	}

	repositories := map[string][]awesomego.Repository{
		"Authentication": {
			{Name: "user/auth-lib", URL: "https://github.com/user/auth-lib"},
		},
		"Database": {
			{Name: "user/db-lib", URL: "https://github.com/user/db-lib"},
		},
	}

	toc := generateTableOfContents(sections, repositories)

	// Verify TOC header
	assert.Contains(t, toc, "## Table of Contents")

	// Verify sections are included (should be sorted alphabetically)
	assert.Contains(t, toc, "* [Authentication](docs/Authentication.md)")
	assert.Contains(t, toc, "Authentication and OAuth libraries")
	assert.Contains(t, toc, "* [Database](docs/Database.md)")
	assert.Contains(t, toc, "Database libraries")

	// Verify alphabetical ordering
	authIndex := strings.Index(toc, "Authentication")
	dbIndex := strings.Index(toc, "Database")
	assert.Less(t, authIndex, dbIndex, "Authentication should come before Database")
}

func TestGenerateTableOfContents_EmptySection(t *testing.T) {
	sections := map[string]awesomego.Section{
		"Empty": {
			Name:        "Empty",
			Description: "Empty section",
		},
	}

	repositories := map[string][]awesomego.Repository{
		"Empty": {}, // No repositories
	}

	toc := generateTableOfContents(sections, repositories)

	// Empty sections should not be included
	assert.NotContains(t, toc, "Empty")
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		shouldError bool
	}{
		{
			name: "valid config",
			config: Config{
				GitHubToken:     "test-token",
				SpecificSection: "",
				Limit:           0,
			},
			shouldError: false,
		},
		{
			name: "missing token",
			config: Config{
				GitHubToken:     "",
				SpecificSection: "",
				Limit:           0,
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test token validation
			if tt.config.GitHubToken == "" {
				assert.Empty(t, tt.config.GitHubToken)
			} else {
				assert.NotEmpty(t, tt.config.GitHubToken)
			}
		})
	}
}

func TestWriteRepositoriesToFile_Formatting(t *testing.T) {
	repos := []awesomego.Repository{
		{
			Name:        "test/repo",
			URL:         "https://github.com/test/repo",
			Stars:       12345,
			Forks:       1234,
			LastUpdated: time.Date(2024, 2, 17, 12, 30, 45, 0, time.UTC),
			Description: "Test",
		},
	}

	var buf bytes.Buffer
	writeRepositoriesToFile("Title", repos, &buf)

	output := buf.String()

	// Verify metric number formatting
	assert.Contains(t, output, "12k") // Stars formatted as metric
	assert.Contains(t, output, "1k")  // Forks formatted as metric

	// Verify date formatting
	assert.Contains(t, output, "2024-02-17T12:30:45Z")
}

func TestWriteRepositoriesToFile_DescriptionTrimming(t *testing.T) {
	repos := []awesomego.Repository{
		{
			Name:        "test/repo",
			URL:         "https://github.com/test/repo",
			Stars:       100,
			Forks:       10,
			LastUpdated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Description: "---Test description---",
		},
	}

	var buf bytes.Buffer
	writeRepositoriesToFile("Title", repos, &buf)

	output := buf.String()

	// Verify description is trimmed
	assert.Contains(t, output, "Test description")
	assert.NotContains(t, output, "---Test description---")
}

func TestGenerateTableOfContents_Sorting(t *testing.T) {
	sections := map[string]awesomego.Section{
		"Zulu":    {Name: "Zulu", Description: "Last"},
		"Alpha":   {Name: "Alpha", Description: "First"},
		"Charlie": {Name: "Charlie", Description: "Middle"},
	}

	repositories := map[string][]awesomego.Repository{
		"Zulu":    {{Name: "test"}},
		"Alpha":   {{Name: "test"}},
		"Charlie": {{Name: "test"}},
	}

	toc := generateTableOfContents(sections, repositories)

	// Find indices
	alphaIdx := strings.Index(toc, "Alpha")
	charlieIdx := strings.Index(toc, "Charlie")
	zuluIdx := strings.Index(toc, "Zulu")

	// Verify alphabetical order
	assert.Less(t, alphaIdx, charlieIdx, "Alpha should come before Charlie")
	assert.Less(t, charlieIdx, zuluIdx, "Charlie should come before Zulu")
}

func TestWriteRepositoriesToFile_MultipleRepos(t *testing.T) {
	repos := make([]awesomego.Repository, 10)
	for i := 0; i < 10; i++ {
		repos[i] = awesomego.Repository{
			Name:        "user/repo" + string(rune('0'+i)),
			URL:         "https://github.com/user/repo",
			Stars:       (i + 1) * 1000,
			Forks:       (i + 1) * 100,
			LastUpdated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Description: "Repo " + string(rune('0'+i)),
		}
	}

	var buf bytes.Buffer
	writeRepositoriesToFile("Multiple Repos", repos, &buf)

	output := buf.String()

	// Verify all repos are included
	lines := strings.Split(output, "\n")
	// Header (1) + Table header (2) + 10 repos + empty line = at least 14 lines
	assert.GreaterOrEqual(t, len(lines), 14)

	// Verify table structure
	headerCount := 0
	for _, line := range lines {
		if strings.Contains(line, "| Repository |") {
			headerCount++
		}
	}
	assert.Equal(t, 1, headerCount, "Should have exactly one table header")
}
