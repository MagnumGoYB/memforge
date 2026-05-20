package project

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCanonicalizeIdentifierStripsCredentialsAndDotGit(t *testing.T) {
	got := CanonicalizeIdentifier("https://token@example.com/Owner/Repo.git")
	want := "https://example.com/Owner/Repo"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestCanonicalizeIdentifierSupportsSCPLikeURL(t *testing.T) {
	got := CanonicalizeIdentifier("git@github.com:MagnumGOYB/memforge.git")
	want := "https://github.com/magnumgoyb/memforge"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestCanonicalizeIdentifierNormalizesGitHubPathCase(t *testing.T) {
	left := CanonicalizeIdentifier("https://github.com/MagnumGOYB/memforge.git")
	right := CanonicalizeIdentifier("git@github.com:magnumgoyb/MemForge.git")
	if left != right {
		t.Fatalf("github identifiers should match: %q != %q", left, right)
	}
}

func TestHashIdentifierIsDeterministic(t *testing.T) {
	if HashIdentifier("abc") != HashIdentifier("abc") {
		t.Fatal("expected deterministic hash")
	}
}

func TestWriteMetaPreservesCreatedAt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meta.json")
	first, err := WriteMeta(path, Meta{ProjectID: "p1", ProjectRoot: "/tmp/p1", Identifier: "id1", SchemaVersion: "1"}, time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	second, err := WriteMeta(path, Meta{ProjectID: "p1", ProjectRoot: "/tmp/p1", Identifier: "id1", SchemaVersion: "1"}, time.Date(2026, 5, 19, 11, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !first.CreatedAt.Equal(second.CreatedAt) {
		t.Fatalf("created_at changed: %s != %s", first.CreatedAt, second.CreatedAt)
	}
	if !second.UpdatedAt.After(first.UpdatedAt) {
		t.Fatalf("updated_at did not advance: %s <= %s", second.UpdatedAt, first.UpdatedAt)
	}
}
