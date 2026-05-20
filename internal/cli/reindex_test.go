package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteReindexJSONRebuildsIndex(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	var sink bytes.Buffer
	if code := Execute([]string{"remember", "Repository layer is framework-agnostic", "--root", projectRoot, "--kind", "decision", "--title", "Repository decision", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &sink, Stderr: &bytes.Buffer{}}); code != 0 {
		t.Fatalf("remember code=%d", code)
	}
	projDir := filepath.Join(storage, "projects")
	entries, err := os.ReadDir(projDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("read project dir: %v len=%d", err, len(entries))
	}
	indexPath := filepath.Join(projDir, entries[0].Name(), "index.sqlite")
	if err := os.Remove(indexPath); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"reindex", "--root", projectRoot, "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Indexed int `json:"indexed"`
		Ghosts  int `json:"ghosts"`
		Orphans int `json:"orphans"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Indexed != 1 || payload.Ghosts != 1 || payload.Orphans != 0 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	searchOut := bytes.Buffer{}
	searchErr := bytes.Buffer{}
	code = Execute([]string{"search", "repository framework", "--root", projectRoot, "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &searchOut, Stderr: &searchErr})
	if code != 0 {
		t.Fatalf("search code=%d stderr=%s", code, searchErr.String())
	}
}
