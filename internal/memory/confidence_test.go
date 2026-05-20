package memory

import (
	"testing"
	"time"
)

func TestEffectiveConfidenceDoesNotDecayManualOrConstraint(t *testing.T) {
	now := time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC)
	old := now.AddDate(-5, 0, 0)
	if got := EffectiveConfidence(KindManual, 0.9, old, now); got != 0.9 {
		t.Fatalf("manual confidence decayed to %f", got)
	}
	if got := EffectiveConfidence(KindConstraint, 0.8, old, now); got != 0.8 {
		t.Fatalf("constraint confidence decayed to %f", got)
	}
}

func TestEffectiveConfidenceDecaysBugfixByHalfLife(t *testing.T) {
	now := time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC)
	updated := now.AddDate(0, 0, -90)
	got := EffectiveConfidence(KindBugfix, 1, updated, now)
	if got < 0.49 || got > 0.51 {
		t.Fatalf("got %f", got)
	}
}
