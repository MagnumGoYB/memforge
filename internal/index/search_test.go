package index

import (
	"context"
	"path/filepath"
	"testing"
	"time"
	"unicode/utf8"
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

func TestSearchMemoriesFallsBackToPartialMatchesForBroadQueries(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	for _, record := range []MemoryRecord{
		{
			ID:         "mem_display",
			ProjectID:  "project-1",
			Kind:       "agent-instruction",
			Title:      "Plugin display name uses interface metadata",
			Content:    "Use interface displayName MemForge for UI display casing.",
			Tags:       []string{"plugin", "display-name"},
			Confidence: 1,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "mem_release",
			ProjectID:  "project-1",
			Kind:       "agent-instruction",
			Title:      "Release acceptance checks plugin assets and curl latest",
			Content:    "Validate release assets and curl latest install.",
			Tags:       []string{"release", "curl"},
			Confidence: 1,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	} {
		if err := UpsertMemory(context.Background(), db, record); err != nil {
			t.Fatal(err)
		}
	}
	results, err := SearchMemories(context.Background(), db, SearchQuery{ProjectID: "project-1", Query: "MemForge plugin displayName bundled runtimes release acceptance curl latest shared MEMFORGE_HOME", Limit: 10, Hybrid: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) < 2 {
		t.Fatalf("broad query should return partial matches, got %#v", results)
	}
	seen := map[string]bool{}
	for _, result := range results {
		seen[result.ID] = true
	}
	if !seen["mem_display"] || !seen["mem_release"] {
		t.Fatalf("broad query missed expected partial matches: %#v", results)
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

func TestSearchMemoriesSnippetKeepsUTF8Valid(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	content := "前缀内容用于制造较长文本，让片段截取边界更容易落在多字节字符附近。搜索关键词后面继续追加中文内容，确保结尾也可能被截断。"
	if err := UpsertMemory(context.Background(), db, MemoryRecord{ID: "mem_utf8", ProjectID: "project-1", Kind: "decision", Title: "中文片段", Content: content, Tags: []string{"repository"}, Confidence: 1, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	results, err := SearchMemories(context.Background(), db, SearchQuery{ProjectID: "project-1", Query: "repository", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results", len(results))
	}
	if !utf8.ValidString(results[0].Snippet) {
		t.Fatalf("snippet is not valid UTF-8: %q", results[0].Snippet)
	}
}
