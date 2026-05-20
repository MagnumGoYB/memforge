package index

import (
	"context"
	"database/sql"
)

type RebuildStats struct {
	Indexed int
	Deleted int
	Orphans int
	Ghosts  int
}

func ResetMemories(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `DELETE FROM memories`)
	return err
}

func RebuildMemories(ctx context.Context, db *sql.DB, records []MemoryRecord) (RebuildStats, error) {
	var stats RebuildStats
	var existingIDs []string
	rows, err := db.QueryContext(ctx, `SELECT id FROM memories`)
	if err != nil {
		return stats, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return stats, err
		}
		existingIDs = append(existingIDs, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return stats, err
	}
	rows.Close()
	existingSet := make(map[string]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		existingSet[id] = struct{}{}
	}
	recordSet := make(map[string]struct{}, len(records))
	for _, record := range records {
		recordSet[record.ID] = struct{}{}
	}
	for _, id := range existingIDs {
		if _, ok := recordSet[id]; !ok {
			stats.Orphans++
		}
	}
	for _, record := range records {
		if _, ok := existingSet[record.ID]; !ok {
			stats.Ghosts++
		}
	}
	stats.Deleted = len(existingIDs)
	if err := ResetMemories(ctx, db); err != nil {
		return stats, err
	}
	for _, record := range records {
		if err := UpsertMemory(ctx, db, record); err != nil {
			return stats, err
		}
		stats.Indexed++
	}
	return stats, nil
}
