package config

import (
	"errors"
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

type ProjectSettings struct {
	DefaultBudget int
	KindWeights   map[string]int
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

func LoadProjectSettings(projectRoot string) (ProjectSettings, error) {
	settings := ProjectSettings{KindWeights: map[string]int{}}
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigName("config")
	if configHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); configHome != "" {
		v.AddConfigPath(filepath.Join(configHome, "memforge"))
	} else if home, err := os.UserHomeDir(); err == nil && home != "" {
		v.AddConfigPath(filepath.Join(home, ".config", "memforge"))
	}
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return ProjectSettings{}, err
		}
	}
	mergeProjectSettings(&settings, v)

	if strings.TrimSpace(projectRoot) != "" {
		projectConfig := viper.New()
		projectConfig.SetConfigFile(filepath.Join(projectRoot, ".memoryrc"))
		projectConfig.SetConfigType("toml")
		if err := projectConfig.ReadInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if !errors.As(err, &notFound) && !os.IsNotExist(err) {
				return ProjectSettings{}, err
			}
		} else {
			mergeProjectSettings(&settings, projectConfig)
		}
	}
	return settings, nil
}

func mergeProjectSettings(settings *ProjectSettings, v *viper.Viper) {
	if v == nil {
		return
	}
	if budget := v.GetInt("default_budget"); budget > 0 {
		settings.DefaultBudget = budget
	}
	for key, value := range v.GetStringMap("kind_weights") {
		weight, ok := value.(int)
		if !ok {
			switch typed := value.(type) {
			case int64:
				weight = int(typed)
				ok = true
			case float64:
				weight = int(typed)
				ok = true
			}
		}
		if ok && weight > 0 {
			settings.KindWeights[strings.TrimSpace(key)] = weight
		}
	}
}
