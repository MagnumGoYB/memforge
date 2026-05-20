package memory

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func RenderMarkdownBlock(record Record) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("<!-- memforge:memory id=%s kind=%s -->", record.ID, record.Kind))
	lines = append(lines, "---")
	lines = append(lines, fmt.Sprintf("title: %s", quoteYAMLString(record.Title)))
	lines = append(lines, fmt.Sprintf("tags: [%s]", renderTags(record.Tags)))
	if record.Source != "" {
		lines = append(lines, fmt.Sprintf("source: %s", quoteYAMLString(record.Source)))
	}
	lines = append(lines, fmt.Sprintf("confidence: %s", strconv.FormatFloat(record.Confidence, 'f', -1, 64)))
	lines = append(lines, fmt.Sprintf("created_at: %s", record.CreatedAt.Format(time.RFC3339)))
	lines = append(lines, fmt.Sprintf("updated_at: %s", record.UpdatedAt.Format(time.RFC3339)))
	lines = append(lines, "---", "")
	lines = append(lines, record.Content, "", "<!-- /memforge:memory -->", "")
	return strings.Join(lines, "\n")
}

func ParseMarkdownBlocks(data string, kind Kind, projectID string) ([]Record, error) {
	blocks := strings.Split(data, "<!-- memforge:memory ")
	if len(blocks) == 1 {
		return nil, nil
	}
	records := make([]Record, 0, len(blocks)-1)
	for _, block := range blocks[1:] {
		record, err := parseMarkdownBlock(block, kind, projectID)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func parseMarkdownBlock(block string, kind Kind, projectID string) (Record, error) {
	markerEnd := strings.Index(block, "-->")
	if markerEnd < 0 {
		return Record{}, fmt.Errorf("invalid memory marker")
	}
	marker := strings.TrimSpace(block[:markerEnd])
	rest := strings.TrimLeft(block[markerEnd+3:], "\n")
	if !strings.HasPrefix(rest, "---\n") {
		return Record{}, fmt.Errorf("invalid memory frontmatter")
	}
	rest = strings.TrimPrefix(rest, "---\n")
	sections := strings.SplitN(rest, "\n---\n", 2)
	if len(sections) != 2 {
		return Record{}, fmt.Errorf("invalid memory frontmatter")
	}
	meta, err := parseFrontmatter(sections[0])
	if err != nil {
		return Record{}, err
	}
	bodyAndTail := sections[1]
	endMarker := "\n<!-- /memforge:memory -->"
	endIdx := strings.Index(bodyAndTail, endMarker)
	if endIdx < 0 {
		return Record{}, fmt.Errorf("missing memory end marker")
	}
	body := strings.TrimSpace(bodyAndTail[:endIdx])
	id, markerKind, err := parseMarker(marker)
	if err != nil {
		return Record{}, err
	}
	if markerKind != string(kind) {
		return Record{}, fmt.Errorf("marker kind %q does not match file kind %q", markerKind, kind)
	}
	createdAt, err := time.Parse(time.RFC3339, meta["created_at"])
	if err != nil {
		return Record{}, fmt.Errorf("parse created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, meta["updated_at"])
	if err != nil {
		return Record{}, fmt.Errorf("parse updated_at: %w", err)
	}
	confidence, err := strconv.ParseFloat(meta["confidence"], 64)
	if err != nil {
		return Record{}, fmt.Errorf("parse confidence: %w", err)
	}
	tags, err := parseTags(meta["tags"])
	if err != nil {
		return Record{}, err
	}
	record := Record{
		ID:         id,
		ProjectID:  projectID,
		Kind:       kind,
		Title:      unquoteYAMLString(meta["title"]),
		Content:    body,
		Tags:       tags,
		Source:     unquoteYAMLString(meta["source"]),
		Confidence: confidence,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
	return record, nil
}

func parseMarker(marker string) (id string, kind string, err error) {
	parts := strings.Fields(marker)
	for _, part := range parts {
		if strings.HasPrefix(part, "id=") {
			id = strings.TrimPrefix(part, "id=")
		}
		if strings.HasPrefix(part, "kind=") {
			kind = strings.TrimPrefix(part, "kind=")
		}
	}
	if id == "" || kind == "" {
		return "", "", fmt.Errorf("invalid memory marker fields")
	}
	return id, kind, nil
}

func parseFrontmatter(data string) (map[string]string, error) {
	meta := map[string]string{}
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid frontmatter line %q", line)
		}
		meta[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	for _, key := range []string{"title", "tags", "confidence", "created_at", "updated_at"} {
		if _, ok := meta[key]; !ok {
			return nil, fmt.Errorf("missing frontmatter field %q", key)
		}
	}
	return meta, nil
}

func parseTags(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "[]" {
		return nil, nil
	}
	if !strings.HasPrefix(raw, "[") || !strings.HasSuffix(raw, "]") {
		return nil, fmt.Errorf("invalid tags list %q", raw)
	}
	inner := strings.TrimSpace(raw[1 : len(raw)-1])
	if inner == "" {
		return nil, nil
	}
	parts := strings.Split(inner, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := unquoteYAMLString(strings.TrimSpace(part))
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	sort.Strings(tags)
	return tags, nil
}

func renderTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(tags))
	for _, tag := range tags {
		quoted = append(quoted, quoteYAMLString(tag))
	}
	return strings.Join(quoted, ", ")
}

func quoteYAMLString(value string) string {
	return strconv.Quote(value)
}

func unquoteYAMLString(value string) string {
	if value == "" {
		return ""
	}
	unquoted, err := strconv.Unquote(value)
	if err == nil {
		return unquoted
	}
	return value
}
