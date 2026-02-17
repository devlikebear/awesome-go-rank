package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/devlikebear/awesome-go-rank/pkg/awesomego"
	"github.com/devlikebear/awesome-go-rank/pkg/config"
	"github.com/devlikebear/awesome-go-rank/pkg/stringutil"
	"github.com/spf13/cobra"
)

// Config holds the application configuration
type Config struct {
	GitHubToken     string
	SpecificSection string
	Limit           int
	OutputDir       string
}

// printRepositories prints repositories to stdout
func printRepositories(title string, repositories []awesomego.Repository) {
	fmt.Println(title)
	for _, repo := range repositories {
		fmt.Printf("%s (Stars: %s, Forks: %s, Last updated: %v, Desc: %s)\n", repo.Name,
			stringutil.FormatMetricNumber(repo.Stars),
			stringutil.FormatMetricNumber(repo.Forks),
			repo.LastUpdated.Format("2006-01-02T15:04:05Z"),
			repo.Description,
		)
	}
}

// writeRepositoriesToFile writes a ranked list of repositories to a file
func writeRepositoriesToFile(title string, repositories []awesomego.Repository,
	file io.Writer) {
	fmt.Fprintf(file, "### %s\n\n", title)
	fmt.Fprintf(file, "| Repository | Stars | Forks | Last Updated | Description | \n")
	fmt.Fprintf(file, "|------------|-------|-------|--------------|-------------|\n")
	for _, repo := range repositories {
		fmt.Fprintf(file, "| [%s](%s) | %s | %s | %v | %s |\n", repo.Name,
			repo.URL,
			stringutil.FormatMetricNumber(repo.Stars),
			stringutil.FormatMetricNumber(repo.Forks),
			repo.LastUpdated.Format("2006-01-02T15:04:05Z"),
			strings.Trim(repo.Description, "-"))
	}
	fmt.Fprintf(file, "\n")
}

// generateTableOfContents creates the table of contents section
func generateTableOfContents(sections map[string]awesomego.Section,
	repositories map[string][]awesomego.Repository) string {
	var sb strings.Builder
	sb.WriteString("## Table of Contents\n\n")

	// Extract and sort sections
	sectionNames := make([]string, 0, len(repositories))
	for section := range repositories {
		sectionNames = append(sectionNames, section)
	}
	sort.Strings(sectionNames)

	// Generate TOC entries
	for _, sectionName := range sectionNames {
		section := sections[sectionName]
		repo := repositories[section.Name]
		if len(repo) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("* [%s](docs/%s.md)<br/>%s\n",
			section.Name, convertToFilename(section.Name), section.Description))
	}

	return sb.String()
}

// writeReadmeHeader writes the header section of README.md
func writeReadmeHeader(w io.Writer) error {
	_, err := w.Write([]byte(`# Awesome Go Ranking

This is a ranking of GitHub repositories from
 [awesome-go](https://github.com/avelino/awesome-go)
 by Stars, Forks and Last Updated.

## How to use

1. Clone this repository
1. Create a GitHub personal access token with ` + "`public_repo`" + ` scope
1. Set the token to the ` + "`GITHUB_TOKEN`" + ` environment variable
1. Install Go
1. Install dependencies with ` + "`go mod tidy`" + `
1. Run ` + "`go run cmd/main.go`" + `
1. Check the results in README.md


`))
	return err
}

