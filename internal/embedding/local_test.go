package embedding

import "testing"

func TestEmbedIsDeterministicAndNormalized(t *testing.T) {
	left := Embed("repository framework architecture")
	right := Embed("repository framework architecture")
	if left != right {
		t.Fatal("expected deterministic vectors")
	}
	if got := Cosine(left, right); got < 0.99 || got > 1.01 {
		t.Fatalf("unexpected cosine %f", got)
	}
}

func TestEmbedScoresRelatedTextAboveUnrelatedText(t *testing.T) {
	query := Embed("repository framework")
	related := Embed("framework agnostic repository layer")
	unrelated := Embed("token budget optimizer")
	if Cosine(query, related) <= Cosine(query, unrelated) {
		t.Fatalf("expected related text to score higher")
	}
}
