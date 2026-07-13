package awesomego

import (
	"testing"
	"time"
)

func TestEnrichRepositoryTrendsCalculatesExactDeltasAndGrowth(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)
	writeTrendSnapshot(t, dir, now.AddDate(0, 0, -30), map[string][]Repository{
		"Database": {{Name: "example/database", Stars: 50}},
	})
	writeTrendSnapshot(t, dir, now.AddDate(0, 0, -7), map[string][]Repository{
		"Database": {{Name: "example/database", Stars: 100}},
	})
	writeTrendSnapshot(t, dir, now.AddDate(0, 0, -1), map[string][]Repository{
		"Database": {{Name: "example/database", Stars: 140}},
	})

	repos := map[string][]Repository{
		"Database": {
			{Name: "example/database", Stars: 150},
			{Name: "example/new", Stars: 120},
		},
	}
	if err := EnrichRepositoryTrends(repos, dir, now); err != nil {
		t.Fatal(err)
	}

	existing := repos["Database"][0]
	if existing.StarsDelta7d == nil || *existing.StarsDelta7d != 50 {
		t.Fatalf("7d delta = %v, want 50", existing.StarsDelta7d)
	}
	if existing.StarsDelta30d == nil || *existing.StarsDelta30d != 100 {
		t.Fatalf("30d delta = %v, want 100", existing.StarsDelta30d)
	}
	if existing.StarsGrowth7d == nil || *existing.StarsGrowth7d != 0.5 {
		t.Fatalf("7d growth = %v, want 0.5", existing.StarsGrowth7d)
	}
	if existing.StarsGrowth30d == nil || *existing.StarsGrowth30d != 2.0 {
		t.Fatalf("30d growth = %v, want 2.0", existing.StarsGrowth30d)
	}
	if existing.IsNew {
		t.Fatal("existing repository marked as new")
	}
	if !repos["Database"][1].IsNew {
		t.Fatal("new repository was not marked as new")
	}
}

func TestEnrichRepositoryTrendsLeavesMissingPeriodsNull(t *testing.T) {
	repos := map[string][]Repository{
		"Database": {{Name: "example/database", Stars: 150}},
	}
	if err := EnrichRepositoryTrends(repos, t.TempDir(), time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	got := repos["Database"][0]
	if got.StarsDelta7d != nil || got.StarsDelta30d != nil || got.StarsGrowth7d != nil || got.StarsGrowth30d != nil {
		t.Fatalf("missing periods must remain nil: %#v", got)
	}
	if got.IsNew {
		t.Fatal("repository must not be marked new without comparison history")
	}
}

func writeTrendSnapshot(t *testing.T, dir string, date time.Time, repos map[string][]Repository) {
	t.Helper()
	if _, err := SaveSnapshot(repos, dir, date); err != nil {
		t.Fatal(err)
	}
}