// runRanking executes the main ranking logic
func runRanking(cmdConfig Config) error {
	// Load configuration from environment
	cfg, err := config.FromEnv()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Initialize client and fetch repositories
	client := awesomego.NewGithubClient(cfg.GitHub.Token)
	ag := awesomego.NewAwesomeGo(client, cfg)
	if err := ag.FetchAndRankRepositories(cmdConfig.SpecificSection, cmdConfig.Limit); err != nil {
		return fmt.Errorf("failed to fetch repositories: %w", err)
	}

	// Create output directory if it doesn't exist
	docsPath := cfg.GetDocsPath()
	if err := os.MkdirAll(docsPath, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Write README.md
	if err := writeReadme(ag, cmdConfig.SpecificSection); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	// Write section files
	if err := writeSectionFiles(ag, cmdConfig.SpecificSection); err != nil {
		return fmt.Errorf("failed to write section files: %w", err)
	}

	// Export JSON data
	if err := exportJSON(ag, cfg); err != nil {
		return fmt.Errorf("failed to export JSON: %w", err)
	}

	return nil
}

// exportJSON exports repository data to JSON format
func exportJSON(ag *awesomego.AwesomeGo, cfg *config.Config) error {
	exporter := awesomego.NewJSONExporter(ag.Repositories(), ag.Sections())

	// Export to public/data/repos.json
	outputPath := "public/data/repos.json"
	if err := exporter.Export(outputPath, cfg.GitHub.Owner, cfg.GitHub.Repository); err != nil {
		return err
	}

	return nil
}

// writeReadme writes the main README.md file
func writeReadme(ag *awesomego.AwesomeGo, specificSection string) error {
	outputFile, err := os.Create("README.md")
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Write header
	if err := writeReadmeHeader(outputFile); err != nil {
		return err
	}

	// Write table of contents
	toc := generateTableOfContents(ag.Sections(), ag.Repositories())
	if _, err := outputFile.WriteString(toc); err != nil {
		return err
	}

	return nil
}

// writeSectionFiles writes individual section markdown files
func writeSectionFiles(ag *awesomego.AwesomeGo, specificSection string) error {
	repositories := ag.Repositories()
	sections := ag.Sections()

	// Sort section names
	sectionNames := make([]string, 0, len(repositories))
	for section := range repositories {
		sectionNames = append(sectionNames, section)
	}
	sort.Strings(sectionNames)

	// Write each section file
	for _, sectionName := range sectionNames {
		section := sections[sectionName]
		repo := repositories[section.Name]
		if len(repo) == 0 {
			continue
		}

		// Skip if specific section is set and doesn't match
		if specificSection != "" && !strings.Contains(specificSection, section.Name) {
			continue
		}

		if err := writeSectionFile(&section, repo); err != nil {
			return err
		}
	}

	return nil
}

// writeSectionFile writes a single section's markdown file
func writeSectionFile(section *awesomego.Section, repo []awesomego.Repository) error {
	filename := "docs/" + convertToFilename(section.Name) + ".md"

	outputFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", filename, err)
	}
	defer outputFile.Close()

	// Write section header
	fmt.Printf("## %s\n\n%s\n\n", section.Name, section.Description)
	fmt.Fprintf(outputFile, "## %s\n\n%s\n\n", section.Name, section.Description)

	// Write rankings by Stars
	sort.Slice(repo, func(i, j int) bool {
		return repo[i].Stars > repo[j].Stars
	})
	printRepositories("\nRanked by Stars", repo)
	writeRepositoriesToFile("Ranked by Stars", repo, outputFile)

	// Write rankings by Forks
	sort.Slice(repo, func(i, j int) bool {
		return repo[i].Forks > repo[j].Forks
	})
	printRepositories("\nRanked by Forks", repo)
	writeRepositoriesToFile("Ranked by Forks", repo, outputFile)

	// Write rankings by Last Updated
	sort.Slice(repo, func(i, j int) bool {
		return repo[i].LastUpdated.After(repo[j].LastUpdated)
	})
	printRepositories("\nRanked by Last Updated", repo)
	writeRepositoriesToFile("Ranked by Last Updated", repo, outputFile)

	return nil
}

// convertToFilename converts section name to lowercase and replaces spaces with hyphens
func convertToFilename(name string) string {
	return strings.Replace(name, " ", "-", -1)
}

func main() {
	var (
		specificSection string
		limit           int
	)

	rootCmd := &cobra.Command{
		Use:   "awesome-go",
		Short: "A CLI tool to rank GitHub repositories from awesome-go's README.md",
		Run: func(cmd *cobra.Command, args []string) {
			config := Config{
				GitHubToken:     os.Getenv("GITHUB_TOKEN"),
				SpecificSection: specificSection,
				Limit:           limit,
				OutputDir:       ".",
			}

			if err := runRanking(config); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Results written to README.md")
		},
	}

	rootCmd.Flags().StringVarP(&specificSection, "section", "s", "", "A specific section to rank")
	rootCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit the number of results")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
