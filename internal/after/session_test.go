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
	text, err := ExtractSessionText("claude-code", []byte(`{"message":{"content":"Decision: Keep MCP stdio"}}`+"\n"+`{"content":[{"text":"Use JSON-RPC lines."}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Decision: Keep MCP stdio") || !strings.Contains(text, "Use JSON-RPC lines") {
		t.Fatalf("got %q", text)
	}
}
