package cli

import (
	"bytes"
	"encoding/json"
	"os"
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

func TestExecuteMCPUpsertProjectMemoryUpdatesExistingTitle(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"upsert_project_memory","arguments":{"kind":"decision","title":"Plugin memory policy","content":"Agents may save stable memories from active threads.","tags":["plugin"]}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"upsert_project_memory","arguments":{"kind":"decision","title":"Plugin memory policy","content":"Agents may save or update stable memories from active Claude Code and Codex threads.","tags":["plugin","automation"]}}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_memory","arguments":{"query":"Claude Code Codex threads","limit":5}}}`,
	}, "\n") + "\n"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"mcp", "--root", projectRoot}, Streams{Stdin: strings.NewReader(input), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("unexpected responses: %s", stdout.String())
	}
	first := decodeMCPTextPayload(t, lines[0])
	second := decodeMCPTextPayload(t, lines[1])
	if first["action"] != "created" || second["action"] != "updated" || first["id"] != second["id"] {
		t.Fatalf("unexpected upsert actions: first=%v second=%v", first, second)
	}
	search := decodeMCPTextPayload(t, lines[2])
	if got := int(search["count"].(float64)); got != 1 {
		t.Fatalf("search count=%d, want 1: %v", got, search)
	}
	matches, err := filepath.Glob(filepath.Join(storage, "projects", "*", "memories", "decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one decisions file, got %v", matches)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), "<!-- memforge:memory ") != 1 || !strings.Contains(string(data), "Claude Code and Codex threads") {
		t.Fatalf("memory file should contain one updated block: %s", string(data))
	}
}

func TestExecuteMCPSearchFindsPartialMatchesForBroadQueries(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"upsert_project_memory","arguments":{"kind":"agent-instruction","title":"Plugin display name uses interface metadata","content":"Use interface displayName MemForge for UI display casing.","tags":["plugin","display-name"]}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"upsert_project_memory","arguments":{"kind":"agent-instruction","title":"Release acceptance checks plugin assets and curl latest","content":"Validate release assets and curl latest install.","tags":["release","curl"]}}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_memory","arguments":{"query":"MemForge plugin displayName bundled runtimes release acceptance curl latest shared MEMFORGE_HOME","limit":10,"hybrid":true}}}`,
	}, "\n") + "\n"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"mcp", "--root", projectRoot}, Streams{Stdin: strings.NewReader(input), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("unexpected responses: %s", stdout.String())
	}
	search := decodeMCPTextPayload(t, lines[2])
	if got := int(search["count"].(float64)); got < 2 {
		t.Fatalf("search count=%d, want at least 2: %v", got, search)
	}
	text := search["results"].([]any)
	joined, err := json.Marshal(text)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(joined), "Plugin display name uses interface metadata") || !strings.Contains(string(joined), "Release acceptance checks plugin assets and curl latest") {
		t.Fatalf("search missed expected memories: %s", string(joined))
	}
}

func TestExecuteMCPUpsertProjectMemoryPreservesOtherKindRecords(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"save_memory","arguments":{"kind":"decision","title":"Other decision","content":"Keep this decision.","tags":["keep"]}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"upsert_project_memory","arguments":{"kind":"decision","title":"Updated decision","content":"Write this decision.","tags":["update"]}}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_memory","arguments":{"query":"decision","limit":10}}}`,
	}, "\n") + "\n"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"mcp", "--root", projectRoot}, Streams{Stdin: strings.NewReader(input), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("unexpected responses: %s", stdout.String())
	}
	search := decodeMCPTextPayload(t, lines[2])
	if got := int(search["count"].(float64)); got != 2 {
		t.Fatalf("search count=%d, want 2: %v", got, search)
	}
	matches, err := filepath.Glob(filepath.Join(storage, "projects", "*", "memories", "decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one decisions file, got %v", matches)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "Other decision") || !strings.Contains(text, "Updated decision") {
		t.Fatalf("upsert should preserve existing kind records: %s", text)
	}
}

func TestExecuteMCPCompileUsesProjectDefaultBudget(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	projectRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectRoot, ".memoryrc"), []byte("default_budget = 1400\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"save_memory","arguments":{"kind":"manual","title":"Manual note","content":"manual note body"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"compile_context","arguments":{}}}`,
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
	payload := decodeMCPTextPayload(t, lines[1])
	if got := int(payload["budget"].(float64)); got != 1400 {
		t.Fatalf("budget=%d, want 1400: %v", got, payload)
	}
}

func decodeMCPTextPayload(t *testing.T, line string) map[string]any {
	t.Helper()
	var resp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Result.Content) != 1 {
		t.Fatalf("unexpected MCP response: %s", line)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(resp.Result.Content[0].Text), &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}
