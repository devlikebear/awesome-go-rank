package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cache := New()
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.Data)
	assert.Equal(t, 0, len(cache.Data))
}

func TestCache_SetGet(t *testing.T) {
	cache := New()
	now := time.Now()

	// Set a repository
	cache.Set("owner", "repo", 1000, 100, now)

	// Get the repository
	repo, exists := cache.Get("owner", "repo")
	assert.True(t, exists)
	assert.Equal(t, "owner", repo.Owner)
	assert.Equal(t, "repo", repo.Name)
	assert.Equal(t, 1000, repo.Stars)
	assert.Equal(t, 100, repo.Forks)
	assert.Equal(t, now.Unix(), repo.LastUpdated.Unix())

	// Get non-existent repository
	_, exists = cache.Get("owner", "nonexistent")
	assert.False(t, exists)
}

func TestCache_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "test-cache.json")

	// Create and populate cache
	cache := New()
	now := time.Now()
	cache.Set("owner1", "repo1", 1000, 100, now)
	cache.Set("owner2", "repo2", 2000, 200, now)

	// Save cache
	err := cache.Save(cacheFile)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(cacheFile)
	require.NoError(t, err)

	// Load cache
	loaded, err := Load(cacheFile)
	require.NoError(t, err)
	assert.Equal(t, 2, len(loaded.Data))

	// Verify data
	repo1, exists := loaded.Get("owner1", "repo1")
	assert.True(t, exists)
	assert.Equal(t, 1000, repo1.Stars)

	repo2, exists := loaded.Get("owner2", "repo2")
	assert.True(t, exists)
	assert.Equal(t, 2000, repo2.Stars)
}

func TestLoad_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "nonexistent.json")

	// Load should return empty cache without error
	cache, err := Load(cacheFile)
	require.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Equal(t, 0, len(cache.Data))
}

func TestCache_IsExpired(t *testing.T) {
	cache := New()
	now := time.Now()

	// Add fresh entry
	cache.Set("owner", "fresh", 1000, 100, now)

	// Add old entry by manipulating CachedAt
	cache.Set("owner", "stale", 1000, 100, now)
	key := "owner/stale"
	repo := cache.Data[key]
	repo.CachedAt = now.Add(-25 * time.Hour)
	cache.Data[key] = repo

	// Test fresh entry
	assert.False(t, cache.IsExpired("owner", "fresh", 24*time.Hour))

	// Test stale entry
	assert.True(t, cache.IsExpired("owner", "stale", 24*time.Hour))

	// Test non-existent entry
	assert.True(t, cache.IsExpired("owner", "nonexistent", 24*time.Hour))
}

func TestCache_NeedsUpdate(t *testing.T) {
	cache := New()
	now := time.Now()

	// Add fresh entry
	cache.Set("owner", "fresh", 1000, 100, now)

	// Add old entry
	cache.Set("owner", "stale", 1000, 100, now)
	key := "owner/stale"
	repo := cache.Data[key]
	repo.CachedAt = now.Add(-25 * time.Hour)
	cache.Data[key] = repo

	// Fresh entry doesn't need update
	assert.False(t, cache.NeedsUpdate("owner", "fresh", 24*time.Hour))

	// Stale entry needs update
	assert.True(t, cache.NeedsUpdate("owner", "stale", 24*time.Hour))

	// Non-existent entry needs update
	assert.True(t, cache.NeedsUpdate("owner", "nonexistent", 24*time.Hour))
}

func TestCache_Stats(t *testing.T) {
	cache := New()
	now := time.Now()

	// Add fresh entries
	cache.Set("owner", "fresh1", 1000, 100, now)
	cache.Set("owner", "fresh2", 2000, 200, now)

	// Add stale entries
	cache.Set("owner", "stale1", 3000, 300, now)
	cache.Set("owner", "stale2", 4000, 400, now)

	// Manually set stale entries
	for _, name := range []string{"stale1", "stale2"} {
		key := "owner/" + name
		repo := cache.Data[key]
		repo.CachedAt = now.Add(-25 * time.Hour)
		cache.Data[key] = repo
	}

	stats := cache.Stats()
	assert.Equal(t, 4, stats.TotalEntries)
	assert.Equal(t, 2, stats.FreshEntries)
	assert.Equal(t, 2, stats.StaleEntries)
}

func TestCache_Prune(t *testing.T) {
	cache := New()
	now := time.Now()

	// Add fresh entries
	cache.Set("owner", "fresh", 1000, 100, now)

	// Add stale entries
	cache.Set("owner", "stale1", 2000, 200, now)
	cache.Set("owner", "stale2", 3000, 300, now)

	// Manually set stale entries
	for _, name := range []string{"stale1", "stale2"} {
		key := "owner/" + name
		repo := cache.Data[key]
		repo.CachedAt = now.Add(-25 * time.Hour)
		cache.Data[key] = repo
	}

	// Prune stale entries
	removed := cache.Prune(24 * time.Hour)
	assert.Equal(t, 2, removed)
	assert.Equal(t, 1, len(cache.Data))

	// Only fresh entry should remain
	_, exists := cache.Get("owner", "fresh")
	assert.True(t, exists)

	_, exists = cache.Get("owner", "stale1")
	assert.False(t, exists)
}

func TestCache_Clear(t *testing.T) {
	cache := New()
	now := time.Now()

	// Add entries
	cache.Set("owner1", "repo1", 1000, 100, now)
	cache.Set("owner2", "repo2", 2000, 200, now)
	assert.Equal(t, 2, len(cache.Data))

	// Clear cache
	cache.Clear()
	assert.Equal(t, 0, len(cache.Data))
}

func TestCache_SaveLoad_WithNestedDir(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "nested", "dir", "cache.json")

	cache := New()
	cache.Set("owner", "repo", 1000, 100, time.Now())

	// Save should create nested directories
	err := cache.Save(cacheFile)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(cacheFile)
	require.NoError(t, err)

	// Load and verify
	loaded, err := Load(cacheFile)
	require.NoError(t, err)
	assert.Equal(t, 1, len(loaded.Data))
}

func TestCache_UpdateExistingEntry(t *testing.T) {
	cache := New()
	now := time.Now()

	// Set initial entry
	cache.Set("owner", "repo", 1000, 100, now)
	repo, _ := cache.Get("owner", "repo")
	assert.Equal(t, 1000, repo.Stars)

	// Update entry
	time.Sleep(10 * time.Millisecond)
	cache.Set("owner", "repo", 2000, 200, now)
	repo, _ = cache.Get("owner", "repo")
	assert.Equal(t, 2000, repo.Stars)
	assert.Equal(t, 200, repo.Forks)
}

func TestCache_MultipleOwners(t *testing.T) {
	cache := New()
	now := time.Now()

	// Add repos from different owners
	cache.Set("owner1", "repo", 1000, 100, now)
	cache.Set("owner2", "repo", 2000, 200, now)

	// Both should exist and be different
	repo1, exists1 := cache.Get("owner1", "repo")
	repo2, exists2 := cache.Get("owner2", "repo")

	assert.True(t, exists1)
	assert.True(t, exists2)
	assert.Equal(t, 1000, repo1.Stars)
	assert.Equal(t, 2000, repo2.Stars)
}
