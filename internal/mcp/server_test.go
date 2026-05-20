package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestServerListsAndCallsTools(t *testing.T) {
	server := Server{
		Tools: []Tool{{Name: "echo", Description: "Echo input", InputSchema: map[string]any{"type": "object"}}},
		Handlers: map[string]Handler{
			"echo": func(ctx context.Context, arguments json.RawMessage) (any, error) {
				return map[string]any{"ok": true}, nil
			},
		},
	}
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{}}}`,
	}, "\n") + "\n"
	var out bytes.Buffer
	if err := server.Serve(context.Background(), strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("unexpected output: %s", out.String())
	}
	if !strings.Contains(lines[0], "echo") || !strings.Contains(lines[1], `\"ok\":true`) {
		t.Fatalf("unexpected responses: %s", out.String())
	}
}
