package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MagnumGOYB/memforge/internal/buildinfo"
	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/versioncheck"
	"github.com/spf13/cobra"
)

type Streams struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func Execute(args []string, streams Streams) int {
	if streams.Stdin == nil {
		streams.Stdin = strings.NewReader("")
	}
	if streams.Stdout == nil {
		streams.Stdout = io.Discard
	}
	if streams.Stderr == nil {
		streams.Stderr = io.Discard
	}
	root := newRootCmd(streams)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		var cmdErr *commandError
		if errors.As(err, &cmdErr) {
			_, _ = fmt.Fprintln(streams.Stderr, cmdErr.Error())
			return cmdErr.code
		}
		_, _ = fmt.Fprintln(streams.Stderr, err)
		return 3
	}
	return 0
}

func newRootCmd(streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "memforge",
		Short:         "Local-first project memory layer for AI coding agents",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.SetOut(streams.Stdout)
	cmd.SetErr(streams.Stderr)
	cmd.PersistentFlags().String("format", baseconfig.FormatText, "output format: text or json")
	cmd.PersistentFlags().Bool("no-version-check", false, "skip version check")
	cmd.Version = buildinfo.Version()
	cmd.SetVersionTemplate("{{.Version}}\n")

	cmd.AddCommand(newVersionCmd(streams))
	cmd.AddCommand(newInitCmd(streams))
	cmd.AddCommand(newRememberCmd(streams))
	cmd.AddCommand(newSearchCmd(streams))
	cmd.AddCommand(newContextCmd(streams))
	cmd.AddCommand(newBeforeCmd(streams))
	cmd.AddCommand(newAfterCmd(streams))
	cmd.AddCommand(newReindexCmd(streams))
	cmd.AddCommand(newMCPCmd(streams))
	cmd.AddCommand(newDiffSummaryCmd(streams))
	cmd.AddCommand(newDebugCmd(streams))
	wrapVersionCheck(cmd, streams)
	return cmd
}

func wrapVersionCheck(root *cobra.Command, streams Streams) {
	for _, cmd := range root.Commands() {
		wrapVersionCheck(cmd, streams)
	}
	if root.RunE == nil {
		return
	}
	runE := root.RunE
	root.RunE = func(cmd *cobra.Command, args []string) error {
		if err := maybeCheckVersion(cmd, streams); err != nil {
			return err
		}
		return runE(cmd, args)
	}
}

func maybeCheckVersion(cmd *cobra.Command, streams Streams) error {
	settings, err := baseconfig.LoadBase(cmd)
	if err != nil {
		return invalidError("%v", err)
	}
	if settings.NoVersionCheck || strings.EqualFold(strings.TrimSpace(cmd.Name()), "mcp") {
		return nil
	}
	if strings.HasSuffix(os.Args[0], ".test") && os.Getenv("MEMFORGE_VERSION_CHECK_LATEST") == "" && os.Getenv("MEMFORGE_VERSION_CHECK_URL") == "" {
		return nil
	}
	storageRoot, err := baseconfig.ResolveStorageRoot()
	if err != nil {
		return nil
	}
	result, err := versioncheck.Check(context.Background(), storageRoot, buildinfo.Version())
	if err != nil {
		return nil
	}
	if result.HasUpdate() {
		_, _ = fmt.Fprintf(streams.Stderr, "new MemForge version available: %s (current %s)\n", result.Latest, result.Current)
	}
	return nil
}

func newVersionCmd(streams Streams) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the build version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			settings, err := baseconfig.LoadBase(cmd)
			if err != nil {
				return invalidError("%v", err)
			}
			if settings.Format == baseconfig.FormatJSON {
				return internalError(writeJSON(streams.Stdout, map[string]any{"version": buildinfo.Version()}))
			}
			_, err = fmt.Fprintln(streams.Stdout, buildinfo.Version())
			return internalError(err)
		},
	}
}
