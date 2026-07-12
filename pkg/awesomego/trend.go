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

// EnrichRepositoryTrends adds exact 7-day and 30-day trend values from snapshots.
func EnrichRepositoryTrends(repos map[string][]Repository, snapshotDir string, capturedAt time.Time) error {
	sevenDay, err := loadSnapshotDate(snapshotDir, capturedAt.AddDate(0, 0, -7))
	if err != nil {
		return err
	}
	thirtyDay, err := loadSnapshotDate(snapshotDir, capturedAt.AddDate(0, 0, -30))
	if err != nil {
		return err
	}
	previous, err := loadLatestSnapshotBefore(snapshotDir, capturedAt)
	if err != nil {
		return err
	}

	sevenDayStars := snapshotStars(sevenDay)
	thirtyDayStars := snapshotStars(thirtyDay)
	previousRepos := snapshotRepositories(previous)
	for section, sectionRepos := range repos {
		for i := range sectionRepos {
			repo := &sectionRepos[i]
			repo.StarsDelta7d, repo.StarsGrowth7d = calculateStarTrend(repo.Stars, sevenDayStars, repo.Name)
			repo.StarsDelta30d, repo.StarsGrowth30d = calculateStarTrend(repo.Stars, thirtyDayStars, repo.Name)
			if previous != nil {
				_, existed := previousRepos[repo.Name]
				repo.IsNew = !existed
			}
		}
		repos[section] = sectionRepos
	}
	return nil
}

func calculateStarTrend(current int, previous map[string]int, fullName string) (*int, *float64) {
	previousStars, ok := previous[fullName]
	if !ok {
		return nil, nil
	}
	delta := current - previousStars
	if previousStars == 0 {
		return &delta, nil
	}
	growth := float64(delta) / float64(previousStars)
	return &delta, &growth
}

func loadSnapshotDate(dir string, date time.Time) (*Snapshot, error) {
	path := filepath.Join(dir, date.UTC().Format("2006-01-02")+".json.gz")
	return loadSnapshot(path)
}

func loadLatestSnapshotBefore(dir string, before time.Time) (*Snapshot, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read snapshot directory: %w", err)
	}
	type candidate struct {
		date time.Time
		path string
	}
	var candidates []candidate
	beforeDate := time.Date(before.UTC().Year(), before.UTC().Month(), before.UTC().Day(), 0, 0, 0, 0, time.UTC)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json.gz") {
			continue
		}
		date, err := time.Parse("2006-01-02", strings.TrimSuffix(entry.Name(), ".json.gz"))
		if err == nil && date.Before(beforeDate) {
			candidates = append(candidates, candidate{date: date, path: filepath.Join(dir, entry.Name())})
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].date.After(candidates[j].date) })
	return loadSnapshot(candidates[0].path)
}

func loadSnapshot(path string) (*Snapshot, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open snapshot %s: %w", path, err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("open snapshot gzip %s: %w", path, err)
	}
	defer reader.Close()
	var snapshot Snapshot
	if err := json.NewDecoder(reader).Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("decode snapshot %s: %w", path, err)
	}
	if snapshot.Version != snapshotVersion {
		return nil, fmt.Errorf("unsupported snapshot version %d in %s", snapshot.Version, path)
	}
	return &snapshot, nil
}

func snapshotStars(snapshot *Snapshot) map[string]int {
	stars := make(map[string]int)
	if snapshot == nil {
		return stars
	}
	for _, repo := range snapshot.Repositories {
		stars[repo.FullName] = repo.Stars
	}
	return stars
}

func snapshotRepositories(snapshot *Snapshot) map[string]struct{} {
	repos := make(map[string]struct{})
	if snapshot == nil {
		return repos
	}
	for _, repo := range snapshot.Repositories {
		repos[repo.FullName] = struct{}{}
	}
	return repos
}
