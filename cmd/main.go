package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/devlikebear/awesome-go-rank/pkg/awesomego"
	"github.com/spf13/cobra"
)

func formatNumber(number int) string {
	numberStr := strconv.Itoa(number)
	result := ""

	for i, j := len(numberStr)-1, 0; i >= 0; i, j = i-1, j+1 {
		if j%3 == 0 && j != 0 {
			result = "," + result
		}
		result = string(numberStr[i]) + result
	}

	return result
}

func formatMetricNumber(number int) string {
	if number < 1000 {
		return strconv.Itoa(number)
	}
	suffixes := []string{"k", "m", "b", "t"}
	num := float64(number)
	var suffix string
	for i := len(suffixes) - 1; i >= 0; i-- {
		divisor := math.Pow(10, float64((i+1)*3))
		if num >= divisor {
			suffix = suffixes[i]
			num /= divisor
			break
		}
	}
	return fmt.Sprintf("%.1f%s", num, suffix)
}

func printRepositories(title string, repositories []awesomego.Repository) {
	fmt.Println(title)
	for _, repo := range repositories {
		fmt.Printf("%s (Stars: %s, Forks: %s, Last updated: %v)\n", repo.Name, formatMetricNumber(repo.Stars), formatMetricNumber(repo.Forks), repo.LastUpdated.Format("2006-01-02T15:04:05Z"))
	}
}

func writeRepositoriesToFile(title string, repositories []awesomego.Repository, file *os.File) {
	file.WriteString("### " + title + "\n\n")
	file.WriteString("| Repository | Stars | Forks | Last Updated |\n")
	file.WriteString("|------------|-------|-------|--------------|\n")
	for _, repo := range repositories {
		file.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s | %v |\n", repo.Name, repo.URL, formatMetricNumber(repo.Stars), formatMetricNumber(repo.Forks), repo.LastUpdated.Format("2006-01-02T15:04:05Z")))
	}
	file.WriteString("\n")
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
			githubToken := os.Getenv("GITHUB_TOKEN")
			if githubToken == "" {
				fmt.Println("Error: GITHUB_TOKEN environment variable is not set.")
				os.Exit(1)
			}

			ag := awesomego.NewAwesomeGo(githubToken)
			err := ag.FetchAndRankRepositories(specificSection, limit)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			outputFile, err := os.Create("README.md")
			if err != nil {
				fmt.Printf("Error creating README.md: %v\n", err)
				os.Exit(1)
			}
			defer outputFile.Close()

			repositories := ag.Repositories()

			// Writing the title and description to the README.md file
			outputFile.WriteString("# Awesome Go Ranking\n\n")
			outputFile.WriteString(`This is a ranking of GitHub repositories from
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
 `)

			outputFile.WriteString("\n\n")
			outputFile.WriteString("## Table of Contents\n\n")

			// Extract keys into a slice
			sections := make([]string, 0, len(repositories))
			for section := range repositories {
				sections = append(sections, section)
			}

			// Sort sections in ascending order
			sort.Strings(sections)

			// Iterate over sorted sections
			for _, section := range sections {
				repo := repositories[section]
				if len(repo) == 0 {
					continue
				}

				// Table of Contents
				outputFile.WriteString(fmt.Sprintf("* [%s](docs/%s.md)\n", section, convertToFilename(section)))

				// Skip section if specificSection is set and the section is not contain the specificSection
				if specificSection != "" && !strings.Contains(specificSection, section) {
					continue
				}

				// Printing and writing results to file by Stars in each section
				printSectionRank(section, repo)

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

// Converting section name to lowercase and replacing spaces with hyphens
func convertToFilename(name string) string {
	return strings.Replace(name, " ", "-", -1)
}

func printSectionRank(section string, repo []awesomego.Repository) {
	filename := "docs/" + convertToFilename(section) + ".md"

	// If the directory does not exist, create it
	if _, err := os.Stat("docs"); os.IsNotExist(err) {
		os.Mkdir("docs", 0755)
	}

	outputFile, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer outputFile.Close()

	// Printing and writing a section name
	fmt.Printf("## %s", section)
	outputFile.WriteString("## " + section + "\n\n")

	// Printing and writing results to file by Star
	sort.Slice(repo, func(i, j int) bool {
		return repo[i].Stars > repo[j].Stars
	})
	// Printing and writing results to file by Stars
	printRepositories("\nRanked by Stars", repo)
	writeRepositoriesToFile("Ranked by Stars", repo, outputFile)

	// Printing and writing results to file by Forks
	sort.Slice(repo, func(i, j int) bool {
		return repo[i].Forks > repo[j].Forks
	})
	printRepositories("\nRanked by Forks", repo)
	writeRepositoriesToFile("Ranked by Forks", repo, outputFile)

	// Printing and writing results to file by Last Updated
	sort.Slice(repo, func(i, j int) bool {
		return repo[i].LastUpdated.After(repo[j].LastUpdated)
	})
	printRepositories("\nRanked by Last Updated", repo)
	writeRepositoriesToFile("Ranked by Last Updated", repo, outputFile)
}
