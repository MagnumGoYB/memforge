package cli

import (
	"fmt"
	"strings"

	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/memory"
	"github.com/spf13/cobra"
)

const defaultContextBudget = 3000

func loadCompileSettings(projectRoot string) (baseconfig.ProjectSettings, map[memory.Kind]int, error) {
	settings, err := baseconfig.LoadProjectSettings(projectRoot)
	if err != nil {
		return baseconfig.ProjectSettings{}, nil, err
	}
	weights, err := parseKindWeights(settings.KindWeights)
	if err != nil {
		return baseconfig.ProjectSettings{}, nil, err
	}
	return settings, weights, nil
}

func parseKindWeights(values map[string]int) (map[memory.Kind]int, error) {
	if len(values) == 0 {
		return nil, nil
	}
	weights := make(map[memory.Kind]int, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		if value <= 0 {
			continue
		}
		kind, err := memory.ParseKind(key)
		if err != nil {
			return nil, fmt.Errorf("invalid kind_weights key %q: %w", key, err)
		}
		weights[kind] = value
	}
	if len(weights) == 0 {
		return nil, nil
	}
	return weights, nil
}

func resolveCLIBudget(cmd *cobra.Command, flagValue int, settings baseconfig.ProjectSettings) int {
	if cmd.Flags().Changed("budget") {
		return flagValue
	}
	return resolveBudget(flagValue, settings)
}

func resolveBudget(value int, settings baseconfig.ProjectSettings) int {
	if value > 0 {
		return value
	}
	if settings.DefaultBudget > 0 {
		return settings.DefaultBudget
	}
	return defaultContextBudget
}
