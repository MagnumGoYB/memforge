package memory

import (
	"strings"
	"testing"
	"time"
)

func TestNewRecordNormalizesInput(t *testing.T) {
	record, err := NewRecord(NewRecordInput{
		ProjectID:  "project-1",
		Kind:       KindDecision,
		Title:      "  CLI   framework  ",
		Content:    "\nBody\n",
		Tags:       []string{" cli ", "architecture", "cli"},
		Source:     " planning ",
		Confidence: 0.9,
		Now:        time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if record.Title != "CLI framework" {
		t.Fatalf("got %q", record.Title)
	}
	if record.Content != "Body" {
		t.Fatalf("got %q", record.Content)
	}
	if len(record.Tags) != 2 {
		t.Fatalf("got tags %v", record.Tags)
	}
}

func TestNewRecordRejectsInvalidInput(t *testing.T) {
	cases := []NewRecordInput{
		{ProjectID: "p", Kind: KindDecision, Title: "", Content: "Body", Confidence: 1},
		{ProjectID: "p", Kind: KindDecision, Title: "Title", Content: "", Confidence: 1},
		{ProjectID: "p", Kind: KindDecision, Title: "Title", Content: "Body", Confidence: 2},
	}
	for _, input := range cases {
		if _, err := NewRecord(input); err == nil {
			t.Fatalf("expected error for %+v", input)
		}
	}
}

func TestRenderMarkdownBlock(t *testing.T) {
	record, err := NewRecord(NewRecordInput{
		ProjectID:  "project-1",
		Kind:       KindDecision,
		Title:      "CLI framework",
		Content:    "Body",
		Tags:       []string{"cli", "architecture"},
		Source:     "planning",
		Confidence: 1,
		Now:        time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	block := RenderMarkdownBlock(record)
	for _, snippet := range []string{"<!-- memforge:memory id=", "title: \"CLI framework\"", "tags: [\"architecture\", \"cli\"]", "confidence: 1", "Body", "<!-- /memforge:memory -->"} {
		if !strings.Contains(block, snippet) {
			t.Fatalf("missing %q in block:\n%s", snippet, block)
		}
	}
}
