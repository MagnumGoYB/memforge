package cli

import (
	"context"
	"fmt"
	"os"

	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/index"
	"github.com/MagnumGOYB/memforge/internal/memory"
	"github.com/MagnumGOYB/memforge/internal/project"
	"github.com/spf13/cobra"
)

func newReindexCmd(streams Streams) *cobra.Command {
	var rootOverride string
	cmd := &cobra.Command{
		Use:   "reindex",
		Short: "Rebuild SQLite index from markdown memories",
		RunE: func(cmd *cobra.Command, _ []string) error {
			settings, err := baseconfig.LoadBase(cmd)
			if err != nil {
				return invalidError("%v", err)
			}
			storageRoot, err := baseconfig.ResolveStorageRoot()
			if err != nil {
				return userError("%v", err)
			}
			proj, err := project.Detect(rootOverride)
			if err != nil {
				return userError("%v", err)
			}
			paths := derivePaths(storageRoot, proj)
			if err := os.MkdirAll(paths.ProjectDir, 0o755); err != nil {
				return internalError(err)
			}
			if err := memory.EnsureLayout(paths.MemoriesDir); err != nil {
				return internalError(err)
			}
			records, err := memory.LoadRecords(paths.MemoriesDir, proj.ID)
			if err != nil {
				return internalError(err)
			}
			db, err := index.Open(paths.Index)
			if err != nil {
				return internalError(err)
			}
			defer db.Close()
			memoryRecords := make([]index.MemoryRecord, 0, len(records))
			for _, record := range records {
				memoryRecords = append(memoryRecords, index.MemoryRecord{
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
				})
			}
			stats, err := index.RebuildMemories(context.Background(), db, memoryRecords)
			if err != nil {
				return internalError(err)
			}
			payload := map[string]any{
				"indexed": stats.Indexed,
				"deleted": stats.Deleted,
				"orphans": stats.Orphans,
				"ghosts":  stats.Ghosts,
			}
			if settings.Format == baseconfig.FormatJSON {
				return internalError(writeJSON(streams.Stdout, payload))
			}
			_, err = fmt.Fprintf(streams.Stdout, "Reindexed %d memories\n", stats.Indexed)
			return internalError(err)
		},
	}
	cmd.Flags().StringVar(&rootOverride, "root", "", "project root override")
	return cmd
}
