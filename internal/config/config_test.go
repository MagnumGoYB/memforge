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

func TestLoadProjectSettingsReadsMemoryRCAndUserConfig(t *testing.T) {
	projectRoot := t.TempDir()
	userConfigDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", userConfigDir)
	userConfigPath := filepath.Join(userConfigDir, "memforge", "config.toml")
	if err := os.MkdirAll(filepath.Dir(userConfigPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(userConfigPath, []byte("default_budget = 900\n[kind_weights]\ndecision = 5\nbugfix = 95\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, ".memoryrc"), []byte("default_budget = 1200\n[kind_weights]\ndecision = 99\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	settings, err := LoadProjectSettings(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if settings.DefaultBudget != 1200 {
		t.Fatalf("default budget=%d, want 1200", settings.DefaultBudget)
	}
	if settings.KindWeights["decision"] != 99 {
		t.Fatalf("project kind weight did not override user config: %#v", settings.KindWeights)
	}
	if settings.KindWeights["bugfix"] != 95 {
		t.Fatalf("user config kind weight should be retained: %#v", settings.KindWeights)
	}
}
