package awesomego

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/devlikebear/awesome-go-rank/pkg/logger"
	"go.uber.org/zap"
)

// RepoData represents repository data for JSON export
type RepoData struct {
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Stars       int       `json:"stars"`
	Forks       int       `json:"forks"`
	LastUpdated time.Time `json:"lastUpdated"`
	Description string    `json:"description"`
}

// SectionData represents a section with its repositories
type SectionData struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	RepoCount   int        `json:"repoCount"`
	Repos       []RepoData `json:"repos"`
}

// JSONOutput represents the complete JSON output structure
type JSONOutput struct {
	UpdatedAt    time.Time       `json:"updatedAt"`
	TotalRepos   int             `json:"totalRepos"`
	TotalSections int            `json:"totalSections"`
	Sections     []SectionData   `json:"sections"`
	Metadata     OutputMetadata  `json:"metadata"`
}

// OutputMetadata contains metadata about the export
type OutputMetadata struct {
	SourceOwner      string `json:"sourceOwner"`
	SourceRepo       string `json:"sourceRepo"`
	SourceURL        string `json:"sourceUrl"`
	GeneratedBy      string `json:"generatedBy"`
	Version          string `json:"version"`
}

// JSONExporter handles exporting repository data to JSON
type JSONExporter struct {
	repos    map[string][]Repository
	sections map[string]Section
}

// NewJSONExporter creates a new JSONExporter
func NewJSONExporter(repos map[string][]Repository, sections map[string]Section) *JSONExporter {
	return &JSONExporter{
		repos:    repos,
		sections: sections,
	}
}

// Export exports the data to a JSON file
func (je *JSONExporter) Export(outputPath, sourceOwner, sourceRepo string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build JSON output
	output := je.buildJSONOutput(sourceOwner, sourceRepo)

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Encode JSON with indentation
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	logger.Info("JSON export completed",
		zap.String("output_path", outputPath),
		zap.Int("total_repos", output.TotalRepos),
		zap.Int("total_sections", output.TotalSections))

	return nil
}

// buildJSONOutput builds the complete JSON output structure
func (je *JSONExporter) buildJSONOutput(sourceOwner, sourceRepo string) JSONOutput {
	var sections []SectionData
	totalRepos := 0

	// Get sorted section names
	sectionNames := make([]string, 0, len(je.repos))
	for sectionName := range je.repos {
		sectionNames = append(sectionNames, sectionName)
	}
	sort.Strings(sectionNames)

	// Build sections data
	for _, sectionName := range sectionNames {
		repos := je.repos[sectionName]
		section := je.sections[sectionName]

		if len(repos) == 0 {
			continue
		}

		// Convert repositories to RepoData
		repoData := make([]RepoData, 0, len(repos))
		for _, repo := range repos {
			repoData = append(repoData, RepoData{
				Name:        repo.Name,
				URL:         repo.URL,
				Stars:       repo.Stars,
				Forks:       repo.Forks,
				LastUpdated: repo.LastUpdated,
				Description: repo.Description,
			})
		}

		sections = append(sections, SectionData{
			Name:        section.Name,
			Description: section.Description,
			RepoCount:   len(repos),
			Repos:       repoData,
		})

		totalRepos += len(repos)
	}

	return JSONOutput{
		UpdatedAt:     time.Now(),
		TotalRepos:    totalRepos,
		TotalSections: len(sections),
		Sections:      sections,
		Metadata: OutputMetadata{
			SourceOwner: sourceOwner,
			SourceRepo:  sourceRepo,
			SourceURL:   fmt.Sprintf("https://github.com/%s/%s", sourceOwner, sourceRepo),
			GeneratedBy: "awesome-go-rank",
			Version:     "1.0.0",
		},
	}
}

// ExportStats represents statistics about the exported data
type ExportStats struct {
	TotalRepos    int
	TotalSections int
	TopStarred    []RepoData
	RecentlyUpdated []RepoData
}

// GetStats returns statistics about the repositories
func (je *JSONExporter) GetStats(topN int) ExportStats {
	allRepos := make([]RepoData, 0)

	// Collect all repositories
	for _, repos := range je.repos {
		for _, repo := range repos {
			allRepos = append(allRepos, RepoData{
				Name:        repo.Name,
				URL:         repo.URL,
				Stars:       repo.Stars,
				Forks:       repo.Forks,
				LastUpdated: repo.LastUpdated,
				Description: repo.Description,
			})
		}
	}

	// Sort by stars for top starred
	topStarred := make([]RepoData, len(allRepos))
	copy(topStarred, allRepos)
	sort.Slice(topStarred, func(i, j int) bool {
		return topStarred[i].Stars > topStarred[j].Stars
	})
	if len(topStarred) > topN {
		topStarred = topStarred[:topN]
	}

	// Sort by last updated for recently updated
	recentlyUpdated := make([]RepoData, len(allRepos))
	copy(recentlyUpdated, allRepos)
	sort.Slice(recentlyUpdated, func(i, j int) bool {
		return recentlyUpdated[i].LastUpdated.After(recentlyUpdated[j].LastUpdated)
	})
	if len(recentlyUpdated) > topN {
		recentlyUpdated = recentlyUpdated[:topN]
	}

	return ExportStats{
		TotalRepos:      len(allRepos),
		TotalSections:   len(je.sections),
		TopStarred:      topStarred,
		RecentlyUpdated: recentlyUpdated,
	}
}
