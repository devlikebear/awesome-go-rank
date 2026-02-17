package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// GitHubConfig holds GitHub-related configuration
type GitHubConfig struct {
	Token      string
	Owner      string
	Repository string
}

// OutputConfig holds output-related configuration
type OutputConfig struct {
	OutputDir     string
	DocsDir       string
	ReadmeFile    string
	EnableConsole bool
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int
	MinInterval       time.Duration
	MaxRetries        int
	BackoffMultiplier float64
}

// Config holds all application configuration
type Config struct {
	GitHub    GitHubConfig
	Output    OutputConfig
	RateLimit RateLimitConfig
}

// Default returns a Config with default values
func Default() *Config {
	return &Config{
		GitHub: GitHubConfig{
			Token:      "",
			Owner:      "avelino",
			Repository: "awesome-go",
		},
		Output: OutputConfig{
			OutputDir:     ".",
			DocsDir:       "docs",
			ReadmeFile:    "README.md",
			EnableConsole: true,
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: 10,
			MinInterval:       100 * time.Millisecond,
			MaxRetries:        3,
			BackoffMultiplier: 2.0,
		},
	}
}

// FromEnv creates a Config from environment variables, falling back to defaults
func FromEnv() (*Config, error) {
	cfg := Default()

	// GitHub configuration
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		cfg.GitHub.Token = token
	}
	if owner := os.Getenv("SOURCE_OWNER"); owner != "" {
		cfg.GitHub.Owner = owner
	}
	if repo := os.Getenv("SOURCE_REPO"); repo != "" {
		cfg.GitHub.Repository = repo
	}

	// Output configuration
	if outputDir := os.Getenv("OUTPUT_DIR"); outputDir != "" {
		cfg.Output.OutputDir = outputDir
	}
	if docsDir := os.Getenv("DOCS_DIR"); docsDir != "" {
		cfg.Output.DocsDir = docsDir
	}
	if readmeFile := os.Getenv("README_FILE"); readmeFile != "" {
		cfg.Output.ReadmeFile = readmeFile
	}
	if enableConsole := os.Getenv("ENABLE_CONSOLE"); enableConsole != "" {
		cfg.Output.EnableConsole = enableConsole == "true" || enableConsole == "1"
	}

	// Rate limit configuration
	if rps := os.Getenv("RATE_LIMIT_RPS"); rps != "" {
		val, err := strconv.Atoi(rps)
		if err != nil {
			return nil, fmt.Errorf("invalid RATE_LIMIT_RPS: %w", err)
		}
		if val <= 0 {
			return nil, fmt.Errorf("RATE_LIMIT_RPS must be positive, got %d", val)
		}
		cfg.RateLimit.RequestsPerSecond = val
		cfg.RateLimit.MinInterval = time.Second / time.Duration(val)
	}
	if maxRetries := os.Getenv("MAX_RETRIES"); maxRetries != "" {
		val, err := strconv.Atoi(maxRetries)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_RETRIES: %w", err)
		}
		if val < 0 {
			return nil, fmt.Errorf("MAX_RETRIES must be non-negative, got %d", val)
		}
		cfg.RateLimit.MaxRetries = val
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.GitHub.Token == "" {
		return fmt.Errorf("GitHub token is required (set GITHUB_TOKEN environment variable)")
	}
	if c.GitHub.Owner == "" {
		return fmt.Errorf("GitHub owner is required")
	}
	if c.GitHub.Repository == "" {
		return fmt.Errorf("GitHub repository is required")
	}
	if c.RateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("requests per second must be positive")
	}
	if c.RateLimit.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}
	return nil
}

// GetSourceURL returns the full GitHub repository URL
func (c *Config) GetSourceURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", c.GitHub.Owner, c.GitHub.Repository)
}

// GetReadmePath returns the full path to the README file
func (c *Config) GetReadmePath() string {
	if c.Output.OutputDir == "." {
		return c.Output.ReadmeFile
	}
	return fmt.Sprintf("%s/%s", c.Output.OutputDir, c.Output.ReadmeFile)
}

// GetDocsPath returns the full path to the docs directory
func (c *Config) GetDocsPath() string {
	if c.Output.OutputDir == "." {
		return c.Output.DocsDir
	}
	return fmt.Sprintf("%s/%s", c.Output.OutputDir, c.Output.DocsDir)
}
