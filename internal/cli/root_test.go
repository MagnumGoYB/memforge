package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/MagnumGOYB/memforge/internal/buildinfo"
)

func TestExecuteVersionJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"version", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["version"] == "" {
		t.Fatal("missing version")
	}
}

func TestExecuteRejectsInvalidFormat(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"version", "--format", "yaml"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 2 {
		t.Fatalf("got code %d stderr=%s", code, stderr.String())
	}
}

func TestExecuteVersionCheckWritesPromptToStderr(t *testing.T) {
	originalTag := buildinfo.Tag
	buildinfo.Tag = "0.1.0"
	t.Cleanup(func() { buildinfo.Tag = originalTag })
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	t.Setenv("MEMFORGE_VERSION_CHECK_LATEST", "99.0.0")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"version", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("new MemForge version available")) {
		t.Fatalf("missing version prompt on stderr: %s", stderr.String())
	}
	if !json.Valid(stdout.Bytes()) {
		t.Fatalf("stdout must remain JSON: %s", stdout.String())
	}
}

func TestExecuteNoVersionCheckSkipsVersionCheck(t *testing.T) {
	originalTag := buildinfo.Tag
	buildinfo.Tag = "0.1.0"
	t.Cleanup(func() { buildinfo.Tag = originalTag })
	storage := filepath.Join(t.TempDir(), "storage")
	t.Setenv("MEMFORGE_HOME", storage)
	t.Setenv("MEMFORGE_VERSION_CHECK_LATEST", "99.0.0")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute([]string{"version", "--no-version-check", "--format", "json"}, Streams{Stdin: bytes.NewBuffer(nil), Stdout: &stdout, Stderr: &stderr})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(storage, "cache", "version-check.json")); !os.IsNotExist(err) {
		t.Fatalf("version check cache should not be written when skipped: %v", err)
	}
}
