package awesomego

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func TestSaveSnapshotWritesVersionedGzip(t *testing.T) {
	capturedAt := time.Date(2026, 7, 12, 8, 30, 0, 0, time.UTC)
	repos := map[string][]Repository{
		"Database": {{
			Name:        "example/database",
			Stars:       120,
			Forks:       12,
			LastUpdated: time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC),
			Archived:    true,
		}},
	}

	path, err := SaveSnapshot(repos, t.TempDir(), capturedAt)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "2026-07-12.json.gz" {
		t.Fatalf("unexpected snapshot path: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.Close() }()

	var snapshot Snapshot
	if err := json.NewDecoder(reader).Decode(&snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Version != 1 || len(snapshot.Repositories) != 1 {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	got := snapshot.Repositories[0]
	if got.FullName != "example/database" || got.Section != "Database" || !got.Archived {
		t.Fatalf("unexpected repository snapshot: %#v", got)
	}
}

func TestThinSnapshotsKeepsDailyNinetyDaysAndWeeklyHistory(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	dates := []string{
		"2026-03-02", "2026-03-04", // same old ISO week: keep the latest
		"2026-03-09",               // another old week
		"2026-04-13", "2026-04-14", // within the 90-day daily window
	}
	for _, date := range dates {
		if err := os.WriteFile(filepath.Join(dir, date+".json.gz"), []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignore"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ThinSnapshots(dir, now, 90); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	want := []string{
		"2026-03-04.json.gz",
		"2026-03-09.json.gz",
		"2026-04-13.json.gz",
		"2026-04-14.json.gz",
		"README.md",
	}
	if len(names) != len(want) {
		t.Fatalf("snapshot files = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("snapshot files = %v, want %v", names, want)
		}
	}
}
