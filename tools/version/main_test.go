package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadVersionAcceptsSemver(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(path, []byte("0.1.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := readVersion(path)
	if err != nil {
		t.Fatal(err)
	}
	if v != "0.1.0" {
		t.Fatalf("got %q want 0.1.0", v)
	}
}

func TestReadVersionRejectsNonSemver(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(path, []byte("v0.1.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readVersion(path); err == nil {
		t.Fatal("expected leading v to fail")
	}
}

func TestCheckRefName(t *testing.T) {
	if err := checkRefName("0.1.0", "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	if err := checkRefName("0.1.0", "v0.2.0"); err == nil {
		t.Fatal("expected mismatch")
	}
	if err := checkRefName("0.1.0", ""); err == nil {
		t.Fatal("expected empty ref to fail")
	}
}
