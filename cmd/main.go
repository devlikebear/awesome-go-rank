package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
	Verbose         bool
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
	file io.Writer) error {
	if _, err := fmt.Fprintf(file, "### %s\n\n", title); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(file, "| Repository | Stars | Forks | Last Updated | Description | "); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(file, "|------------|-------|-------|--------------|-------------|"); err != nil {
		return err
	}
	for _, repo := range repositories {
		if _, err := fmt.Fprintf(file, "| [%s](%s) | %s | %s | %v | %s |\n", repo.Name,
			repo.URL,
			stringutil.FormatMetricNumber(repo.Stars),
			stringutil.FormatMetricNumber(repo.Forks),
			repo.LastUpdated.Format("2006-01-02T15:04:05Z"),
			strings.Trim(repo.Description, "-")); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(file)
	return err
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
		_, _ = fmt.Fprintf(&sb, "* [%s](docs/%s.md)<br/>%s\n",
			section.Name, convertToFilename(section.Name), section.Description)
	}

	return sb.String()
}

// writeReadmeHeader writes the header section of README.md
func writeReadmeHeader(w io.Writer) error {
	_, err := w.Write([]byte(`# Awesome Go Ranking

[![Website](https://img.shields.io/badge/Website-awesome--go--rank.vercel.app-blue?style=for-the-badge&logo=vercel)](https://awesome-go-rank.vercel.app/)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go)](go.mod)
[![Next.js](https://img.shields.io/badge/Next.js-14-black?style=for-the-badge&logo=next.js)](web/package.json)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

> **🌐 [Visit the Live Website](https://awesome-go-rank.vercel.app/)** - Explore Go repositories with real-time search, filtering, and sorting!

Discover and explore the best Go repositories from [awesome-go](https://github.com/avelino/awesome-go) ranked by Stars, Forks, and Last Updated. This project provides both a **web interface** and a **Go CLI tool** for browsing and analyzing Go libraries.

## ✨ Features

### 🖥️ Web Interface (https://awesome-go-rank.vercel.app/)
- **🔍 Real-time Search** - Instantly find repositories by name or description
- **📊 Smart Filtering** - Browse by category (45+ categories) and minimum stars (1K+, 5K+, 10K+)
- **⚡ Multi-criteria Sorting** - Sort by Stars, Forks, or Last Updated (ASC/DESC)
- **🎨 Dark/Light Mode** - Beautiful responsive design with theme support
- **📱 Mobile-Friendly** - Works perfectly on all devices

### 🛠️ CLI Tool

` + "```bash" + `
# Clone the repository
git clone https://github.com/devlikebear/awesome-go-rank.git
cd awesome-go-rank

# Set up GitHub token
export GITHUB_TOKEN=your_github_token_here

# Run the tool
go run cmd/main.go
` + "```" + `

## 🌟 Why Use This?

- **Discover Popular Go Libraries** - Quickly find the most starred and actively maintained Go projects
- **Stay Updated** - Daily automated updates ensure you see the latest trends
- **Filter by Category** - Easily find libraries for specific use cases

## 🔗 Links

- **Live Website**: https://awesome-go-rank.vercel.app/
- **Source Repository**: https://github.com/devlikebear/awesome-go-rank
- **awesome-go**: https://github.com/avelino/awesome-go

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

	if cmdConfig.SpecificSection == "" && cmdConfig.Limit == 0 {
		capturedAt := time.Now().UTC()
		snapshotDir := filepath.Join("data", "snapshots")
		if _, err := awesomego.SaveSnapshot(ag.Repositories(), snapshotDir, capturedAt); err != nil {
			return fmt.Errorf("failed to save repository snapshot: %w", err)
		}
		if err := awesomego.EnrichRepositoryTrends(ag.Repositories(), snapshotDir, capturedAt); err != nil {
			return fmt.Errorf("failed to calculate repository trends: %w", err)
		}
		if err := awesomego.ThinSnapshots(snapshotDir, capturedAt, 90); err != nil {
			return fmt.Errorf("failed to thin repository snapshots: %w", err)
		}
	}

	// Create output directory if it doesn't exist
	docsPath := cfg.GetDocsPath()
	if err := os.MkdirAll(docsPath, 0o750); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Write README.md
	if err := writeReadme(ag, cmdConfig.SpecificSection); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	// Write section files
	if err := writeSectionFiles(ag, cmdConfig.SpecificSection, cmdConfig.Verbose); err != nil {
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

	// web/public/data is the single canonical location consumed by the static site.
	outputPath := filepath.Join("web", "public", "data", "repos.json")
	if err := exporter.Export(outputPath, cfg.GitHub.Owner, cfg.GitHub.Repository); err != nil {
		return err
	}

	return nil
}

// writeReadme writes the main README.md file
func writeReadme(ag *awesomego.AwesomeGo, specificSection string) (err error) {
	outputFile, err := os.OpenFile("README.md", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := outputFile.Close(); err == nil {
			err = closeErr
		}
	}()

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
func writeSectionFiles(ag *awesomego.AwesomeGo, specificSection string, verbose bool) error {
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
		if !awesomego.MatchesSection(specificSection, section.Name) {
			continue
		}

		if err := writeSectionFile(&section, repo, verbose); err != nil {
			return err
		}
	}

	return nil
}

// writeSectionFile writes a single section's markdown file
func writeSectionFile(section *awesomego.Section, repo []awesomego.Repository, verbose bool) (err error) {
	filename := "docs/" + convertToFilename(section.Name) + ".md"

	outputFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- filename is derived from a sanitized section name.
	if err != nil {
		return fmt.Errorf("error creating %s: %w", filename, err)
	}
	defer func() {
		if closeErr := outputFile.Close(); err == nil {
			err = closeErr
		}
	}()

	// Write section header
	if verbose {
		fmt.Printf("## %s\n\n%s\n\n", section.Name, section.Description)
	}
	if _, err := fmt.Fprintf(outputFile, "## %s\n\n%s\n\n", section.Name, section.Description); err != nil {
		return err
	}

	// Write rankings by Stars
	sort.Slice(repo, func(i, j int) bool {
		return repo[i].Stars > repo[j].Stars
	})
	if verbose {
		printRepositories("\nRanked by Stars", repo)
	}
	if err := writeRepositoriesToFile("Ranked by Stars", repo, outputFile); err != nil {
		return err
	}

	// Write rankings by Forks
	sort.Slice(repo, func(i, j int) bool {
		return repo[i].Forks > repo[j].Forks
	})
	if verbose {
		printRepositories("\nRanked by Forks", repo)
	}
	if err := writeRepositoriesToFile("Ranked by Forks", repo, outputFile); err != nil {
		return err
	}

	// Write rankings by Last Updated
	sort.Slice(repo, func(i, j int) bool {
		return repo[i].LastUpdated.After(repo[j].LastUpdated)
	})
	if verbose {
		printRepositories("\nRanked by Last Updated", repo)
	}
	if err := writeRepositoriesToFile("Ranked by Last Updated", repo, outputFile); err != nil {
		return err
	}

	return nil
}

// convertToFilename converts section name to lowercase and replaces spaces with hyphens
func convertToFilename(name string) string {
	replacer := strings.NewReplacer(" ", "-", "/", "-", `\`, "-")
	return replacer.Replace(name)
}

func main() {
	var (
		specificSection string
		limit           int
		verbose         bool
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
				Verbose:         verbose,
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
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print repository details")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
