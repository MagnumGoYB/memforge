package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSummarizeNumstat(t *testing.T) {
	payload := summarizeNumstat("2\t1\tinternal/cli/root.go\n5\t0\tdocs/mcp.md\n-\t-\tassets/logo.png\n")
	if payload.Files != 3 || payload.Added != 7 || payload.Deleted != 1 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestExecuteDiffSummaryFromFileJSON(t *testing.T) {
	input := filepath.Join(t.TempDir(), "numstat.txt")
	if err := os.WriteFile(input, []byte("2\t1\tinternal/cli/root.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"diff-summary", "--from", input, "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload diffSummaryPayload
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Files != 1 || payload.Summaries[0].Path != "internal/cli/root.go" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
