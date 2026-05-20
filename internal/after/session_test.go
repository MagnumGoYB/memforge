package after

import (
	"strings"
	"testing"
)

func TestExtractSessionTextPlain(t *testing.T) {
	text, err := ExtractSessionText("plain", []byte("Decision: Keep local\nBody"))
	if err != nil {
		t.Fatal(err)
	}
	if text != "Decision: Keep local\nBody" {
		t.Fatalf("got %q", text)
	}
}

func TestExtractSessionTextJSONL(t *testing.T) {
	text, err := ExtractSessionText("claude-code", []byte(`{"message":{"content":"Decision: Keep MCP stdio"}}`+"\n"+`{"type":"message","role":"assistant","content":[{"text":"Use JSON-RPC lines."}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Decision: Keep MCP stdio") || !strings.Contains(text, "Use JSON-RPC lines") {
		t.Fatalf("got %q", text)
	}
}

func TestExtractSessionTextJSONLSkipsToolArguments(t *testing.T) {
	text, err := ExtractSessionText("codex", []byte(`{"type":"tool_result","content":"Decision: should not be extracted"}`+"\n"+`{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Decision: Keep MCP stdio"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text, "should not be extracted") {
		t.Fatalf("tool arguments leaked into session text: %q", text)
	}
	if !strings.Contains(text, "Keep MCP stdio") {
		t.Fatalf("assistant text missing: %q", text)
	}
}

func TestExtractSessionTextJSONLSeparatesMessagesIntoBlocks(t *testing.T) {
	text, err := ExtractSessionText("codex", []byte(strings.Join([]string{
		`{"role":"assistant","content":"decision: Keep Codex plugin runtime packaged\nThe plugin bundle must include bin/darwin-arm64/memforge."}`,
		`{"role":"assistant","content":"constraint: Do not rely on PATH memforge for plugin smoke\nThe launcher must work after uninstalling the PATH binary."}`,
		`{"role":"assistant","content":"bugfix: Codex marketplace source package mismatch\nSource marketplace installs can miss platform runtime binaries."}`,
	}, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	candidates := ExtractCandidatesFromText(text)
	if len(candidates) != 3 {
		t.Fatalf("got %d candidates from text %q", len(candidates), text)
	}
	for i, want := range []string{"decision", "constraint", "bugfix"} {
		if string(candidates[i].Kind) != want {
			t.Fatalf("candidate %d kind = %s, want %s", i, candidates[i].Kind, want)
		}
	}
}
