package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveStorageRootUsesMemforgeHome(t *testing.T) {
	t.Setenv("MEMFORGE_HOME", filepath.Join(string(filepath.Separator), "tmp", "memforge-home"))
	t.Setenv("XDG_DATA_HOME", "")
	root, err := ResolveStorageRoot()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := root, filepath.Join(string(filepath.Separator), "tmp", "memforge-home"); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveStorageRootRejectsRelativeMemforgeHome(t *testing.T) {
	t.Setenv("MEMFORGE_HOME", "relative/path")
	_, err := ResolveStorageRoot()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveStorageRootUsesXDGDataHome(t *testing.T) {
	t.Setenv("MEMFORGE_HOME", "")
	t.Setenv("XDG_DATA_HOME", filepath.Join(string(filepath.Separator), "tmp", "xdg-home"))
	root, err := ResolveStorageRoot()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(string(filepath.Separator), "tmp", "xdg-home", "memforge")
	if root != want {
		t.Fatalf("got %q want %q", root, want)
	}
}

func TestResolveStorageRootUsesHomeFallback(t *testing.T) {
	origMemforgeHome := os.Getenv("MEMFORGE_HOME")
	origXDGDataHome := os.Getenv("XDG_DATA_HOME")
	origHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("MEMFORGE_HOME", origMemforgeHome)
		_ = os.Setenv("XDG_DATA_HOME", origXDGDataHome)
		_ = os.Setenv("HOME", origHome)
	}()
	_ = os.Unsetenv("MEMFORGE_HOME")
	_ = os.Unsetenv("XDG_DATA_HOME")
	tmp := t.TempDir()
	_ = os.Setenv("HOME", tmp)
	root, err := ResolveStorageRoot()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(tmp, ".local", "share", "memforge")
	if root != want {
		t.Fatalf("got %q want %q", root, want)
	}
}
