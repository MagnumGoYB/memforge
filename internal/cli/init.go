package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/index"
	"github.com/MagnumGOYB/memforge/internal/memory"
	"github.com/MagnumGOYB/memforge/internal/project"
	"github.com/spf13/cobra"
)

func newInitCmd(streams Streams) *cobra.Command {
	var rootOverride string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize local MemForge storage for the current project",
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
			if err := os.MkdirAll(paths.CacheDir, 0o755); err != nil {
				return internalError(err)
			}
			if err := memory.EnsureLayout(paths.MemoriesDir); err != nil {
				return internalError(err)
			}
			if err := os.WriteFile(paths.SchemaVersion, []byte(index.CurrentSchemaVersion+"\n"), 0o644); err != nil {
				return internalError(err)
			}
			meta, err := project.WriteMeta(paths.Meta, project.Meta{
				ProjectID:     proj.ID,
				ProjectRoot:   proj.Root,
				Identifier:    proj.Identifier,
				SchemaVersion: index.CurrentSchemaVersion,
			}, time.Now())
			if err != nil {
				return internalError(err)
			}
			db, err := index.Open(paths.Index)
			if err != nil {
				return internalError(err)
			}
			defer db.Close()

			payload := map[string]any{
				"ok":             true,
				"project_id":     proj.ID,
				"project_root":   proj.Root,
				"project_dir":    paths.ProjectDir,
				"meta":           paths.Meta,
				"index":          paths.Index,
				"memories_dir":   paths.MemoriesDir,
				"schema_version": meta.SchemaVersion,
			}
			if settings.Format == baseconfig.FormatJSON {
				return internalError(writeJSON(streams.Stdout, payload))
			}
			_, err = fmt.Fprintf(streams.Stdout, "Initialized MemForge storage for %s at %s\n", proj.ID, paths.ProjectDir)
			return internalError(err)
		},
	}
	cmd.Flags().StringVar(&rootOverride, "root", "", "project root override")
	return cmd
}

type resolvedPaths struct {
	StorageRoot   string
	ProjectDir    string
	Meta          string
	SchemaVersion string
	MemoriesDir   string
	Index         string
	CacheDir      string
}

func derivePaths(storageRoot string, proj project.Project) resolvedPaths {
	projectDir := filepath.Join(storageRoot, "projects", proj.ID)
	return resolvedPaths{
		StorageRoot:   storageRoot,
		ProjectDir:    projectDir,
		Meta:          filepath.Join(projectDir, "meta.json"),
		SchemaVersion: filepath.Join(projectDir, "schema_version"),
		MemoriesDir:   filepath.Join(projectDir, "memories"),
		Index:         filepath.Join(projectDir, "index.sqlite"),
		CacheDir:      filepath.Join(projectDir, "cache"),
	}
}
