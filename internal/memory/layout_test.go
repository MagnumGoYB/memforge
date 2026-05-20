package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureLayoutCreatesAllFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "memories")
	if err := EnsureLayout(dir); err != nil {
		t.Fatal(err)
	}
	for _, name := range DefaultFileNames() {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestEnsureLayoutPreservesExistingContent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "memories")
	if err := EnsureLayout(dir); err != nil {
		t.Fatal(err)
	}
	manual := filepath.Join(dir, "manual.md")
	if err := os.WriteFile(manual, []byte("keep me"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureLayout(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(manual)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "keep me" {
		t.Fatalf("got %q want keep me", string(data))
	}
}
