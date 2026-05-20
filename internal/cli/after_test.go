package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/MagnumGOYB/memforge/internal/memory"
	"github.com/MagnumGOYB/memforge/internal/project"
)

func TestExecuteAfterJSONExtractsCandidatesWithoutPersistingByDefault(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	sessionFile := filepath.Join(t.TempDir(), "session.md")
	if err := os.WriteFile(sessionFile, []byte("Decision: Repository layer stays framework-agnostic\nDo not import Cobra from repository packages."), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"after", "--root", projectRoot, "--from", sessionFile, "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Candidates []struct {
			ID    string `json:"id"`
			Kind  string `json:"kind"`
			Title string `json:"title"`
		} `json:"candidates"`
		Persisted []struct {
			ID string `json:"id"`
		} `json:"persisted"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Candidates) != 1 || payload.Candidates[0].ID != "cand_1" || payload.Candidates[0].Kind != "decision" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if len(payload.Persisted) != 0 {
		t.Fatalf("default after must not persist: %#v", payload.Persisted)
	}
	proj, err := project.Detect(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	records, err := memory.LoadRecords(derivePaths(storage, proj).MemoriesDir, proj.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("unexpected persisted records: %#v", records)
	}
}

func TestExecuteAfterAdapterJSONL(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	sessionFile := filepath.Join(t.TempDir(), "session.jsonl")
	if err := os.WriteFile(sessionFile, []byte(`{"message":{"content":"Decision: MCP stays stdio\\nUse JSON-RPC lines."}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"after", "--root", projectRoot, "--from", sessionFile, "--adapter", "claude-code", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Candidates []struct {
			Title string `json:"title"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Candidates) != 1 || payload.Candidates[0].Title != "MCP stays stdio" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestExecuteAfterApproveAllPersistsCandidates(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	sessionFile := filepath.Join(t.TempDir(), "session.md")
	if err := os.WriteFile(sessionFile, []byte("Constraint: JSON stdout stays machine-readable\nWarnings must go to stderr."), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"after", "--root", projectRoot, "--from", sessionFile, "--approve", "all", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Persisted []struct {
			Kind  string `json:"kind"`
			Title string `json:"title"`
		} `json:"persisted"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Persisted) != 1 || payload.Persisted[0].Kind != "constraint" {
		t.Fatalf("unexpected persisted payload: %#v", payload)
	}
	proj, err := project.Detect(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	records, err := memory.LoadRecords(derivePaths(storage, proj).MemoriesDir, proj.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Title != "JSON stdout stays machine-readable" {
		t.Fatalf("unexpected records: %#v", records)
	}
}

func TestExecuteAfterApproveAllSkipsDuplicates(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	if code := Execute([]string{"remember", "Do not import Cobra from repository packages.", "--root", projectRoot, "--kind", "decision", "--title", "Repository layer stays framework-agnostic", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); code != 0 {
		t.Fatalf("remember code=%d", code)
	}
	sessionFile := filepath.Join(t.TempDir(), "session.md")
	if err := os.WriteFile(sessionFile, []byte("Decision: Repository layer stays framework-agnostic\nDo not import Cobra from repository packages."), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"after", "--root", projectRoot, "--from", sessionFile, "--approve", "all", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Duplicates []struct {
			CandidateID string `json:"candidate_id"`
		} `json:"duplicates"`
		Persisted []struct {
			ID string `json:"id"`
		} `json:"persisted"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Duplicates) != 1 || len(payload.Persisted) != 0 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	proj, err := project.Detect(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	records, err := memory.LoadRecords(derivePaths(storage, proj).MemoriesDir, proj.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("duplicate should not create a second record: %#v", records)
	}
}
