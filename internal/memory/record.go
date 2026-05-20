package memory

import (
	"crypto/rand"
	"fmt"
	"sort"
	"strings"
	"time"

	ulid "github.com/oklog/ulid/v2"
)

type NewRecordInput struct {
	ProjectID  string
	Kind       Kind
	Title      string
	Content    string
	Tags       []string
	Source     string
	Confidence float64
	Now        time.Time
}

type Record struct {
	ID         string
	ProjectID  string
	Kind       Kind
	Title      string
	Content    string
	Tags       []string
	Source     string
	Confidence float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewRecord(input NewRecordInput) (Record, error) {
	if input.ProjectID == "" {
		return Record{}, fmt.Errorf("project id is required")
	}
	if _, ok := FileNameForKind(input.Kind); !ok {
		return Record{}, fmt.Errorf("invalid kind %q", input.Kind)
	}
	title := strings.Join(strings.Fields(strings.TrimSpace(input.Title)), " ")
	if title == "" {
		return Record{}, fmt.Errorf("title is required")
	}
	content := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(input.Content, "\r\n", "\n"), "\r", "\n"))
	if content == "" {
		return Record{}, fmt.Errorf("content is required")
	}
	if input.Confidence < 0 || input.Confidence > 1 {
		return Record{}, fmt.Errorf("confidence must be between 0 and 1")
	}
	now := input.Now.UTC()
	if input.Now.IsZero() {
		now = time.Now().UTC()
	}
	tags := normalizeTags(input.Tags)
	return Record{
		ID:         newID(),
		ProjectID:  input.ProjectID,
		Kind:       input.Kind,
		Title:      title,
		Content:    content,
		Tags:       tags,
		Source:     strings.TrimSpace(input.Source),
		Confidence: input.Confidence,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func normalizeTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

func newID() string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(time.Now().UTC()), entropy).String()
}
