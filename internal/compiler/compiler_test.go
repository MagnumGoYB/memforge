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
