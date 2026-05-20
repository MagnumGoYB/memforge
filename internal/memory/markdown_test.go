package memory

import (
	"strings"
	"testing"
	"time"
)

func TestParseMarkdownBlocksRoundTrip(t *testing.T) {
	record, err := NewRecord(NewRecordInput{
		ProjectID:  "project-1",
		Kind:       KindDecision,
		Title:      "CLI framework",
		Content:    "Body in markdown.",
		Tags:       []string{"cli", "architecture"},
		Source:     "cli",
		Confidence: 1,
		Now:        time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	data := RenderMarkdownBlock(record)
	records, err := ParseMarkdownBlocks(data, KindDecision, "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records", len(records))
	}
	got := records[0]
	if got.ID != record.ID || got.Title != record.Title || got.Content != record.Content {
		t.Fatalf("unexpected record: %#v", got)
	}
	if strings.Join(got.Tags, ",") != "architecture,cli" {
		t.Fatalf("got tags %v", got.Tags)
	}
}

func TestParseMarkdownBlocksRejectsInvalidBlock(t *testing.T) {
	_, err := ParseMarkdownBlocks("<!-- memforge:memory id=x kind=decision -->\n---\ntitle: \"x\"\n---\nBody\n<!-- /memforge:memory -->\n", KindDecision, "project-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseMarkdownBlocksPreservesIndentedBodyEdges(t *testing.T) {
	data := "<!-- memforge:memory id=mem_1 kind=decision -->\n---\ntitle: \"Indented markdown\"\ntags: []\nconfidence: 1\ncreated_at: 2026-05-19T10:00:00Z\nupdated_at: 2026-05-19T10:00:00Z\n---\n    code block\n\nbody\n    tail\n<!-- /memforge:memory -->\n"
	records, err := ParseMarkdownBlocks(data, KindDecision, "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records", len(records))
	}
	want := "    code block\n\nbody\n    tail"
	if records[0].Content != want {
		t.Fatalf("content changed:\nwant %q\n got %q", want, records[0].Content)
	}
}
