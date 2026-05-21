package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/MagnumGOYB/memforge/internal/compiler"
	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/index"
	"github.com/MagnumGOYB/memforge/internal/memory"
	"github.com/MagnumGOYB/memforge/internal/project"
	"github.com/spf13/cobra"
)

func newBeforeCmd(streams Streams) *cobra.Command {
	var rootOverride string
	var budget int
	cmd := &cobra.Command{
		Use:   "before TASK",
		Short: "Compile task-conditioned project context",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
				return invalidError("before requires exactly one non-empty task argument")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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
			projectSettings, kindWeights, err := loadCompileSettings(proj.Root)
			if err != nil {
				return userError("%v", err)
			}
			resolvedBudget := resolveCLIBudget(cmd, budget, projectSettings)
			paths := derivePaths(storageRoot, proj)
			records, err := memory.LoadRecords(paths.MemoriesDir, proj.ID)
			if err != nil {
				return internalError(err)
			}
			db, err := index.Open(paths.Index)
			if err != nil {
				return internalError(err)
			}
			defer db.Close()
			matches, err := index.SearchMemories(context.Background(), db, index.SearchQuery{ProjectID: proj.ID, Query: args[0], Limit: 20})
			if err != nil && !strings.Contains(err.Error(), "query is required") {
				return internalError(err)
			}
			selected := selectBeforeRecords(records, matches, args[0])
			result := compiler.CompileContext(compiler.CompileInput{Memories: selected, Budget: resolvedBudget, Heading: args[0], KindWeights: kindWeights})
			for _, warning := range result.Warnings {
				_, _ = fmt.Fprintln(streams.Stderr, warning)
			}
			payload := map[string]any{
				"task":             args[0],
				"budget":           resolvedBudget,
				"count":            len(result.Entries),
				"context":          result.Markdown,
				"warnings":         result.Warnings,
				"estimated_tokens": result.EstimatedTokens,
				"usage": map[string]any{
					"estimated_tokens": result.EstimatedTokens,
				},
			}
			if settings.Format == baseconfig.FormatJSON {
				return internalError(writeJSON(streams.Stdout, payload))
			}
			_, err = fmt.Fprint(streams.Stdout, result.Markdown)
			return internalError(err)
		},
	}
	cmd.Flags().StringVar(&rootOverride, "root", "", "project root override")
	cmd.Flags().IntVar(&budget, "budget", 0, "token budget")
	return cmd
}

func selectBeforeRecords(records []memory.Record, matches []index.SearchResult, task string) []memory.Record {
	selected := make([]memory.Record, 0)
	seen := map[string]struct{}{}
	taskTerms := strings.Fields(strings.ToLower(task))
	for _, record := range records {
		if record.Kind == memory.KindManual || record.Kind == memory.KindConstraint {
			selected = append(selected, record)
			seen[record.ID] = struct{}{}
		}
	}
	byID := make(map[string]memory.Record, len(records))
	for _, record := range records {
		byID[record.ID] = record
	}
	for _, match := range matches {
		if _, ok := seen[match.ID]; ok {
			continue
		}
		record, ok := byID[match.ID]
		if !ok {
			continue
		}
		selected = append(selected, record)
		seen[record.ID] = struct{}{}
	}
	for _, record := range records {
		if _, ok := seen[record.ID]; ok {
			continue
		}
		text := strings.ToLower(record.Title + " " + record.Content + " " + strings.Join(record.Tags, " "))
		for _, term := range taskTerms {
			if len(term) < 4 {
				continue
			}
			if strings.Contains(text, term) {
				selected = append(selected, record)
				seen[record.ID] = struct{}{}
				break
			}
		}
	}
	return selected
}
