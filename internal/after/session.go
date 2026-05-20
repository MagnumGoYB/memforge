package after

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ExtractSessionText(adapter string, data []byte) (string, error) {
	adapter = strings.TrimSpace(strings.ToLower(adapter))
	if adapter == "" || adapter == "plain" {
		return string(data), nil
	}
	switch adapter {
	case "jsonl":
		return extractJSONLText(data, false), nil
	case "claude-code", "codex", "cursor":
		return extractJSONLText(data, true), nil
	default:
		return "", fmt.Errorf("unknown session adapter %q", adapter)
	}
}

func extractJSONLText(data []byte, messagesOnly bool) string {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var value any
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			parts = append(parts, line)
			continue
		}
		before := len(parts)
		collectJSONLText(value, &parts, messagesOnly)
		if len(parts) > before {
			parts = append(parts, "")
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func collectJSONLText(value any, parts *[]string, messagesOnly bool) {
	if messagesOnly {
		if message, ok := messagePayload(value); ok {
			collectText(message, parts)
		}
		return
	}
	collectText(value, parts)
}

func messagePayload(value any) (any, bool) {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}
	if message, ok := obj["message"]; ok {
		return message, true
	}
	if role, ok := obj["role"].(string); ok && (role == "assistant" || role == "user") {
		return obj["content"], true
	}
	if typ, ok := obj["type"].(string); ok && typ == "message" {
		return obj["content"], true
	}
	return nil, false
}

func collectText(value any, parts *[]string) {
	switch typed := value.(type) {
	case string:
		typed = strings.ReplaceAll(typed, "\\n", "\n")
		if strings.TrimSpace(typed) != "" {
			*parts = append(*parts, typed)
		}
	case []any:
		for _, item := range typed {
			collectText(item, parts)
		}
	case map[string]any:
		for _, key := range []string{"content", "text", "message", "output", "input"} {
			if item, ok := typed[key]; ok {
				collectText(item, parts)
			}
		}
	}
}
