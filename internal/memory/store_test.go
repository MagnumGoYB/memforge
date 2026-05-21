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

func TestLoadRecordsReportsFileContextForMalformedBlock(t *testing.T) {
	memoriesDir := filepath.Join(t.TempDir(), "memories")
	if err := EnsureLayout(memoriesDir); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(memoriesDir, "decisions.md")
	if err := os.WriteFile(path, []byte("<!-- memforge:memory id=x kind=decision -->\nnot-frontmatter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadRecords(memoriesDir, "project-1")
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "decisions.md") || !strings.Contains(err.Error(), "decision") {
		t.Fatalf("missing file context: %v", err)
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

func TestRewriteKindMarkdownKeepsOriginalFileWhenReplacementFails(t *testing.T) {
	memoriesDir := filepath.Join(t.TempDir(), "memories")
	if err := EnsureLayout(memoriesDir); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(memoriesDir, "decisions.md")
	original := []byte("existing decision\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}
	record, err := NewRecord(NewRecordInput{ProjectID: "project-1", Kind: KindDecision, Title: "CLI framework", Content: "Body", Confidence: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(memoriesDir, 0o500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(memoriesDir, 0o755)
	if _, err := RewriteKindMarkdown(memoriesDir, KindDecision, []Record{record}); err == nil {
		t.Fatal("expected rewrite to fail when memories directory is not writable")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(original) {
		t.Fatalf("original file changed after failed rewrite: %q", string(data))
	}
}

func TestRewriteKindMarkdownRemovesTemporaryFileAfterSuccess(t *testing.T) {
	memoriesDir := filepath.Join(t.TempDir(), "memories")
	record, err := NewRecord(NewRecordInput{ProjectID: "project-1", Kind: KindDecision, Title: "CLI framework", Content: "Body", Confidence: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RewriteKindMarkdown(memoriesDir, KindDecision, []Record{record}); err != nil {
		t.Fatal(err)
	}
	matches, err := filepath.Glob(filepath.Join(memoriesDir, "*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary files were not cleaned up: %v", matches)
	}
}
