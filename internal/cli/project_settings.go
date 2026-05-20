package cli

import (
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
	return settings, parseKindWeights(settings.KindWeights), nil
}

func parseKindWeights(values map[string]int) map[memory.Kind]int {
	if len(values) == 0 {
		return nil
	}
	weights := make(map[memory.Kind]int, len(values))
	for key, value := range values {
		if value <= 0 {
			continue
		}
		kind, err := memory.ParseKind(key)
		if err != nil {
			continue
		}
		weights[kind] = value
	}
	if len(weights) == 0 {
		return nil
	}
	return weights
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
