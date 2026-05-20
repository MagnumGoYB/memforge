package harness_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from the test working directory until it finds go.mod,
// so harness tests work regardless of where `go test` is invoked from.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root")
		}
		dir = parent
	}
}

func read(t *testing.T, parts ...string) string {
	t.Helper()
	path := filepath.Join(append([]string{repoRoot(t)}, parts...)...)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", filepath.Join(parts...), err)
	}
	return string(data)
}

func TestHarnessDocsAndCommandsStayAligned(t *testing.T) {
	files := map[string]string{
		"agents":    read(t, "AGENTS.md"),
		"agentsZH":  read(t, "AGENTS.zh-CN.md"),
		"harness":   read(t, "docs", "harness-engineering.md"),
		"harnessZH": read(t, "docs", "zh-CN", "harness-engineering.md"),
		"readme":    read(t, "README.md"),
	}
	for _, command := range []string{
		"make setup",
		"make check",
		"make test",
		"make test-packages",
		"make test-harness",
		"make vet",
		"make build",
		"make validate",
		"make validate-pr-body",
	} {
		for name, content := range files {
			if !strings.Contains(content, command) {
				t.Fatalf("%s must mention %s", name, command)
			}
		}
	}
}

func TestReleaseWorkflowBuildsPackagesSmokesAndPublishesAssets(t *testing.T) {
	releaseWorkflow := read(t, ".github", "workflows", "release.yml")

	for _, expected := range []string{
		"push:",
		"tags:",
		"v*",
		"workflow_dispatch:",
		"contents: write",
		"darwin-arm64",
		"darwin-amd64",
		"linux-arm64",
		"linux-amd64",
		"windows-arm64",
		"windows-amd64",
		"GOOS: ${{ matrix.goos }}",
		"GOARCH: ${{ matrix.goarch }}",
		"GOCACHE:",
		"GOMODCACHE:",
		"actions/upload-artifact",
		"actions/download-artifact",
		"path: dist/bin",
		"make validate",
		"make test-harness",
		"./tools/package_plugins.sh",
		"memforge-claude-code-plugin-v*.zip",
		"memforge-codex-plugin-v*.zip",
		"MEMFORGE_PLUGIN_ROOT",
		"memforge-mcp-launcher.js",
		"tools/list",
		"softprops/action-gh-release@v2",
	} {
		if !strings.Contains(releaseWorkflow, expected) {
			t.Fatalf("release workflow must contain %s", expected)
		}
	}
}

func TestCommitWorkflowSupportsHumanAndHostedValidation(t *testing.T) {
	makefile := read(t, "Makefile")
	for _, expected := range []string{
		".PHONY: setup",
		"setup:",
		"git config core.hooksPath .githooks",
		"commitlint:",
		"COMMIT_MSG_FILE",
	} {
		if !strings.Contains(makefile, expected) {
			t.Fatalf("Makefile commit workflow must contain %s", expected)
		}
	}

	hook := read(t, ".githooks", "commit-msg")
	for _, expected := range []string{
		"make commitlint",
		`COMMIT_MSG_FILE="$1"`,
	} {
		if !strings.Contains(hook, expected) {
			t.Fatalf("commit-msg hook must contain %s", expected)
		}
	}

	prWorkflow := read(t, ".github", "workflows", "pr.yml")
	for _, expected := range []string{
		"fetch-depth: 0",
		"github.event.pull_request.base.sha",
		"github.event.pull_request.head.sha",
		"make commitlint-range COMMIT_RANGE=",
	} {
		if !strings.Contains(prWorkflow, expected) {
			t.Fatalf("PR workflow must validate every PR commit message with %s", expected)
		}
	}
}

func TestModulePathMatchesGitHubRepository(t *testing.T) {
	combined := strings.Join([]string{
		read(t, "go.mod"),
		read(t, "README.md"),
	}, "\n")
	if !strings.Contains(combined, "github.com/MagnumGOYB/memforge") {
		t.Fatal("module path github.com/MagnumGOYB/memforge must be referenced in go.mod and README")
	}
}

func TestOfflineGuardrailsArePresent(t *testing.T) {
	agents := read(t, "AGENTS.md")
	agentsZH := read(t, "AGENTS.zh-CN.md")
	for _, snippet := range []string{
		"MEMFORGE_NO_VERSION_CHECK",
		"--no-version-check",
		"MEMFORGE_HOME",
		"XDG_DATA_HOME",
		"opt-in",
	} {
		if !strings.Contains(agents, snippet) {
			t.Fatalf("AGENTS.md must mention %s", snippet)
		}
	}
	for _, snippet := range []string{
		"MEMFORGE_NO_VERSION_CHECK",
		"--no-version-check",
		"MEMFORGE_HOME",
		"XDG_DATA_HOME",
		"opt-in",
	} {
		if !strings.Contains(agentsZH, snippet) {
			t.Fatalf("AGENTS.zh-CN.md must mention %s", snippet)
		}
	}
}

func TestPRTemplateMentionsRequiredSections(t *testing.T) {
	template := read(t, ".github", "pull_request_template.md")
	for _, section := range []string{
		"## Summary",
		"## Requirement Classification",
		"## Acceptance Criteria",
		"## Changed Areas",
		"## Release Decision",
		"## TDD / Test Evidence",
		"## Validation",
		"## Risk and Rollback",
	} {
		if !strings.Contains(template, section) {
			t.Fatalf("PR template missing section: %s", section)
		}
	}
}

func TestBilingualMirrorsExist(t *testing.T) {
	root := repoRoot(t)
	pairs := [][2]string{
		{"AGENTS.md", "AGENTS.zh-CN.md"},
		{"CLAUDE.md", "CLAUDE.zh-CN.md"},
		{"README.md", "README.zh-CN.md"},
		{"CONTRIBUTING.md", "CONTRIBUTING.zh-CN.md"},
		{"plan.md", "plan.zh-CN.md"},
		{filepath.Join("docs", "harness-engineering.md"), filepath.Join("docs", "zh-CN", "harness-engineering.md")},
		{filepath.Join("docs", "github-automation.md"), filepath.Join("docs", "zh-CN", "github-automation.md")},
		{filepath.Join(".github", "pull_request_template.md"), filepath.Join(".github", "pull_request_template.zh-CN.md")},
	}
	for _, pair := range pairs {
		for _, path := range pair {
			if _, err := os.Stat(filepath.Join(root, path)); errors.Is(err, os.ErrNotExist) {
				t.Fatalf("missing bilingual mirror file: %s", path)
			}
		}
	}
}

func TestProductionPackagesDoNotImportHarnessOrTools(t *testing.T) {
	root := repoRoot(t)
	internalDir := filepath.Join(root, "internal")
	cmdDir := filepath.Join(root, "cmd")
	walk := func(base string) {
		_ = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			content := string(data)
			for _, banned := range []string{
				`"github.com/MagnumGOYB/memforge/harness`,
				`"github.com/MagnumGOYB/memforge/tools`,
			} {
				if strings.Contains(content, banned) {
					t.Fatalf("%s must not import %s", path, banned)
				}
			}
			return nil
		})
	}
	walk(internalDir)
	walk(cmdDir)
}
