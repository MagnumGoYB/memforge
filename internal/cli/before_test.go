package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteBeforeJSON(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	for _, args := range [][]string{
		{"remember", "Always keep repository framework-agnostic.", "--root", projectRoot, "--kind", "constraint", "--title", "Repository constraint", "--format", "json"},
		{"remember", "Auth middleware currently depends on repository interfaces.", "--root", projectRoot, "--kind", "decision", "--title", "Auth middleware decision", "--format", "json"},
	} {
		if code := Execute(args, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); code != 0 {
			t.Fatalf("remember code=%d args=%v", code, args)
		}
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"before", "Refactor auth middleware", "--root", projectRoot, "--budget", "200", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Task            string `json:"task"`
		Count           int    `json:"count"`
		Context         string `json:"context"`
		EstimatedTokens int    `json:"estimated_tokens"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Task != "Refactor auth middleware" || payload.Count == 0 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if !strings.Contains(payload.Context, "# Refactor auth middleware") {
		t.Fatalf("unexpected context: %s", payload.Context)
	}
	if payload.EstimatedTokens <= 0 {
		t.Fatalf("unexpected estimated tokens: %#v", payload)
	}
}
