package cli

import (
	"context"
	"fmt"
	"strings"

	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/index"
	"github.com/MagnumGOYB/memforge/internal/project"
	"github.com/spf13/cobra"
)

func newSearchCmd(streams Streams) *cobra.Command {
	var rootOverride string
	var kindsValue string
	var tags []string
	var limit int
	var hybrid bool
	cmd := &cobra.Command{
		Use:   "search QUERY",
		Short: "Search project memories",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
				return invalidError("search requires exactly one non-empty query argument")
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
			db, err := index.Open(derivePaths(storageRoot, proj).Index)
			if err != nil {
				return internalError(err)
			}
			defer db.Close()
			results, err := index.SearchMemories(context.Background(), db, index.SearchQuery{
				ProjectID: proj.ID,
				Query:     args[0],
				Kinds:     splitCSV(kindsValue),
				Limit:     limit,
				Hybrid:    hybrid,
			})
			if err != nil {
				return userError("%v", err)
			}
			results = filterSearchResultsByTags(results, tags)
			payload := map[string]any{
				"query":   args[0],
				"count":   len(results),
				"hybrid":  hybrid,
				"results": results,
			}
			if settings.Format == baseconfig.FormatJSON {
				return internalError(writeJSON(streams.Stdout, payload))
			}
			if len(results) == 0 {
				_, err = fmt.Fprintln(streams.Stdout, "No memories found.")
				return internalError(err)
			}
			for _, result := range results {
				if _, err := fmt.Fprintf(streams.Stdout, "- [%s] %s\n  tags=%s\n  score=%.3f\n  %s\n", result.Kind, result.Title, strings.Join(result.Tags, ", "), result.Score, result.Snippet); err != nil {
					return internalError(err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&rootOverride, "root", "", "project root override")
	cmd.Flags().StringVar(&kindsValue, "kind", "", "comma-separated memory kinds")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "required tag filter")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum number of results")
	cmd.Flags().BoolVar(&hybrid, "hybrid", false, "rerank results with local deterministic embeddings")
	return cmd
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func filterSearchResultsByTags(results []index.SearchResult, required []string) []index.SearchResult {
	required = normalizeRequiredTags(required)
	if len(required) == 0 {
		return results
	}
	filtered := make([]index.SearchResult, 0, len(results))
	for _, result := range results {
		tagSet := map[string]struct{}{}
		for _, tag := range result.Tags {
			tagSet[tag] = struct{}{}
		}
		match := true
		for _, tag := range required {
			if _, ok := tagSet[tag]; !ok {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func normalizeRequiredTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}
