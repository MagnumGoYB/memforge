package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintCommitMessageAccepts(t *testing.T) {
	good := []string{
		"✨ feat(cli): add init command",
		"🐛 fix(index): handle WAL recovery on reopen",
		"📝 docs: extend memory format reference",
		"♻️ refactor(compiler): split ranker from budget allocator",
		"✅ test(memory): cover markdown frontmatter edge cases",
	}
	for _, header := range good {
		if problems := lintCommitMessage(header); len(problems) != 0 {
			t.Fatalf("expected %q to pass, got %v", header, problems)
		}
	}
}

func TestLintCommitMessageRejects(t *testing.T) {
	cases := map[string]string{
		"missing emoji":       "feat(cli): add init command",
		"wrong emoji":         "🐛 feat(cli): add init command",
		"unknown type":        "✨ bogus(cli): add init command",
		"unknown scope":       "✨ feat(unknown-scope): add init command",
		"empty subject":       "✨ feat(cli): ",
		"long header":         "✨ feat(cli): " + repeat("x", 200),
	}
	for name, header := range cases {
		if problems := lintCommitMessage(header); len(problems) == 0 {
			t.Fatalf("%s: expected failure", name)
		}
	}
}

func TestLintCommitMessageFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "msg")
	if err := os.WriteFile(path, []byte("# comment\n\n✨ feat(cli): add init command\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if problems := lintCommitMessage(string(data)); len(problems) != 0 {
		t.Fatalf("expected pass, got %v", problems)
	}
}

func repeat(s string, n int) string {
	out := ""
	for range n {
		out += s
	}
	return out
}
