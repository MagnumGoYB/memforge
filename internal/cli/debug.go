package cli

import (
	"fmt"

	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/project"
	"github.com/spf13/cobra"
)

func newDebugCmd(streams Streams) *cobra.Command {
	var rootOverride string
	debugCmd := &cobra.Command{Use: "debug", Hidden: true}
	pathsCmd := &cobra.Command{
		Use:   "paths",
		Short: "Print resolved MemForge paths",
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
			payload := map[string]any{
				"storage_root":   paths.StorageRoot,
				"project_root":   proj.Root,
				"project_id":     proj.ID,
				"project_dir":    paths.ProjectDir,
				"meta":           paths.Meta,
				"schema_version": paths.SchemaVersion,
				"memories_dir":   paths.MemoriesDir,
				"index":          paths.Index,
				"cache_dir":      paths.CacheDir,
			}
			if settings.Format == baseconfig.FormatJSON {
				return internalError(writeJSON(streams.Stdout, payload))
			}
			_, err = fmt.Fprintf(streams.Stdout, "project=%s\nroot=%s\nstorage=%s\n", proj.ID, proj.Root, paths.ProjectDir)
			return internalError(err)
		},
	}
	pathsCmd.Flags().StringVar(&rootOverride, "root", "", "project root override")
	debugCmd.AddCommand(pathsCmd)
	return debugCmd
}
