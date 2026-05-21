package compiler

import (
	"strings"
	"testing"
	"time"

	"github.com/MagnumGOYB/memforge/internal/memory"
)

func TestCompileContextPrioritizesManualAndConstraint(t *testing.T) {
	result := CompileContext(CompileInput{
		Budget: 3000,
		Memories: []memory.Record{
			{ID: "1", Kind: memory.KindDecision, Title: "Decision", Content: "decision body", Confidence: 0.8, UpdatedAt: time.Now().Add(-48 * time.Hour)},
			{ID: "2", Kind: memory.KindManual, Title: "Manual", Content: "manual body", Confidence: 1, UpdatedAt: time.Now()},
			{ID: "3", Kind: memory.KindConstraint, Title: "Constraint", Content: "constraint body", Confidence: 1, UpdatedAt: time.Now()},
		},
	})
	if len(result.Entries) < 2 {
		t.Fatalf("expected entries, got %#v", result)
	}
	if result.Entries[0].Record.Kind != memory.KindManual || result.Entries[1].Record.Kind != memory.KindConstraint {
		t.Fatalf("unexpected order: %#v", result.Entries)
	}
	if !strings.Contains(result.Markdown, "# Project Context") {
		t.Fatalf("unexpected markdown: %s", result.Markdown)
	}
	if result.Markdown == "" {
		t.Fatal("expected markdown output")
	}
}

func TestCompileContextUsesCustomKindWeights(t *testing.T) {
	now := time.Now()
	result := CompileContext(CompileInput{
		Budget: 3000,
		KindWeights: map[memory.Kind]int{
			memory.KindDecision: 10,
			memory.KindBugfix:   100,
		},
		Memories: []memory.Record{
			{ID: "1", Kind: memory.KindDecision, Title: "Decision", Content: "decision body", Confidence: 1, UpdatedAt: now},
			{ID: "2", Kind: memory.KindBugfix, Title: "Bugfix", Content: "bugfix body", Confidence: 1, UpdatedAt: now},
		},
	})
	if len(result.Entries) < 2 {
		t.Fatalf("expected entries, got %#v", result)
	}
	if result.Entries[0].Record.Kind != memory.KindBugfix {
		t.Fatalf("custom kind weights were not applied: %#v", result.Entries)
	}
}

func TestCompileContextReportsEstimatedTokens(t *testing.T) {
	result := CompileContext(CompileInput{
		Budget: 3000,
		Memories: []memory.Record{
			{ID: "1", Kind: memory.KindManual, Title: "Manual", Content: "manual body", Confidence: 1, UpdatedAt: time.Now()},
			{ID: "2", Kind: memory.KindDecision, Title: "Decision", Content: "decision body", Confidence: 1, UpdatedAt: time.Now()},
		},
	})
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %#v", result.Entries)
	}
	expected := result.Entries[0].Tokens + result.Entries[1].Tokens
	if result.EstimatedTokens != expected {
		t.Fatalf("estimated tokens=%d, want %d", result.EstimatedTokens, expected)
	}
}
