package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	// GitHub defaults
	assert.Equal(t, "", cfg.GitHub.Token)
	assert.Equal(t, "avelino", cfg.GitHub.Owner)
	assert.Equal(t, "awesome-go", cfg.GitHub.Repository)

	// Output defaults
	assert.Equal(t, ".", cfg.Output.OutputDir)
	assert.Equal(t, "docs", cfg.Output.DocsDir)
	assert.Equal(t, "README.md", cfg.Output.ReadmeFile)
	assert.True(t, cfg.Output.EnableConsole)

	// Rate limit defaults
	assert.Equal(t, 10, cfg.RateLimit.RequestsPerSecond)
	assert.Equal(t, 100*time.Millisecond, cfg.RateLimit.MinInterval)
	assert.Equal(t, 3, cfg.RateLimit.MaxRetries)
	assert.Equal(t, 2.0, cfg.RateLimit.BackoffMultiplier)
}

func TestFromEnv_Defaults(t *testing.T) {
	// Clear all relevant env vars
	clearEnvVars(t)

	cfg, err := FromEnv()
	require.NoError(t, err)

	// Should return defaults
	assert.Equal(t, "avelino", cfg.GitHub.Owner)
	assert.Equal(t, "awesome-go", cfg.GitHub.Repository)
}

func TestFromEnv_GitHubConfig(t *testing.T) {
	clearEnvVars(t)

	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("SOURCE_OWNER", "test-owner")
	os.Setenv("SOURCE_REPO", "test-repo")
	defer clearEnvVars(t)

	cfg, err := FromEnv()
	require.NoError(t, err)

	assert.Equal(t, "test-token", cfg.GitHub.Token)
	assert.Equal(t, "test-owner", cfg.GitHub.Owner)
	assert.Equal(t, "test-repo", cfg.GitHub.Repository)
}

func TestFromEnv_OutputConfig(t *testing.T) {
	clearEnvVars(t)

	os.Setenv("OUTPUT_DIR", "/tmp/output")
	os.Setenv("DOCS_DIR", "documentation")
	os.Setenv("README_FILE", "INDEX.md")
	os.Setenv("ENABLE_CONSOLE", "false")
	defer clearEnvVars(t)

	cfg, err := FromEnv()
	require.NoError(t, err)

	assert.Equal(t, "/tmp/output", cfg.Output.OutputDir)
	assert.Equal(t, "documentation", cfg.Output.DocsDir)
	assert.Equal(t, "INDEX.md", cfg.Output.ReadmeFile)
	assert.False(t, cfg.Output.EnableConsole)
}

func TestFromEnv_RateLimitConfig(t *testing.T) {
	clearEnvVars(t)

	os.Setenv("RATE_LIMIT_RPS", "5")
	os.Setenv("MAX_RETRIES", "5")
	defer clearEnvVars(t)

	cfg, err := FromEnv()
	require.NoError(t, err)

	assert.Equal(t, 5, cfg.RateLimit.RequestsPerSecond)
	assert.Equal(t, 200*time.Millisecond, cfg.RateLimit.MinInterval)
	assert.Equal(t, 5, cfg.RateLimit.MaxRetries)
}

func TestFromEnv_InvalidRateLimit(t *testing.T) {
	tests := []struct {
		name  string
		rps   string
		retry string
	}{
		{"invalid RPS", "abc", ""},
		{"zero RPS", "0", ""},
		{"negative RPS", "-1", ""},
		{"invalid retries", "", "abc"},
		{"negative retries", "", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)

			if tt.rps != "" {
				os.Setenv("RATE_LIMIT_RPS", tt.rps)
			}
			if tt.retry != "" {
				os.Setenv("MAX_RETRIES", tt.retry)
			}
			defer clearEnvVars(t)

			_, err := FromEnv()
			assert.Error(t, err)
		})
	}
}

func TestFromEnv_EnableConsole(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"1", "1", true},
		{"false", "false", false},
		{"0", "0", false},
		{"other", "anything", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			os.Setenv("ENABLE_CONSOLE", tt.value)
			defer clearEnvVars(t)

			cfg, err := FromEnv()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.Output.EnableConsole)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "valid config",
			config:    validConfig(),
			shouldErr: false,
		},
		{
			name: "missing token",
			config: func() *Config {
				c := validConfig()
				c.GitHub.Token = ""
				return c
			}(),
			shouldErr: true,
			errMsg:    "token is required",
		},
		{
			name: "missing owner",
			config: func() *Config {
				c := validConfig()
				c.GitHub.Owner = ""
				return c
			}(),
			shouldErr: true,
			errMsg:    "owner is required",
		},
		{
			name: "missing repository",
			config: func() *Config {
				c := validConfig()
				c.GitHub.Repository = ""
				return c
			}(),
			shouldErr: true,
			errMsg:    "repository is required",
		},
		{
			name: "invalid RPS",
			config: func() *Config {
				c := validConfig()
				c.RateLimit.RequestsPerSecond = 0
				return c
			}(),
			shouldErr: true,
			errMsg:    "must be positive",
		},
		{
			name: "negative retries",
			config: func() *Config {
				c := validConfig()
				c.RateLimit.MaxRetries = -1
				return c
			}(),
			shouldErr: true,
			errMsg:    "must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSourceURL(t *testing.T) {
	cfg := &Config{
		GitHub: GitHubConfig{
			Owner:      "testowner",
			Repository: "testrepo",
		},
	}

	assert.Equal(t, "https://github.com/testowner/testrepo", cfg.GetSourceURL())
}

func TestGetReadmePath(t *testing.T) {
	tests := []struct {
		name       string
		outputDir  string
		readmeFile string
		expected   string
	}{
		{"current dir", ".", "README.md", "README.md"},
		{"custom dir", "/tmp", "README.md", "/tmp/README.md"},
		{"custom file", ".", "INDEX.md", "INDEX.md"},
		{"both custom", "output", "INDEX.md", "output/INDEX.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Output: OutputConfig{
					OutputDir:  tt.outputDir,
					ReadmeFile: tt.readmeFile,
				},
			}
			assert.Equal(t, tt.expected, cfg.GetReadmePath())
		})
	}
}

func TestGetDocsPath(t *testing.T) {
	tests := []struct {
		name      string
		outputDir string
		docsDir   string
		expected  string
	}{
		{"current dir", ".", "docs", "docs"},
		{"custom output", "/tmp", "docs", "/tmp/docs"},
		{"custom docs", ".", "documentation", "documentation"},
		{"both custom", "output", "documentation", "output/documentation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Output: OutputConfig{
					OutputDir: tt.outputDir,
					DocsDir:   tt.docsDir,
				},
			}
			assert.Equal(t, tt.expected, cfg.GetDocsPath())
		})
	}
}

// Helper functions

func validConfig() *Config {
	cfg := Default()
	cfg.GitHub.Token = "test-token"
	return cfg
}

func clearEnvVars(t *testing.T) {
	vars := []string{
		"GITHUB_TOKEN",
		"SOURCE_OWNER",
		"SOURCE_REPO",
		"OUTPUT_DIR",
		"DOCS_DIR",
		"README_FILE",
		"ENABLE_CONSOLE",
		"RATE_LIMIT_RPS",
		"MAX_RETRIES",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}
