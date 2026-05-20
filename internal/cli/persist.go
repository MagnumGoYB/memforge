package cli

import (
	"context"
	"fmt"

	"github.com/MagnumGOYB/memforge/internal/index"
	"github.com/MagnumGOYB/memforge/internal/memory"
)

func persistMemory(ctx context.Context, paths resolvedPaths, record memory.Record) (memory.AppendResult, string, error) {
	result, err := memory.AppendMarkdown(paths.MemoriesDir, record)
	if err != nil {
		return memory.AppendResult{}, "", err
	}
	db, err := index.Open(paths.Index)
	if err != nil {
		return result, fmt.Sprintf("memory saved to markdown but index update failed; run memforge reindex: %v", err), nil
	}
	defer db.Close()
	if err := index.UpsertMemory(ctx, db, indexRecord(record)); err != nil {
		return result, fmt.Sprintf("memory saved to markdown but index update failed; run memforge reindex: %v", err), nil
	}
	return result, "", nil
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
