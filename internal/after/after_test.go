package after

import (
	"testing"
	"time"

	"github.com/MagnumGOYB/memforge/internal/memory"
)

func TestExtractCandidatesFromTextFindsKindBlocks(t *testing.T) {
	candidates := ExtractCandidatesFromText(`Assistant: done

Decision: Repository layer stays framework-agnostic
The repository package must not depend on Cobra or CLI types.

Constraint: JSON stdout stays machine-readable
Warnings must go to stderr.`)
	if len(candidates) != 2 {
		t.Fatalf("got %d candidates: %#v", len(candidates), candidates)
	}
	if candidates[0].ID != "cand_1" || candidates[0].Kind != memory.KindDecision || candidates[0].Title != "Repository layer stays framework-agnostic" {
		t.Fatalf("unexpected first candidate: %#v", candidates[0])
	}
	if candidates[1].Kind != memory.KindConstraint || candidates[1].Content == "" {
		t.Fatalf("unexpected second candidate: %#v", candidates[1])
	}
}

func TestFindDuplicateCandidatesMatchesExistingMemory(t *testing.T) {
	existing := []memory.Record{{
		ID:        "mem_1",
		Kind:      memory.KindDecision,
		Title:     "Repository layer stays framework-agnostic",
		Content:   "The repository package must not depend on Cobra or CLI types.",
		UpdatedAt: time.Now(),
	}}
	candidates := []Candidate{{ID: "cand_1", Kind: memory.KindDecision, Title: existing[0].Title, Content: existing[0].Content}}
	duplicates := FindDuplicateCandidates(candidates, existing)
	if len(duplicates) != 1 || duplicates[0].CandidateID != "cand_1" || duplicates[0].MemoryID != "mem_1" {
		t.Fatalf("unexpected duplicates: %#v", duplicates)
	}
	approved := ApproveCandidates(candidates, "all", duplicates)
	if len(approved) != 0 {
		t.Fatalf("duplicates must not be auto-persisted: %#v", approved)
	}
}

func TestBuildMergeProposalsFindsRelatedMemory(t *testing.T) {
	existing := []memory.Record{{
		ID:      "mem_1",
		Kind:    memory.KindDecision,
		Title:   "Repository layer stays framework-agnostic",
		Content: "Keep storage independent from CLI framework.",
		Tags:    []string{"repository", "architecture"},
	}}
	candidates := []Candidate{{
		ID:      "cand_1",
		Kind:    memory.KindDecision,
		Title:   "Repository layer remains framework-agnostic",
		Content: "Do not import Cobra from repository packages.",
		Tags:    []string{"repository", "architecture"},
	}}
	proposals := BuildMergeProposals(candidates, existing)
	if len(proposals) != 1 || proposals[0].CandidateID != "cand_1" || proposals[0].MemoryID != "mem_1" {
		t.Fatalf("unexpected proposals: %#v", proposals)
	}
}
