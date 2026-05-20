package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestExecuteContextJSON(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	for _, args := range [][]string{
		{"remember", "Manual note body", "--root", projectRoot, "--kind", "manual", "--title", "Manual note", "--format", "json"},
		{"remember", "Decision note body", "--root", projectRoot, "--kind", "decision", "--title", "Decision note", "--format", "json"},
	} {
		if code := Execute(args, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); code != 0 {
			t.Fatalf("remember code=%d args=%v", code, args)
		}
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"context", "--root", projectRoot, "--budget", "200", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Count    int      `json:"count"`
		Context  string   `json:"context"`
		Warnings []string `json:"warnings"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Count == 0 || payload.Context == "" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
