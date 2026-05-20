package provider

import (
	"context"
	"testing"

	"github.com/MagnumGOYB/memforge/internal/memory"
)

func TestSelectDefaultsToHeuristicExtractor(t *testing.T) {
	extractor, warnings, err := Select("")
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %#v", warnings)
	}
	candidates, err := extractor.ExtractCandidates(context.Background(), "Decision: Use local storage\nMarkdown remains canonical.", ExistingContext{})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 || candidates[0].Kind != memory.KindDecision {
		t.Fatalf("unexpected candidates: %#v", candidates)
	}
}

func TestSelectRejectsUnconfiguredNetworkProvider(t *testing.T) {
	if _, _, err := Select("anthropic"); err == nil {
		t.Fatal("expected unconfigured provider error")
	}
}
