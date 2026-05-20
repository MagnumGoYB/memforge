package harness_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeCodePluginPackageIsInstallable(t *testing.T) {
	var plugin struct {
		Name       string `json:"name"`
		Version    string `json:"version"`
		MCPServers string `json:"mcpServers"`
	}
	mustReadJSON(t, filepath.Join("plugins", "claude-code", "memforge", ".claude-plugin", "plugin.json"), &plugin)
	if plugin.Name != "memforge" || plugin.Version == "" || plugin.MCPServers != "./.mcp.json" {
		t.Fatalf("unexpected Claude plugin manifest: %#v", plugin)
	}
	assertMCPLauncher(t, filepath.Join("plugins", "claude-code", "memforge", ".mcp.json"), "")

	var marketplace struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Owner       struct {
			Name string `json:"name"`
		} `json:"owner"`
		Plugins []struct {
			Name   string `json:"name"`
			Source string `json:"source"`
		} `json:"plugins"`
	}
	mustReadJSON(t, filepath.Join(".claude-plugin", "marketplace.json"), &marketplace)
	if marketplace.Name == "" || marketplace.Description == "" || marketplace.Owner.Name == "" {
		t.Fatalf("unexpected Claude marketplace metadata: %#v", marketplace)
	}
	if len(marketplace.Plugins) != 1 || marketplace.Plugins[0].Name != "memforge" || marketplace.Plugins[0].Source != "./plugins/claude-code/memforge" {
		t.Fatalf("unexpected Claude marketplace: %#v", marketplace)
	}
}

func TestCodexPluginPackageIsInstallable(t *testing.T) {
	var plugin struct {
		Name       string `json:"name"`
		Version    string `json:"version"`
		Skills     string `json:"skills"`
		MCPServers string `json:"mcpServers"`
	}
	mustReadJSON(t, filepath.Join("plugins", "codex", "memforge", ".codex-plugin", "plugin.json"), &plugin)
	if plugin.Name != "memforge" || plugin.Version == "" || plugin.Skills != "./skills/" || plugin.MCPServers != "./.mcp.json" {
		t.Fatalf("unexpected Codex plugin manifest: %#v", plugin)
	}
	assertMCPLauncher(t, filepath.Join("plugins", "codex", "memforge", ".mcp.json"), "approve")

	var marketplace struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Plugins     []struct {
			Name   string `json:"name"`
			Source struct {
				Source string `json:"source"`
				Path   string `json:"path"`
			} `json:"source"`
		} `json:"plugins"`
	}
	mustReadJSON(t, filepath.Join(".agents", "plugins", "marketplace.json"), &marketplace)
	if marketplace.Name == "" || marketplace.Description == "" {
		t.Fatalf("unexpected Codex marketplace metadata: %#v", marketplace)
	}
	if len(marketplace.Plugins) != 1 || marketplace.Plugins[0].Name != "memforge" || marketplace.Plugins[0].Source.Source != "local" || marketplace.Plugins[0].Source.Path != "./dist/plugins/codex/memforge" {
		t.Fatalf("unexpected Codex marketplace: %#v", marketplace)
	}
}

func TestPluginDistributionDocsAreBilingual(t *testing.T) {
	for _, path := range []string{
		filepath.Join("docs", "plugin-distribution.md"),
		filepath.Join("docs", "zh-CN", "plugin-distribution.md"),
		filepath.Join("plugins", "claude-code", "memforge", "README.md"),
		filepath.Join("plugins", "codex", "memforge", "README.md"),
	} {
		if read(t, path) == "" {
			t.Fatalf("empty plugin distribution doc: %s", path)
		}
	}
}

