package awesomego

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONExporter(t *testing.T) {
	repos := map[string][]Repository{
		"Test": {{Name: "test/repo"}},
	}
	sections := map[string]Section{
		"Test": {Name: "Test", Description: "Test section"},
	}

	exporter := NewJSONExporter(repos, sections)
	assert.NotNil(t, exporter)
	assert.Equal(t, 1, len(exporter.repos))
	assert.Equal(t, 1, len(exporter.sections))
}

func TestJSONExporter_Export(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output", "repos.json")

	repos := map[string][]Repository{
		"Authentication": {
			{
				Name:        "user/auth-lib",
				URL:         "https://github.com/user/auth-lib",
				Stars:       1000,
				Forks:       100,
				LastUpdated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Description: "Auth library",
			},
		},
		"Database": {
			{
				Name:        "user/db-lib",
				URL:         "https://github.com/user/db-lib",
				Stars:       2000,
				Forks:       200,
				LastUpdated: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				Description: "Database library",
			},
		},
	}

	sections := map[string]Section{
		"Authentication": {Name: "Authentication", Description: "Auth libraries"},
		"Database":       {Name: "Database", Description: "DB libraries"},
	}

	exporter := NewJSONExporter(repos, sections)
	err := exporter.Export(outputPath, "avelino", "awesome-go")
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Read and parse JSON
	file, err := os.Open(outputPath)
	require.NoError(t, err)
	defer file.Close()

	var output JSONOutput
	err = json.NewDecoder(file).Decode(&output)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, 2, output.TotalRepos)
	assert.Equal(t, 2, output.TotalSections)
	assert.Equal(t, "avelino", output.Metadata.SourceOwner)
	assert.Equal(t, "awesome-go", output.Metadata.SourceRepo)
	assert.Equal(t, "https://github.com/avelino/awesome-go", output.Metadata.SourceURL)

	// Verify sections are sorted alphabetically
	assert.Equal(t, "Authentication", output.Sections[0].Name)
	assert.Equal(t, "Database", output.Sections[1].Name)

	// Verify repo data
	assert.Equal(t, 1, output.Sections[0].RepoCount)
	assert.Equal(t, "user/auth-lib", output.Sections[0].Repos[0].Name)
	assert.Equal(t, 1000, output.Sections[0].Repos[0].Stars)
}

func TestJSONExporter_Export_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "nested", "dir", "repos.json")

	repos := map[string][]Repository{
		"Test": {{Name: "test/repo"}},
	}
	sections := map[string]Section{
		"Test": {Name: "Test"},
	}

	exporter := NewJSONExporter(repos, sections)
	err := exporter.Export(outputPath, "owner", "repo")
	require.NoError(t, err)

	// Verify nested directory was created
	_, err = os.Stat(filepath.Dir(outputPath))
	require.NoError(t, err)
}

func TestJSONExporter_Export_EmptySection(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "repos.json")

	repos := map[string][]Repository{
		"Test":  {{Name: "test/repo"}},
		"Empty": {}, // Empty section
	}
	sections := map[string]Section{
		"Test":  {Name: "Test"},
		"Empty": {Name: "Empty"},
	}

	exporter := NewJSONExporter(repos, sections)
	err := exporter.Export(outputPath, "owner", "repo")
	require.NoError(t, err)

	// Read output
	file, err := os.Open(outputPath)
	require.NoError(t, err)
	defer file.Close()

	var output JSONOutput
	json.NewDecoder(file).Decode(&output)

	// Empty sections should be excluded
	assert.Equal(t, 1, len(output.Sections))
	assert.Equal(t, "Test", output.Sections[0].Name)
}

func TestJSONExporter_GetStats(t *testing.T) {
	repos := map[string][]Repository{
		"Section1": {
			{Name: "user/high-stars", Stars: 10000, LastUpdated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			{Name: "user/low-stars", Stars: 100, LastUpdated: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)},
		},
		"Section2": {
			{Name: "user/medium-stars", Stars: 5000, LastUpdated: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
	sections := map[string]Section{
		"Section1": {Name: "Section1"},
		"Section2": {Name: "Section2"},
	}

	exporter := NewJSONExporter(repos, sections)
	stats := exporter.GetStats(2)

	assert.Equal(t, 3, stats.TotalRepos)
	assert.Equal(t, 2, stats.TotalSections)

	// Top starred should be sorted by stars
	assert.Equal(t, 2, len(stats.TopStarred))
	assert.Equal(t, "user/high-stars", stats.TopStarred[0].Name)
	assert.Equal(t, "user/medium-stars", stats.TopStarred[1].Name)

	// Recently updated should be sorted by last updated
	assert.Equal(t, 2, len(stats.RecentlyUpdated))
	assert.Equal(t, "user/low-stars", stats.RecentlyUpdated[0].Name)
	assert.Equal(t, "user/medium-stars", stats.RecentlyUpdated[1].Name)
}

func TestJSONExporter_GetStats_LimitTopN(t *testing.T) {
	repos := map[string][]Repository{
		"Section": {
			{Name: "repo1", Stars: 1000},
			{Name: "repo2", Stars: 2000},
			{Name: "repo3", Stars: 3000},
			{Name: "repo4", Stars: 4000},
			{Name: "repo5", Stars: 5000},
		},
	}
	sections := map[string]Section{
		"Section": {Name: "Section"},
	}

	exporter := NewJSONExporter(repos, sections)
	stats := exporter.GetStats(3)

	// Should only return top 3
	assert.Equal(t, 3, len(stats.TopStarred))
	assert.Equal(t, 5000, stats.TopStarred[0].Stars)
	assert.Equal(t, 4000, stats.TopStarred[1].Stars)
	assert.Equal(t, 3000, stats.TopStarred[2].Stars)
}

func TestJSONOutput_JSONFormat(t *testing.T) {
	repos := map[string][]Repository{
		"Test": {
			{
				Name:        "test/repo",
				URL:         "https://github.com/test/repo",
				Stars:       100,
				Forks:       10,
				LastUpdated: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Description: "Test description",
			},
		},
	}
	sections := map[string]Section{
		"Test": {Name: "Test", Description: "Test section"},
	}

	exporter := NewJSONExporter(repos, sections)
	output := exporter.buildJSONOutput("owner", "repo")

	// Marshal to JSON and verify structure
	jsonData, err := json.MarshalIndent(output, "", "  ")
	require.NoError(t, err)

	// Verify JSON contains expected fields
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "updatedAt")
	assert.Contains(t, jsonStr, "totalRepos")
	assert.Contains(t, jsonStr, "totalSections")
	assert.Contains(t, jsonStr, "sections")
	assert.Contains(t, jsonStr, "metadata")
	assert.Contains(t, jsonStr, "sourceOwner")
	assert.Contains(t, jsonStr, "generatedBy")
}
