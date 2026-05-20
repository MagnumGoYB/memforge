package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
)

func TestExecuteSearchHybridJSON(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	if code := Execute([]string{"remember", "repository framework", "--root", projectRoot, "--kind", "decision", "--title", "Repository architecture", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); code != 0 {
		t.Fatalf("remember code=%d", code)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"search", "repository", "--root", projectRoot, "--hybrid", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Hybrid bool `json:"hybrid"`
		Count  int  `json:"count"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Hybrid || payload.Count != 1 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestExecuteSearchJSON(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	if code := Execute([]string{"remember", "Repository layer must remain framework-agnostic", "--root", projectRoot, "--kind", "decision", "--title", "Repository decision", "--tag", "architecture", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); code != 0 {
		t.Fatalf("remember code=%d", code)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"search", "repository framework", "--root", projectRoot, "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Count   int `json:"count"`
		Results []struct {
			Kind    string  `json:"kind"`
			Title   string  `json:"title"`
			Snippet string  `json:"snippet"`
			Score   float64 `json:"score"`
		} `json:"results"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Count != 1 || len(payload.Results) != 1 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.Results[0].Kind != "decision" || payload.Results[0].Title == "" || payload.Results[0].Snippet == "" || payload.Results[0].Score <= 0 {
		t.Fatalf("unexpected result: %#v", payload.Results[0])
	}
}

func TestExecuteSearchFiltersTags(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	for _, args := range [][]string{
		{"remember", "repository framework", "--root", projectRoot, "--kind", "decision", "--title", "Decision", "--tag", "architecture", "--format", "json"},
		{"remember", "repository framework", "--root", projectRoot, "--kind", "constraint", "--title", "Constraint", "--tag", "policy", "--format", "json"},
	} {
		if code := Execute(args, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); code != 0 {
			t.Fatalf("remember code=%d args=%v", code, args)
		}
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"search", "repository framework", "--root", projectRoot, "--tag", "policy", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Count   int `json:"count"`
		Results []struct {
			Kind string `json:"kind"`
		} `json:"results"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Count != 1 || payload.Results[0].Kind != "constraint" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestExecuteSearchTagFilterReturnsMatchesBeyondInitialLimit(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	for i := 0; i < 25; i++ {
		args := []string{"remember", "repository framework", "--root", projectRoot, "--kind", "decision", "--title", fmt.Sprintf("General %02d", i), "--tag", "general", "--format", "json"}
		if code := Execute(args, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); code != 0 {
			t.Fatalf("remember code=%d args=%v", code, args)
		}
	}
	if code := Execute([]string{"remember", "repository framework", "--root", projectRoot, "--kind", "constraint", "--title", "Policy match", "--tag", "policy", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); code != 0 {
		t.Fatalf("remember policy code=%d", code)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"search", "repository framework", "--root", projectRoot, "--tag", "policy", "--limit", "1", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload struct {
		Count   int `json:"count"`
		Results []struct {
			Title string `json:"title"`
		} `json:"results"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Count != 1 || len(payload.Results) != 1 || payload.Results[0].Title != "Policy match" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
