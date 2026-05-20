package index

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestRebuildMemoriesReplacesRows(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	for _, record := range []MemoryRecord{
		{ID: "mem_old", ProjectID: "project-1", Kind: "manual", Title: "old", Content: "old", Confidence: 1, CreatedAt: now, UpdatedAt: now},
	} {
		if err := UpsertMemory(context.Background(), db, record); err != nil {
			t.Fatal(err)
		}
	}
	stats, err := RebuildMemories(context.Background(), db, []MemoryRecord{{ID: "mem_new", ProjectID: "project-1", Kind: "decision", Title: "new", Content: "body", Confidence: 1, CreatedAt: now, UpdatedAt: now}})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Indexed != 1 || stats.Deleted != 1 || stats.Orphans != 1 || stats.Ghosts != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM memories`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got count %d", count)
	}
}
