package cli

import (
	"fmt"
	"strings"

	"github.com/MagnumGOYB/memforge/internal/compiler"
	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/memory"
	"github.com/MagnumGOYB/memforge/internal/project"
	"github.com/spf13/cobra"
)

func newContextCmd(streams Streams) *cobra.Command {
	var rootOverride string
	var budget int
	var kindsValue string
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Compile agent-ready project context",
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
			records, err := memory.LoadRecords(derivePaths(storageRoot, proj).MemoriesDir, proj.ID)
			if err != nil {
				return internalError(err)
			}
			kinds, err := parseKinds(splitCSV(kindsValue))
			if err != nil {
				return invalidError("%v", err)
			}
			result := compiler.CompileContext(compiler.CompileInput{Memories: records, Budget: budget, Kinds: kinds})
			for _, warning := range result.Warnings {
				_, _ = fmt.Fprintln(streams.Stderr, warning)
			}
			payload := map[string]any{"budget": budget, "count": len(result.Entries), "context": result.Markdown, "warnings": result.Warnings}
			if settings.Format == baseconfig.FormatJSON {
				return internalError(writeJSON(streams.Stdout, payload))
			}
			_, err = fmt.Fprint(streams.Stdout, result.Markdown)
			return internalError(err)
		},
	}
	cmd.Flags().StringVar(&rootOverride, "root", "", "project root override")
	cmd.Flags().IntVar(&budget, "budget", 3000, "token budget")
	cmd.Flags().StringVar(&kindsValue, "kinds", "", "comma-separated memory kinds")
	return cmd
}

func parseKinds(values []string) ([]memory.Kind, error) {
	if len(values) == 0 {
		return nil, nil
	}
	kinds := make([]memory.Kind, 0, len(values))
	for _, value := range values {
		kind, err := memory.ParseKind(strings.TrimSpace(value))
		if err != nil {
			return nil, err
		}
		kinds = append(kinds, kind)
	}
	return kinds, nil
}
