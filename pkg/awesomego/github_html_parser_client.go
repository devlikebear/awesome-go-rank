package awesomego

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/devlikebear/awesome-go-rank/pkg/stringutil"
	"golang.org/x/net/html"
)

// GithubHtmlParserClient is a struct that represents a Github client.
type GithubHtmlParserClient struct {
}

// NewGithubHtmlParserClient creates a new GithubHtmlParserClient instance.
func NewGithubHtmlParserClient() *GithubHtmlParserClient {
	return &GithubHtmlParserClient{}
}

// FetchReadmeMarkdown fetches the README.md file of a given repository.
func (ghpc *GithubHtmlParserClient) FetchReadmeMarkdown(ctx context.Context, owner, repo string) (string, error) {
	client := NewGithubClient("")
	return client.FetchReadmeMarkdown(ctx, owner, repo)
}

// FetchRepository fetches the repositories from parsing the HTML page.
func (ghpc *GithubHtmlParserClient) FetchRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	repoURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	doc, err := fetchHTML(repoURL)
	if err != nil {
		fmt.Println("Error fetching HTML:", err)
		return nil, err
	}

	stars, forks, lastUpdated, err := parseRepoInfo(doc)
	if err != nil {
		fmt.Println("Error parsing repo info:", err)
		return nil, err
	}

	return &Repository{
		Name:        fmt.Sprintf("%s/%s", owner, repo),
		Stars:       stars,
		Forks:       forks,
		LastUpdated: lastUpdated,
	}, nil
}

func fetchHTML(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := htmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func parseRepoInfo(doc *html.Node) (int, int, time.Time, error) {
	// Parse stars
	starsNode, err := htmlquery.Query(doc, "//span[@class='Counter js-social-count']")
	if err != nil {
		return 0, 0, time.Time{}, err
	}
	starsStr := strings.TrimSpace(htmlquery.InnerText(starsNode))
	stars, err := stringutil.ParseMetricNumber(starsStr)
	if err != nil {
		return 0, 0, time.Time{}, err
	}

	// Parse forks
	forksNode, err := htmlquery.Query(doc, "//span[@class='Counter']")
	if err != nil {
		return 0, 0, time.Time{}, err
	}
	forksStr := strings.TrimSpace(htmlquery.InnerText(forksNode))
	forks, err := stringutil.ParseMetricNumber(forksStr)
	if err != nil {
		return 0, 0, time.Time{}, err
	}

	// Parse last updated
	updatedNode, err := htmlquery.Query(doc, "//relative-time")
	if err != nil {
		return 0, 0, time.Time{}, err
	}
	updatedStr := strings.TrimSpace(htmlquery.SelectAttr(updatedNode, "datetime"))
	var updatedTime time.Time
	if updatedStr != "" {
		updatedTime, err = time.Parse(time.RFC3339, updatedStr)
		if err != nil {
			return 0, 0, time.Time{}, err
		}
	}

	return stars, forks, updatedTime, nil
}