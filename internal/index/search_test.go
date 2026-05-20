package index

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestSearchMemoriesFindsRankedMatches(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	fixtures := []MemoryRecord{
		{
			ID:         "mem_1",
			ProjectID:  "project-1",
			Kind:       "decision",
			Title:      "Repository layer stays framework-agnostic",
			Content:    "The repository package must remain framework-agnostic for portability.",
			Tags:       []string{"architecture", "repository"},
			Confidence: 1,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "mem_2",
			ProjectID:  "project-1",
			Kind:       "constraint",
			Title:      "Auth middleware contract",
			Content:    "Auth middleware cannot depend on repository internals.",
			Tags:       []string{"auth"},
			Confidence: 0.8,
			CreatedAt:  now.Add(-24 * time.Hour),
			UpdatedAt:  now.Add(-24 * time.Hour),
		},
	}
	for _, record := range fixtures {
		if err := UpsertMemory(context.Background(), db, record); err != nil {
			t.Fatal(err)
		}
	}
	results, err := SearchMemories(context.Background(), db, SearchQuery{ProjectID: "project-1", Query: "repository framework", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results", len(results))
	}
	if results[0].ID != "mem_1" {
		t.Fatalf("got first result %q", results[0].ID)
	}
	if results[0].Snippet == "" || results[0].Score <= 0 {
		t.Fatalf("unexpected result: %#v", results[0])
	}
}

func TestSearchMemoriesFiltersKinds(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	for _, record := range []MemoryRecord{
		{ID: "mem_1", ProjectID: "project-1", Kind: "decision", Title: "Repository decision", Content: "repository framework", Confidence: 1, CreatedAt: now, UpdatedAt: now},
		{ID: "mem_2", ProjectID: "project-1", Kind: "constraint", Title: "Repository constraint", Content: "repository framework", Confidence: 1, CreatedAt: now, UpdatedAt: now},
	} {
		if err := UpsertMemory(context.Background(), db, record); err != nil {
			t.Fatal(err)
		}
	}
	results, err := SearchMemories(context.Background(), db, SearchQuery{ProjectID: "project-1", Query: "repository framework", Kinds: []string{"constraint"}, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Kind != "constraint" {
		t.Fatalf("unexpected results: %#v", results)
	}
}

func TestSearchMemoriesHybridKeepsMatchesSearchable(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	for _, record := range []MemoryRecord{
		{ID: "mem_1", ProjectID: "project-1", Kind: "decision", Title: "Repository architecture", Content: "repository framework", Confidence: 1, CreatedAt: now, UpdatedAt: now},
		{ID: "mem_2", ProjectID: "project-1", Kind: "decision", Title: "Repository testing", Content: "repository framework", Confidence: 1, CreatedAt: now, UpdatedAt: now},
	} {
		if err := UpsertMemory(context.Background(), db, record); err != nil {
			t.Fatal(err)
		}
	}
	results, err := SearchMemories(context.Background(), db, SearchQuery{ProjectID: "project-1", Query: "repository framework", Limit: 10, Hybrid: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || results[0].Score <= 0 {
		t.Fatalf("unexpected results: %#v", results)
	}
}

func TestSearchMemoriesConfidenceDecayKeepsOldDecisionBelowConstraint(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC()
	for _, record := range []MemoryRecord{
		{ID: "mem_old", ProjectID: "project-1", Kind: "decision", Title: "Repository framework", Content: "repository framework", Confidence: 1, CreatedAt: now.AddDate(-2, 0, 0), UpdatedAt: now.AddDate(-2, 0, 0)},
		{ID: "mem_constraint", ProjectID: "project-1", Kind: "constraint", Title: "Repository framework", Content: "repository framework", Confidence: 0.8, CreatedAt: now.AddDate(-2, 0, 0), UpdatedAt: now.AddDate(-2, 0, 0)},
	} {
		if err := UpsertMemory(context.Background(), db, record); err != nil {
			t.Fatal(err)
		}
	}
	results, err := SearchMemories(context.Background(), db, SearchQuery{ProjectID: "project-1", Query: "repository framework", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || results[0].ID != "mem_constraint" {
		t.Fatalf("unexpected order: %#v", results)
	}
}
