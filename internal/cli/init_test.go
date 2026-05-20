package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteInitJSON(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"init", "--root", projectRoot, "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	metaPath := payload["meta"].(string)
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("meta missing: %v", err)
	}
	indexPath := payload["index"].(string)
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("index missing: %v", err)
	}
}

func TestExecuteInitRejectsMissingRoot(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	missingRoot := filepath.Join(t.TempDir(), "missing")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"init", "--root", missingRoot, "--format", "json", "--no-version-check"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 1 {
		t.Fatalf("got code %d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(storage, "projects")); !os.IsNotExist(err) {
		t.Fatalf("missing root must not create storage projects dir: %v", err)
	}
}
