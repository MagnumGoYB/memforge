package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	FormatText = "text"
	FormatJSON = "json"
)

type BaseSettings struct {
	Format         string
	NoVersionCheck bool
}

func LoadBase(cmd *cobra.Command) (BaseSettings, error) {
	v := viper.New()
	v.SetEnvPrefix("MEMFORGE")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	_ = v.BindPFlag("format", cmd.Flag("format"))
	_ = v.BindPFlag("no-version-check", cmd.Flag("no-version-check"))
	_ = v.BindEnv("no-version-check", "MEMFORGE_NO_VERSION_CHECK")

	settings := BaseSettings{
		Format:         strings.TrimSpace(v.GetString("format")),
		NoVersionCheck: v.GetBool("no-version-check"),
	}
	if settings.Format == "" {
		settings.Format = FormatText
	}
	if settings.Format != FormatText && settings.Format != FormatJSON {
		return BaseSettings{}, fmt.Errorf("invalid --format %q: must be text or json", settings.Format)
	}
	return settings, nil
}

func ResolveStorageRoot() (string, error) {
	memforgeHome := strings.TrimSpace(os.Getenv("MEMFORGE_HOME"))
	if memforgeHome != "" {
		if !filepath.IsAbs(memforgeHome) {
			return "", fmt.Errorf("MEMFORGE_HOME must be an absolute path")
		}
		return filepath.Clean(memforgeHome), nil
	}

	xdgDataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
	if xdgDataHome != "" {
		if !filepath.IsAbs(xdgDataHome) {
			return "", fmt.Errorf("XDG_DATA_HOME must be an absolute path")
		}
		return filepath.Join(filepath.Clean(xdgDataHome), "memforge"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	if !filepath.IsAbs(homeDir) {
		return "", fmt.Errorf("home directory must be an absolute path")
	}
	return filepath.Join(filepath.Clean(homeDir), ".local", "share", "memforge"), nil
}
