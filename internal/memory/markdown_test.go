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
