package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	afterpkg "github.com/MagnumGOYB/memforge/internal/after"
	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/memory"
	"github.com/MagnumGOYB/memforge/internal/project"
	"github.com/MagnumGOYB/memforge/internal/provider"
	"github.com/spf13/cobra"
)

type afterPayload struct {
	Candidates     []afterpkg.Candidate     `json:"candidates"`
	Duplicates     []afterpkg.Duplicate     `json:"duplicates"`
	MergeProposals []afterpkg.MergeProposal `json:"merge_proposals"`
	Persisted      []persistedMemory        `json:"persisted"`
	Warnings       []string                 `json:"warnings"`
}

type persistedMemory struct {
	ID    string      `json:"id"`
	Kind  memory.Kind `json:"kind"`
	Title string      `json:"title"`
}

func newAfterCmd(streams Streams) *cobra.Command {
	var rootOverride string
	var fromFile string
	var approve string
	var providerName string
	var adapter string
	cmd := &cobra.Command{
		Use:   "after --from FILE",
		Short: "Extract candidate memories from a session log",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return invalidError("after accepts flags only")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			settings, err := baseconfig.LoadBase(cmd)
			if err != nil {
				return invalidError("%v", err)
			}
			if strings.TrimSpace(fromFile) == "" {
				return invalidError("after requires --from")
			}
			data, err := os.ReadFile(fromFile)
			if err != nil {
				return userError("read --from file: %v", err)
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
			if err := memory.EnsureLayout(paths.MemoriesDir); err != nil {
				return internalError(err)
			}
			existing, err := memory.LoadRecords(paths.MemoriesDir, proj.ID)
			if err != nil {
				return internalError(err)
			}
			extractor, warnings, err := provider.Select(providerName)
			if err != nil {
				return invalidError("%v", err)
			}
			sessionText, err := afterpkg.ExtractSessionText(adapter, data)
			if err != nil {
				return invalidError("%v", err)
			}
			candidates, err := extractor.ExtractCandidates(context.Background(), sessionText, provider.ExistingContext{Memories: existing})
			if err != nil {
				return internalError(err)
			}
			duplicates := afterpkg.FindDuplicateCandidates(candidates, existing)
			mergeProposals := afterpkg.BuildMergeProposals(candidates, existing)
			approved := afterpkg.ApproveCandidates(candidates, approve, duplicates)
			persisted := make([]persistedMemory, 0, len(approved))
			for _, candidate := range approved {
				record, err := memory.NewRecord(memory.NewRecordInput{
					ProjectID:  proj.ID,
					Kind:       candidate.Kind,
					Title:      candidate.Title,
					Content:    candidate.Content,
					Tags:       candidate.Tags,
					Source:     candidate.Source,
					Confidence: candidate.Confidence,
				})
				if err != nil {
					return invalidError("candidate %s: %v", candidate.ID, err)
				}
				if err := persistMemory(context.Background(), paths, record); err != nil {
					return internalError(err)
				}
				persisted = append(persisted, persistedMemory{ID: record.ID, Kind: record.Kind, Title: record.Title})
			}
			payload := afterPayload{Candidates: candidates, Duplicates: duplicates, MergeProposals: mergeProposals, Persisted: persisted, Warnings: warnings}
			if settings.Format == baseconfig.FormatJSON {
				return internalError(writeJSON(streams.Stdout, payload))
			}
			return writeAfterText(streams.Stdout, payload)
		},
	}
	cmd.Flags().StringVar(&rootOverride, "root", "", "project root override")
	cmd.Flags().StringVar(&fromFile, "from", "", "read session log from file")
	cmd.Flags().StringVar(&approve, "approve", "none", "persist approved candidates: none, all, or comma-separated candidate ids")
	cmd.Flags().StringVar(&providerName, "provider", "heuristic", "candidate extraction provider")
	cmd.Flags().StringVar(&adapter, "adapter", "plain", "session adapter: plain, jsonl, claude-code, codex, cursor")
	return cmd
}

func writeAfterText(w io.Writer, payload afterPayload) error {
	if len(payload.Candidates) == 0 {
		_, err := fmt.Fprintln(w, "No candidate memories found.")
		return internalError(err)
	}
	for _, candidate := range payload.Candidates {
		if _, err := fmt.Fprintf(w, "- [%s] %s (%s)\n", candidate.ID, candidate.Title, candidate.Kind); err != nil {
			return internalError(err)
		}
	}
	if len(payload.Duplicates) > 0 {
		if _, err := fmt.Fprintf(w, "Duplicates: %d\n", len(payload.Duplicates)); err != nil {
			return internalError(err)
		}
	}
	if len(payload.MergeProposals) > 0 {
		if _, err := fmt.Fprintf(w, "Merge proposals: %d\n", len(payload.MergeProposals)); err != nil {
			return internalError(err)
		}
	}
	if len(payload.Persisted) > 0 {
		if _, err := fmt.Fprintf(w, "Persisted: %d\n", len(payload.Persisted)); err != nil {
			return internalError(err)
		}
	}
	return nil
}
