package awesomego

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const snapshotVersion = 1

// Snapshot is a point-in-time copy of repository ranking inputs.
type Snapshot struct {
	Version      int                  `json:"version"`
	CapturedAt   time.Time            `json:"captured_at"`
	Repositories []SnapshotRepository `json:"repositories"`
}

// SnapshotRepository is the stable, versioned repository snapshot schema.
type SnapshotRepository struct {
	FullName    string    `json:"full_name"`
	Section     string    `json:"section"`
	Stars       int       `json:"stars"`
	Forks       int       `json:"forks"`
	LastUpdated time.Time `json:"last_updated"`
	Archived    bool      `json:"archived"`
}

// SaveSnapshot writes a deterministic gzip-compressed snapshot for capturedAt.
func SaveSnapshot(repos map[string][]Repository, dir string, capturedAt time.Time) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create snapshot directory: %w", err)
	}

	snapshot := Snapshot{
		Version:      snapshotVersion,
		CapturedAt:   capturedAt.UTC(),
		Repositories: flattenSnapshotRepositories(repos),
	}
	outputPath := filepath.Join(dir, capturedAt.UTC().Format("2006-01-02")+".json.gz")
	temp, err := os.CreateTemp(dir, ".snapshot-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temporary snapshot: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	gzipWriter := gzip.NewWriter(temp)
	encoder := json.NewEncoder(gzipWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		gzipWriter.Close()
		temp.Close()
		return "", fmt.Errorf("encode snapshot: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		temp.Close()
		return "", fmt.Errorf("close snapshot compressor: %w", err)
	}
	if err := temp.Close(); err != nil {
		return "", fmt.Errorf("close snapshot file: %w", err)
	}
	if err := os.Chmod(tempPath, 0o644); err != nil {
		return "", fmt.Errorf("set snapshot permissions: %w", err)
	}
	if err := os.Rename(tempPath, outputPath); err != nil {
		return "", fmt.Errorf("publish snapshot: %w", err)
	}
	return outputPath, nil
}

func flattenSnapshotRepositories(repos map[string][]Repository) []SnapshotRepository {
	sections := make([]string, 0, len(repos))
	for section := range repos {
		sections = append(sections, section)
	}
	sort.Strings(sections)

	result := make([]SnapshotRepository, 0)
	for _, section := range sections {
		sectionRepos := append([]Repository(nil), repos[section]...)
		sort.Slice(sectionRepos, func(i, j int) bool { return sectionRepos[i].Name < sectionRepos[j].Name })
		for _, repo := range sectionRepos {
			result = append(result, SnapshotRepository{
				FullName:    repo.Name,
				Section:     section,
				Stars:       repo.Stars,
				Forks:       repo.Forks,
				LastUpdated: repo.LastUpdated,
				Archived:    repo.Archived,
			})
		}
	}
	return result
}

// ThinSnapshots keeps daily snapshots for retentionDays and one latest snapshot
// per ISO week for older history.
func ThinSnapshots(dir string, now time.Time, retentionDays int) error {
	if retentionDays < 0 {
		return fmt.Errorf("snapshot retention days must be non-negative")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read snapshot directory: %w", err)
	}

	today := now.UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	cutoff := today.AddDate(0, 0, -retentionDays)
	type datedFile struct {
		name string
		date time.Time
	}
	var oldSnapshots []datedFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json.gz") {
			continue
		}
		date, err := time.Parse("2006-01-02", strings.TrimSuffix(entry.Name(), ".json.gz"))
		if err != nil || !date.Before(cutoff) {
			continue
		}
		oldSnapshots = append(oldSnapshots, datedFile{name: entry.Name(), date: date})
	}

	sort.Slice(oldSnapshots, func(i, j int) bool { return oldSnapshots[i].date.After(oldSnapshots[j].date) })
	keptWeeks := make(map[string]struct{})
	for _, snapshot := range oldSnapshots {
		year, week := snapshot.date.ISOWeek()
		weekKey := fmt.Sprintf("%04d-%02d", year, week)
		if _, keep := keptWeeks[weekKey]; !keep {
			keptWeeks[weekKey] = struct{}{}
			continue
		}
		if err := os.Remove(filepath.Join(dir, snapshot.name)); err != nil {
			return fmt.Errorf("remove thinned snapshot %s: %w", snapshot.name, err)
		}
	}
	return nil
}
