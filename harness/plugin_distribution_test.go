package harness_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeCodePluginPackageIsInstallable(t *testing.T) {
	version := strings.TrimSpace(read(t, "VERSION"))
	var plugin struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Version     string `json:"version"`
		MCPServers  string `json:"mcpServers"`
		Interface   struct {
			DisplayName string `json:"displayName"`
		} `json:"interface"`
	}
	mustReadJSON(t, filepath.Join("plugins", "claude-code", "memforge", ".claude-plugin", "plugin.json"), &plugin)
	if plugin.Name != "memforge" || plugin.Version != version || plugin.MCPServers != "./.mcp.json" || plugin.Interface.DisplayName != "MemForge" {
		t.Fatalf("unexpected Claude plugin manifest: %#v", plugin)
	}
	assertClaudeMCPLauncher(t, filepath.Join("plugins", "claude-code", "memforge", ".mcp.json"))
	if !strings.Contains(plugin.Description, "MemForge") {
		t.Fatalf("Claude plugin description must use MemForge product casing: %q", plugin.Description)
	}

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
	if len(marketplace.Plugins) != 1 || marketplace.Plugins[0].Name != "memforge" || marketplace.Plugins[0].Source != "./dist/plugins/claude-code/memforge" {
		t.Fatalf("unexpected Claude marketplace: %#v", marketplace)
	}
}

