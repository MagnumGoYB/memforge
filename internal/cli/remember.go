package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/memory"
	"github.com/MagnumGOYB/memforge/internal/project"
	"github.com/spf13/cobra"
)

func newRememberCmd(streams Streams) *cobra.Command {
	var rootOverride string
	var kindValue string
	var title string
	var tags []string
	var source string
	var confidence float64
	var fromFile string

	cmd := &cobra.Command{
		Use:   "remember [body|-]",
		Short: "Persist a project memory locally",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return invalidError("remember accepts exactly one positional content argument")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			settings, err := baseconfig.LoadBase(cmd)
			if err != nil {
				return invalidError("%v", err)
			}
			content, err := resolveRememberContent(args, fromFile, streams.Stdin)
			if err != nil {
				return err
			}
			kind, err := memory.ParseKind(kindValue)
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
			record, err := memory.NewRecord(memory.NewRecordInput{
				ProjectID:  proj.ID,
				Kind:       kind,
				Title:      title,
				Content:    content,
				Tags:       tags,
				Source:     source,
				Confidence: confidence,
			})
			if err != nil {
				return invalidError("%v", err)
			}
			_, warning, err := persistMemory(context.Background(), paths, record)
			if err != nil {
				return internalError(err)
			}
			if warning != "" {
				_, _ = fmt.Fprintln(streams.Stderr, warning)
			}
			payload := map[string]any{"id": record.ID, "kind": record.Kind, "title": record.Title}
			if warning != "" {
				payload["warning"] = warning
			}
			if settings.Format == baseconfig.FormatJSON {
				return internalError(writeJSON(streams.Stdout, payload))
			}
			_, err = fmt.Fprintf(streams.Stdout, "Remembered %s %s %q\n", record.ID, record.Kind, record.Title)
			return internalError(err)
		},
	}
	cmd.Flags().StringVar(&rootOverride, "root", "", "project root override")
	cmd.Flags().StringVar(&kindValue, "kind", string(memory.KindManual), "memory kind")
	cmd.Flags().StringVar(&title, "title", "", "memory title")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "memory tag")
	cmd.Flags().StringVar(&source, "source", "", "memory source")
	cmd.Flags().Float64Var(&confidence, "confidence", 1.0, "memory confidence")
	cmd.Flags().StringVar(&fromFile, "from", "", "read memory content from file")
	return cmd
}

func resolveRememberContent(args []string, fromFile string, stdin io.Reader) (string, error) {
	hasArg := len(args) == 1
	hasFile := strings.TrimSpace(fromFile) != ""
	if hasArg && hasFile {
		return "", invalidError("content source must be exactly one of positional body, '-', or --from")
	}
	if !hasArg && !hasFile {
		return "", invalidError("content source is required")
	}
	if hasFile {
		data, err := os.ReadFile(fromFile)
		if err != nil {
			return "", userError("read --from file: %v", err)
		}
		return string(data), nil
	}
	if args[0] == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", internalError(err)
		}
		return string(data), nil
	}
	return args[0], nil
}
