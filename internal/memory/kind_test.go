package memory

import "testing"

func TestParseKindAcceptsKnownKinds(t *testing.T) {
	for _, value := range []string{"manual", "constraint", "convention", "decision", "bugfix", "api-contract", "agent-instruction"} {
		if _, err := ParseKind(value); err != nil {
			t.Fatalf("unexpected error for %q: %v", value, err)
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
