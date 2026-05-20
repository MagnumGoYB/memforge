package index

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

type MemoryRecord struct {
	ID         string
	ProjectID  string
	Kind       string
	Title      string
	Content    string
	Tags       []string
	Source     string
	Confidence float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func UpsertMemory(ctx context.Context, db *sql.DB, record MemoryRecord) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := upsertMemory(ctx, tx, record); err != nil {
		return err
	}
	return tx.Commit()
}

type memoryExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func upsertMemory(ctx context.Context, exec memoryExecutor, record MemoryRecord) error {
	tagsJSON, err := json.Marshal(record.Tags)
	if err != nil {
		return err
	}
	tagsFlat := strings.Join(record.Tags, " ")
	_, err = exec.ExecContext(ctx, `INSERT INTO memories (
			id, project_id, kind, title, content, tags_json, tags_flat, source, confidence, usage_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		project_id = excluded.project_id,
		kind = excluded.kind,
		title = excluded.title,
		content = excluded.content,
		tags_json = excluded.tags_json,
		tags_flat = excluded.tags_flat,
		source = excluded.source,
		confidence = excluded.confidence,
		updated_at = excluded.updated_at`,
		record.ID,
		record.ProjectID,
		record.Kind,
		record.Title,
		record.Content,
		string(tagsJSON),
		tagsFlat,
		record.Source,
		record.Confidence,
		record.CreatedAt.Format(time.RFC3339),
		record.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}
	return nil
}
