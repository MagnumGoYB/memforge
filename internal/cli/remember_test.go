package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MagnumGOYB/memforge/internal/index"
	"github.com/MagnumGOYB/memforge/internal/project"
)

func TestExecuteRememberJSONPersistsMarkdownAndIndex(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"remember", "Use Cobra", "--root", projectRoot, "--kind", "decision", "--title", "CLI framework", "--tag", "cli", "--tag", "architecture", "--source", "planning", "--confidence", "0.9", "--format", "json"}, Streams{Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["kind"] != "decision" || payload["title"] != "CLI framework" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	proj, err := project.Detect(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	paths := derivePaths(storage, proj)
	data, err := os.ReadFile(filepath.Join(paths.MemoriesDir, "decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "CLI framework") {
		t.Fatalf("memory file missing entry: %s", string(data))
	}
	db, err := index.Open(paths.Index)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var title string
	if err := db.QueryRow(`SELECT title FROM memories WHERE id = ?`, payload["id"]).Scan(&title); err != nil {
		t.Fatal(err)
	}
	if title != "CLI framework" {
		t.Fatalf("got %q", title)
	}
}

func TestExecuteRememberRejectsMultipleContentSources(t *testing.T) {
	projectRoot := t.TempDir()
	inputFile := filepath.Join(t.TempDir(), "note.md")
	if err := os.WriteFile(inputFile, []byte("Body"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"remember", "Body", "--root", projectRoot, "--from", inputFile, "--kind", "manual", "--title", "Title"}, Streams{Stdout: &stdout, Stderr: &stderr})
	if code != 2 {
		t.Fatalf("got code %d stderr=%s", code, stderr.String())
	}
}

func TestExecuteRememberFromFile(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	inputFile := filepath.Join(t.TempDir(), "note.md")
	if err := os.WriteFile(inputFile, []byte("File body"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"remember", "--root", projectRoot, "--from", inputFile, "--kind", "manual", "--title", "Title", "--format", "json"}, Streams{Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}

func TestExecuteRememberFromStdin(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"remember", "-", "--root", projectRoot, "--kind", "manual", "--title", "Pipe", "--format", "json"}, Streams{Stdin: strings.NewReader("Body via stdin"), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}
