package memory

import "testing"

func TestParseKindAcceptsKnownKinds(t *testing.T) {
	for _, value := range []string{"manual", "constraint", "convention", "decision", "bugfix", "api-contract", "agent-instruction"} {
		if _, err := ParseKind(value); err != nil {
			t.Fatalf("unexpected error for %q: %v", value, err)
		}
	}
}

func TestParseKindAcceptsAgentFriendlyAliases(t *testing.T) {
	tests := map[string]Kind{
		"workflow": KindAgentInstruction,
		"note":     KindManual,
		"domain":   KindConvention,
	}
	for value, want := range tests {
		got, err := ParseKind(value)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", value, err)
		}
		if got != want {
			t.Fatalf("ParseKind(%q)=%q, want %q", value, got, want)
		}
	}
}

func TestParseKindRejectsUnknownKind(t *testing.T) {
	if _, err := ParseKind("unknown"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFileNameForKind(t *testing.T) {
	file, ok := FileNameForKind(KindDecision)
	if !ok || file != "decisions.md" {
		t.Fatalf("got %q ok=%v", file, ok)
	}
}
