package provider

import (
	"context"
	"fmt"
	"strings"

	afterpkg "github.com/MagnumGOYB/memforge/internal/after"
	"github.com/MagnumGOYB/memforge/internal/memory"
)

type ExistingContext struct {
	Memories []memory.Record
}

type Extractor interface {
	ExtractCandidates(ctx context.Context, sessionText string, existing ExistingContext) ([]afterpkg.Candidate, error)
}

type HeuristicExtractor struct{}

func (HeuristicExtractor) ExtractCandidates(ctx context.Context, sessionText string, existing ExistingContext) ([]afterpkg.Candidate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return afterpkg.ExtractCandidatesFromText(sessionText), nil
}

func Select(name string) (Extractor, []string, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" || name == "heuristic" || name == "none" || name == "disabled" {
		return HeuristicExtractor{}, nil, nil
	}
	switch name {
	case "openai", "openai-compatible", "anthropic", "ollama":
		return nil, []string{fmt.Sprintf("provider %q is recognized but not configured; using heuristic extraction requires --provider heuristic", name)}, fmt.Errorf("provider %q requires explicit adapter configuration", name)
	default:
		return nil, nil, fmt.Errorf("unknown provider %q", name)
	}
}
