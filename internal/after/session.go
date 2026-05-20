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
	case "jsonl", "claude-code", "codex", "cursor":
		return extractJSONLText(data), nil
	default:
		return "", fmt.Errorf("unknown session adapter %q", adapter)
	}
}

func extractJSONLText(data []byte) string {
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
		collectText(value, &parts)
	}
	return strings.Join(parts, "\n")
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
