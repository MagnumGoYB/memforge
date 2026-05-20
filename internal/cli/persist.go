package cli

import (
	"context"

	"github.com/MagnumGOYB/memforge/internal/index"
	"github.com/MagnumGOYB/memforge/internal/memory"
)

func persistMemory(ctx context.Context, paths resolvedPaths, record memory.Record) error {
	if _, err := memory.AppendMarkdown(paths.MemoriesDir, record); err != nil {
		return err
	}
	db, err := index.Open(paths.Index)
	if err != nil {
		return err
	}
	defer db.Close()
	return index.UpsertMemory(ctx, db, indexRecord(record))
}

func indexRecord(record memory.Record) index.MemoryRecord {
	return index.MemoryRecord{
		ID:         record.ID,
		ProjectID:  record.ProjectID,
		Kind:       string(record.Kind),
		Title:      record.Title,
		Content:    record.Content,
		Tags:       record.Tags,
		Source:     record.Source,
		Confidence: record.Confidence,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
	}
}
