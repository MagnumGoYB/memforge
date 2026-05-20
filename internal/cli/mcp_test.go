package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteMCPListsTools(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"mcp", "--root", projectRoot}, Streams{Stdin: strings.NewReader(input), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "search_memory") || !strings.Contains(stdout.String(), "get_project_context") {
		t.Fatalf("unexpected response: %s", stdout.String())
	}
}

func TestExecuteMCPSaveAndSearchMemory(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"save_memory","arguments":{"kind":"decision","title":"Repository decision","content":"repository framework","tags":["repository"]}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search_memory","arguments":{"query":"repository","limit":5}}}`,
	}, "\n") + "\n"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"mcp", "--root", projectRoot}, Streams{Stdin: strings.NewReader(input), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("unexpected responses: %s", stdout.String())
	}
	var searchResp struct {
		Result struct {
			Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &searchResp); err != nil {
		t.Fatal(err)
	}
	if len(searchResp.Result.Content) != 1 || !strings.Contains(searchResp.Result.Content[0].Text, "Repository decision") {
		t.Fatalf("unexpected search response: %s", lines[1])
	}
}
