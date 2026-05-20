package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendMarkdownWritesToKindFile(t *testing.T) {
	memoriesDir := filepath.Join(t.TempDir(), "memories")
	record, err := NewRecord(NewRecordInput{
		ProjectID:  "project-1",
		Kind:       KindDecision,
		Title:      "CLI framework",
		Content:    "Body",
		Tags:       []string{"cli"},
		Confidence: 1,
		Now:        time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := AppendMarkdown(memoriesDir, record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(result.Path, "decisions.md") {
		t.Fatalf("got %s", result.Path)
	}
	data, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "CLI framework") {
		t.Fatalf("unexpected file content: %s", string(data))
	}
}

func TestAppendMarkdownPreservesExistingContent(t *testing.T) {
	memoriesDir := filepath.Join(t.TempDir(), "memories")
	if err := EnsureLayout(memoriesDir); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(memoriesDir, "decisions.md")
	if err := os.WriteFile(path, []byte("existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	record, err := NewRecord(NewRecordInput{ProjectID: "project-1", Kind: KindDecision, Title: "CLI framework", Content: "Body", Confidence: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendMarkdown(memoriesDir, record); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "existing") {
		t.Fatalf("existing content lost: %s", string(data))
	}
}