func TestCodexPluginPackageIsInstallable(t *testing.T) {
	version := strings.TrimSpace(read(t, "VERSION"))
	var plugin struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Version     string `json:"version"`
		Skills      string `json:"skills"`
		MCPServers  string `json:"mcpServers"`
		Interface   struct {
			DisplayName string `json:"displayName"`
		} `json:"interface"`
	}
	mustReadJSON(t, filepath.Join("plugins", "codex", "memforge", ".codex-plugin", "plugin.json"), &plugin)
	if plugin.Name != "memforge" || plugin.Version != version || plugin.Skills != "./skills/" || plugin.MCPServers != "./.mcp.json" || plugin.Interface.DisplayName != "MemForge" {
		t.Fatalf("unexpected Codex plugin manifest: %#v", plugin)
	}
	assertCodexMCPLauncher(t, filepath.Join("plugins", "codex", "memforge", ".mcp.json"))
	if !strings.Contains(plugin.Description, "MemForge") {
		t.Fatalf("Codex plugin description must use MemForge product casing: %q", plugin.Description)
	}

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

	var altMarketplace struct {
		Plugins []struct {
			Name   string `json:"name"`
			Source struct {
				Source string `json:"source"`
				Path   string `json:"path"`
			} `json:"source"`
		} `json:"plugins"`
	}
	mustReadJSON(t, filepath.Join("marketplaces", "codex", "marketplace.json"), &altMarketplace)
	if len(altMarketplace.Plugins) != 1 || altMarketplace.Plugins[0].Name != "memforge" || altMarketplace.Plugins[0].Source.Source != "local" || altMarketplace.Plugins[0].Source.Path != "./dist/plugins/codex/memforge" {
		t.Fatalf("unexpected alternate Codex marketplace: %#v", altMarketplace)
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

func TestPluginSkillsAllowAutomaticMemoryMaintenance(t *testing.T) {
	for _, path := range []string{
		filepath.Join("plugins", "claude-code", "memforge", "skills", "memforge-memory", "SKILL.md"),
		filepath.Join("plugins", "codex", "memforge", "skills", "memforge-memory", "SKILL.md"),
	} {
		skill := read(t, path)
		for _, expected := range []string{
			"upsert_project_memory",
			"without asking for separate confirmation",
			"Do not auto-scan",
		} {
			if !strings.Contains(skill, expected) {
				t.Fatalf("%s must contain %q", path, expected)
			}
		}
		for _, forbidden := range []string{
			"save_memory only after human confirmation",
			"save_memory only for human-confirmed memories",
			"do not persist candidates unless the user has approved them",
		} {
			if strings.Contains(skill, forbidden) {
				t.Fatalf("%s must not retain old confirmation-only wording %q", path, forbidden)
			}
		}
	}
}

func TestMCPDocsCoverPublishedTools(t *testing.T) {
	for _, path := range []string{
		filepath.Join("docs", "mcp.md"),
		filepath.Join("docs", "zh-CN", "mcp.md"),
	} {
		doc := read(t, path)
		for _, expected := range []string{
			"search_memory",
			"compile_context",
			"save_memory",
			"upsert_project_memory",
			"list_constraints",
			"get_project_context",
			"hybrid",
		} {
			if !strings.Contains(doc, expected) {
				t.Fatalf("%s must document %q", path, expected)
			}
		}
		for _, forbidden := range []string{
			"human-confirmed memory",
			"人工确认的 memory",
		} {
			if strings.Contains(doc, forbidden) {
				t.Fatalf("%s must not retain old MCP wording %q", path, forbidden)
			}
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
		"build-plugin-binaries",
		"package-plugins",
		"smoke-plugin-runtime",
		"./tools/package_plugins.sh",
		"GOOS=darwin GOARCH=arm64",
		"GOOS=windows GOARCH=amd64",
		"package-plugins: build-plugin-binaries",
		"plugins/claude-code/memforge/bin/memforge-mcp-launcher.js",
		"smoke-plugin/claude-code/memforge",
		"$(GO) build -trimpath",
		"MEMFORGE_PLUGIN_ROOT",
		"tools/list",
	} {
		if !strings.Contains(makefile, expected) {
			t.Fatalf("Makefile plugin automation must contain %q", expected)
		}
	}
}

func TestPluginLaunchersForwardProjectRoot(t *testing.T) {
	for _, path := range []string{
		filepath.Join("plugins", "claude-code", "memforge", "bin", "memforge-mcp-launcher.js"),
		filepath.Join("plugins", "codex", "memforge", "bin", "memforge-mcp-launcher.js"),
	} {
		launcher := read(t, path)
		for _, expected := range []string{
			"MEMFORGE_PROJECT_ROOT",
			"process.env.PWD",
			"args.push('--root', projectRoot)",
		} {
			if !strings.Contains(launcher, expected) {
				t.Fatalf("%s must forward project root with %q", path, expected)
			}
		}
	}
}

func TestCurlInstallPathIsDocumentedAndGuarded(t *testing.T) {
	scriptPath := filepath.Join(repoRoot(t), "scripts", "install.sh")
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("curl install script must exist: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("curl install script must be executable: mode %v", info.Mode())
	}

	script := read(t, "scripts", "install.sh")
	for _, expected := range []string{
		"set -euo pipefail",
		"MEMFORGE_VERSION",
		"MEMFORGE_INSTALL_DIR",
		"https://api.github.com/repos/MagnumGOYB/memforge/releases/latest",
		"browser_download_url",
		"memforge-darwin-arm64",
		"memforge-darwin-amd64",
		"memforge-linux-arm64",
		"memforge-linux-amd64",
		"install -m 0755",
	} {
		if !strings.Contains(script, expected) {
			t.Fatalf("curl install script must contain %q", expected)
		}
	}

	for _, path := range []string{"README.md", "README.zh-CN.md"} {
		doc := read(t, path)
		for _, expected := range []string{
			"curl -fsSL https://raw.githubusercontent.com/MagnumGOYB/memforge/main/scripts/install.sh | bash",
			"MEMFORGE_INSTALL_DIR",
			"MEMFORGE_VERSION",
		} {
			if !strings.Contains(doc, expected) {
				t.Fatalf("%s must document curl install path with %q", path, expected)
			}
		}
	}
}

func mustReadJSON(t *testing.T, path string, out any) {
	t.Helper()
	if err := json.Unmarshal([]byte(read(t, path)), out); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}

func assertClaudeMCPLauncher(t *testing.T, path string) {
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
	if read(t, filepath.Join(filepath.Dir(path), "bin", "memforge-mcp-launcher.js")) == "" {
		t.Fatalf("empty MCP launcher referenced by %s", path)
	}
}

func assertCodexMCPLauncher(t *testing.T, path string) {
	t.Helper()
	var config struct {
		MCPServers map[string]struct {
			Command                  string            `json:"command"`
			Args                     []string          `json:"args"`
			CWD                      string            `json:"cwd"`
			Env                      map[string]string `json:"env"`
			DefaultToolsApprovalMode string            `json:"default_tools_approval_mode"`
		} `json:"mcpServers"`
	}
	mustReadJSON(t, path, &config)
	server, ok := config.MCPServers["memforge"]
	if !ok {
		t.Fatalf("%s missing memforge MCP server", path)
	}
	if server.Command != "node" || len(server.Args) != 1 || server.Args[0] != "./bin/memforge-mcp-launcher.js" || server.CWD != "." {
		t.Fatalf("unexpected Codex MCP launcher in %s: %#v", path, server)
	}
	if _, ok := server.Env["MEMFORGE_PLUGIN_ROOT"]; ok {
		t.Fatalf("Codex MCP launcher must not rely on plugin root env in %s: %#v", path, server.Env)
	}
	if server.DefaultToolsApprovalMode != "approve" {
		t.Fatalf("unexpected MCP approval mode in %s: %q", path, server.DefaultToolsApprovalMode)
	}
	if read(t, filepath.Join(filepath.Dir(path), "bin", "memforge-mcp-launcher.js")) == "" {
		t.Fatalf("empty MCP launcher referenced by %s", path)
	}
}
