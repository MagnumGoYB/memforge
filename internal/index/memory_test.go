package index

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestUpsertMemoryInsertsRow(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	record := MemoryRecord{
		ID:         "mem_1",
		ProjectID:  "project-1",
		Kind:       "decision",
		Title:      "CLI framework",
		Content:    "Use Cobra",
		Tags:       []string{"architecture", "cli"},
		Source:     "planning",
		Confidence: 0.9,
		CreatedAt:  time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
	}
	if err := UpsertMemory(context.Background(), db, record); err != nil {
		t.Fatal(err)
	}
	var title, tagsFlat string
	if err := db.QueryRow(`SELECT title, tags_flat FROM memories WHERE id = ?`, record.ID).Scan(&title, &tagsFlat); err != nil {
		t.Fatal(err)
	}
	if title != record.Title {
		t.Fatalf("got %q want %q", title, record.Title)
	}
	if tagsFlat != "architecture cli" {
		t.Fatalf("got %q", tagsFlat)
	}
}