func TestAgentIntegrationDocsDefineSharedMemoryAcceptance(t *testing.T) {
	cases := map[string][]string{
		filepath.Join("docs", "integrations.md"): {
			"## Shared memory acceptance",
			"export MEMFORGE_HOME=\"/absolute/path/to/memforge-home\"",
			"memforge --no-version-check --root /absolute/path/to/project remember",
			"memforge --no-version-check --root /absolute/path/to/project search",
			"same absolute `MEMFORGE_HOME`",
			"same project root",
		},
		filepath.Join("docs", "zh-CN", "integrations.md"): {
			"## Shared memory acceptance",
			"export MEMFORGE_HOME=\"/absolute/path/to/memforge-home\"",
			"memforge --no-version-check --root /absolute/path/to/project remember",
			"memforge --no-version-check --root /absolute/path/to/project search",
			"同一个绝对路径 `MEMFORGE_HOME`",
			"同一个项目根目录",
		},
	}
	for path, expectedSnippets := range cases {
		doc := read(t, path)
		for _, expected := range expectedSnippets {
			if !strings.Contains(doc, expected) {
				t.Fatalf("%s must document shared memory acceptance with %q", path, expected)
			}
		}
	}
}

func TestPluginPackagingAutomationIsPresent(t *testing.T) {
	scriptPath := filepath.Join(repoRoot(t), "tools", "package_plugins.sh")
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("package plugin script must exist: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("package plugin script must be executable: mode %v", info.Mode())
	}

	script := read(t, "tools", "package_plugins.sh")
	for _, expected := range []string{
		"set -euo pipefail",
		"DIST=${DIST:-$ROOT/dist}",
		"VERSION=${VERSION:-$(cat \"$ROOT/VERSION\")}",
		"darwin-arm64",
		"darwin-amd64",
		"linux-arm64",
		"linux-amd64",
		"windows-arm64",
		"windows-amd64",
		"$DIST/plugins/memforge-claude-code-plugin-v$VERSION.zip",
		"$DIST/plugins/memforge-codex-plugin-v$VERSION.zip",
		"exit 2",
	} {
		if !strings.Contains(script, expected) {
			t.Fatalf("package plugin script must contain %q", expected)
		}
	}

	makefile := read(t, "Makefile")
	for _, expected := range []string{
		"package-plugins",
		"smoke-plugin-runtime",
		"./tools/package_plugins.sh",
		"memforge-mcp-launcher.js",
		"tools/list",
	} {
		if !strings.Contains(makefile, expected) {
			t.Fatalf("Makefile plugin automation must contain %q", expected)
		}
	}
}

func mustReadJSON(t *testing.T, path string, out any) {
	t.Helper()
	if err := json.Unmarshal([]byte(read(t, path)), out); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}

func assertMCPLauncher(t *testing.T, path string, wantApproval string) {
	t.Helper()
	var config struct {
		MCPServers map[string]struct {
			Command                  string            `json:"command"`
			Args                     []string          `json:"args"`
			Env                      map[string]string `json:"env"`
			DefaultToolsApprovalMode string            `json:"default_tools_approval_mode"`
		} `json:"mcpServers"`
	}
	mustReadJSON(t, path, &config)
	server, ok := config.MCPServers["memforge"]
	if !ok {
		t.Fatalf("%s missing memforge MCP server", path)
	}
	if server.Command != "node" || len(server.Args) != 1 || server.Args[0] != "${CLAUDE_PLUGIN_ROOT}/bin/memforge-mcp-launcher.js" {
		t.Fatalf("unexpected MCP launcher in %s: %#v", path, server)
	}
	if server.Env["MEMFORGE_PLUGIN_ROOT"] != "${CLAUDE_PLUGIN_ROOT}" {
		t.Fatalf("unexpected MCP launcher env in %s: %#v", path, server.Env)
	}
	if wantApproval != "" && server.DefaultToolsApprovalMode != wantApproval {
		t.Fatalf("unexpected MCP approval mode in %s: %q", path, server.DefaultToolsApprovalMode)
	}
	if read(t, filepath.Join(filepath.Dir(path), "bin", "memforge-mcp-launcher.js")) == "" {
		t.Fatalf("empty MCP launcher referenced by %s", path)
	}
}
