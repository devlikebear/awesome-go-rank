package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CachedRepo represents a cached repository with metadata
type CachedRepo struct {
	Owner       string    `json:"owner"`
	Name        string    `json:"name"`
	Stars       int       `json:"stars"`
	Forks       int       `json:"forks"`
	LastUpdated time.Time `json:"last_updated"`
	CachedAt    time.Time `json:"cached_at"`
}

// Cache represents the cache structure
type Cache struct {
	Data      map[string]CachedRepo `json:"data"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// New creates a new empty cache
func New() *Cache {
	return &Cache{
		Data:      make(map[string]CachedRepo),
		UpdatedAt: time.Now(),
	}
}

// Load loads the cache from a JSON file
func Load(filePath string) (*Cache, error) {
	// If file doesn't exist, return empty cache
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return New(), nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	var cache Cache
	if err := json.NewDecoder(file).Decode(&cache); err != nil {
		return nil, fmt.Errorf("failed to decode cache: %w", err)
	}

	// Initialize map if nil (for backwards compatibility)
	if cache.Data == nil {
		cache.Data = make(map[string]CachedRepo)
	}

	return &cache, nil
}

// Save saves the cache to a JSON file
func (c *Cache) Save(filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Update timestamp
	c.UpdatedAt = time.Now()

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode cache: %w", err)
	}

	return nil
}

// Get retrieves a repository from the cache
func (c *Cache) Get(owner, name string) (CachedRepo, bool) {
	key := fmt.Sprintf("%s/%s", owner, name)
	repo, exists := c.Data[key]
	return repo, exists
}

// Set adds or updates a repository in the cache
func (c *Cache) Set(owner, name string, stars, forks int, lastUpdated time.Time) {
	key := fmt.Sprintf("%s/%s", owner, name)
	c.Data[key] = CachedRepo{
		Owner:       owner,
		Name:        name,
		Stars:       stars,
		Forks:       forks,
		LastUpdated: lastUpdated,
		CachedAt:    time.Now(),
	}
}

// IsExpired checks if a cached repository is expired (older than maxAge)
func (c *Cache) IsExpired(owner, name string, maxAge time.Duration) bool {
	repo, exists := c.Get(owner, name)
	if !exists {
		return true
	}
	return time.Since(repo.CachedAt) > maxAge
}

// NeedsUpdate checks if a repository needs to be updated
// Returns true if:
// - Repository is not in cache
// - Cache entry is older than maxAge
func (c *Cache) NeedsUpdate(owner, name string, maxAge time.Duration) bool {
	return c.IsExpired(owner, name, maxAge)
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	now := time.Now()
	var fresh, stale int

	for _, repo := range c.Data {
		age := now.Sub(repo.CachedAt)
		if age <= 24*time.Hour {
			fresh++
		} else {
			stale++
		}
	}

	return CacheStats{
		TotalEntries: len(c.Data),
		FreshEntries: fresh,
		StaleEntries: stale,
		LastUpdated:  c.UpdatedAt,
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalEntries int
	FreshEntries int
	StaleEntries int
	LastUpdated  time.Time
}

// Prune removes expired entries from the cache
func (c *Cache) Prune(maxAge time.Duration) int {
	removed := 0
	now := time.Now()

	for key, repo := range c.Data {
		if now.Sub(repo.CachedAt) > maxAge {
			delete(c.Data, key)
			removed++
		}
	}

	return removed
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.Data = make(map[string]CachedRepo)
	c.UpdatedAt = time.Now()
}
