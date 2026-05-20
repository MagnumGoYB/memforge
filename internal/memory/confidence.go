package memory

import (
	"math"
	"time"
)

func EffectiveConfidence(kind Kind, base float64, updatedAt time.Time, now time.Time) float64 {
	if base <= 0 {
		return 0
	}
	if base > 1 {
		base = 1
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if updatedAt.IsZero() || now.Before(updatedAt) {
		return base
	}
	halfLifeDays := confidenceHalfLifeDays(kind)
	if halfLifeDays <= 0 {
		return base
	}
	days := now.Sub(updatedAt).Hours() / 24
	return base * math.Pow(0.5, days/halfLifeDays)
}

func confidenceHalfLifeDays(kind Kind) float64 {
	switch kind {
	case KindManual, KindConstraint:
		return 0
	case KindConvention, KindAPIContract:
		return 365
	case KindDecision:
		return 180
	case KindBugfix, KindAgentInstruction:
		return 90
	default:
		return 180
	}
}
