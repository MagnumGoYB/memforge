# Bundled Plugin Marketplace Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Claude Code marketplace install usable without a preinstalled `memforge` CLI, then align Codex marketplace/MCP usage, and automate build/package/validation/release through CI/CD.

**Architecture:** Keep source builds and user plugin installs separate. CI builds platform-specific `memforge` binaries, stages them into plugin package directories, and validates plugin MCP launch through a cross-platform Node launcher that selects the bundled binary using `${CLAUDE_PLUGIN_ROOT}`. Development can still use `go install`, but marketplace users get a self-contained plugin package.

**Tech Stack:** Go 1.26.3, Claude Code plugin manifests, Codex plugin manifests, Node.js launcher, shell packaging scripts, Go harness tests, GitHub Actions.

---

## File structure

- Create: `plugins/claude-code/memforge/bin/memforge-mcp-launcher.js`
  - Cross-platform launcher used by Claude Code MCP config.
- Create: `plugins/codex/memforge/bin/memforge-mcp-launcher.js`
  - Same launcher shape for Codex package.
- Create: `tools/package_plugins.sh`
  - Builds/stages plugin runtime directories from prebuilt artifacts or local build inputs.
- Modify: `plugins/claude-code/memforge/.mcp.json`
  - Switch from `command: "memforge"` to `node ${CLAUDE_PLUGIN_ROOT}/bin/memforge-mcp-launcher.js`.
- Modify: `plugins/codex/memforge/.mcp.json`
  - Switch to bundled launcher and keep `default_tools_approval_mode: "approve"`.
- Modify: `harness/plugin_distribution_test.go`
  - Validate launcher config and required packaged platform directories.
- Modify: `Makefile`
  - Add package/smoke targets using project-local Go cache.
- Create: `.github/workflows/release.yml`
  - Build cross-platform binaries, package plugins, validate manifests, smoke MCP, upload release assets.
- Modify: `docs/plugin-distribution.md`
- Modify: `docs/zh-CN/plugin-distribution.md`
- Modify: `plugins/claude-code/memforge/README.md`
- Modify: `plugins/codex/memforge/README.md`

## Task 1: Add Claude Code bundled MCP launcher

**Files:**
- Create: `plugins/claude-code/memforge/bin/memforge-mcp-launcher.js`
- Modify: `plugins/claude-code/memforge/.mcp.json`
- Test: `harness/plugin_distribution_test.go`

- [ ] **Step 1: Write failing harness assertions for Claude launcher config**

Modify `TestClaudeCodePluginPackageIsInstallable` so it expects the Claude MCP server to launch Node with `${CLAUDE_PLUGIN_ROOT}`:

```go
assertMCPLauncher(t, filepath.Join("plugins", "claude-code", "memforge", ".mcp.json"), false)
```

Add this helper below `assertMCPApprovalMode`:

```go
func assertMCPLauncher(t *testing.T, path string, wantApproval bool) {
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
	if server.Command != "node" {
		t.Fatalf("unexpected MCP command in %s: %q", path, server.Command)
	}
	if len(server.Args) != 1 || server.Args[0] != "${CLAUDE_PLUGIN_ROOT}/bin/memforge-mcp-launcher.js" {
		t.Fatalf("unexpected MCP launcher args in %s: %#v", path, server.Args)
	}
	if server.Env["MEMFORGE_PLUGIN_ROOT"] != "${CLAUDE_PLUGIN_ROOT}" {
		t.Fatalf("unexpected MCP launcher env in %s: %#v", path, server.Env)
	}
	if wantApproval && server.DefaultToolsApprovalMode != "approve" {
		t.Fatalf("unexpected MCP approval mode in %s: %q", path, server.DefaultToolsApprovalMode)
	}
}
```

Keep `assertMCPCommand` temporarily until Codex is migrated in Task 2, then remove it after both plugin configs use launchers.

- [ ] **Step 2: Run harness to verify it fails**

Run:

```bash
make test-harness
```

Expected: FAIL because `plugins/claude-code/memforge/.mcp.json` still has `command: "memforge"`.

- [ ] **Step 3: Create Claude launcher**

Create `plugins/claude-code/memforge/bin/memforge-mcp-launcher.js`:

```js
#!/usr/bin/env node

const { spawn } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const root = process.env.MEMFORGE_PLUGIN_ROOT || path.resolve(__dirname, "..");

const platformMap = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const archMap = {
  x64: "amd64",
  arm64: "arm64",
};

const os = platformMap[process.platform];
const arch = archMap[process.arch];

if (!os || !arch) {
  console.error(`Unsupported platform: ${process.platform}-${process.arch}`);
  process.exit(1);
}

const exe = os === "windows" ? "memforge.exe" : "memforge";
const binary = path.join(root, "bin", `${os}-${arch}`, exe);

if (!fs.existsSync(binary)) {
  console.error(`Bundled memforge binary not found: ${binary}`);
  process.exit(1);
}

const child = spawn(binary, ["--no-version-check", "mcp"], {
  stdio: "inherit",
  env: process.env,
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 1);
});

child.on("error", (error) => {
  console.error(error.message);
  process.exit(1);
});
```

- [ ] **Step 4: Update Claude `.mcp.json`**

Replace `plugins/claude-code/memforge/.mcp.json` with:

```json
{
  "mcpServers": {
    "memforge": {
      "command": "node",
      "args": ["${CLAUDE_PLUGIN_ROOT}/bin/memforge-mcp-launcher.js"],
      "env": {
        "MEMFORGE_PLUGIN_ROOT": "${CLAUDE_PLUGIN_ROOT}"
      }
    }
  }
}
```

- [ ] **Step 5: Run harness to verify config passes**

Run:

```bash
make test-harness
```

Expected: PASS for Claude launcher config. It may still validate Codex old command until Task 2.

- [ ] **Step 6: Commit**

```bash
git add plugins/claude-code/memforge/bin/memforge-mcp-launcher.js plugins/claude-code/memforge/.mcp.json harness/plugin_distribution_test.go
git commit -m "✨ feat(plugin): add bundled Claude Code launcher"
```

## Task 2: Add Codex bundled MCP launcher

**Files:**
- Create: `plugins/codex/memforge/bin/memforge-mcp-launcher.js`
- Modify: `plugins/codex/memforge/.mcp.json`
- Modify: `harness/plugin_distribution_test.go`

- [ ] **Step 1: Update harness to expect launcher for Codex**

In `TestCodexPluginPackageIsInstallable`, replace:

```go
assertMCPCommand(t, filepath.Join("plugins", "codex", "memforge", ".mcp.json"))
assertMCPApprovalMode(t, filepath.Join("plugins", "codex", "memforge", ".mcp.json"), "approve")
```

with:

```go
assertMCPLauncher(t, filepath.Join("plugins", "codex", "memforge", ".mcp.json"), true)
```

Remove `assertMCPCommand` and `assertMCPApprovalMode` after no tests call them.

- [ ] **Step 2: Run harness to verify it fails**

Run:

```bash
make test-harness
```

Expected: FAIL because Codex `.mcp.json` still has `command: "memforge"`.

- [ ] **Step 3: Create Codex launcher**

Create `plugins/codex/memforge/bin/memforge-mcp-launcher.js` with the same contents as the Claude launcher in Task 1 Step 3.

- [ ] **Step 4: Update Codex `.mcp.json`**

Replace `plugins/codex/memforge/.mcp.json` with:

```json
{
  "mcpServers": {
    "memforge": {
      "command": "node",
      "args": ["${CLAUDE_PLUGIN_ROOT}/bin/memforge-mcp-launcher.js"],
      "env": {
        "MEMFORGE_PLUGIN_ROOT": "${CLAUDE_PLUGIN_ROOT}"
      },
      "default_tools_approval_mode": "approve"
    }
  }
}
```

If Codex later documents a different plugin-root variable, update both `.mcp.json` and harness in the same task. Keep the launcher fallback to `__dirname` so direct local `codex mcp add` smoke can still run when invoked from the package path.

- [ ] **Step 5: Run harness to verify both launcher configs pass**

Run:

```bash
make test-harness
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add plugins/codex/memforge/bin/memforge-mcp-launcher.js plugins/codex/memforge/.mcp.json harness/plugin_distribution_test.go
git commit -m "✨ feat(plugin): add bundled Codex launcher"
```

## Task 3: Add plugin runtime packaging script

**Files:**
- Create: `tools/package_plugins.sh`
- Modify: `Makefile`
- Test: `harness/plugin_distribution_test.go`

- [ ] **Step 1: Write failing harness assertions for staged platform directories**

Add this test to `harness/plugin_distribution_test.go`:

```go
func TestPluginRuntimePlatformsAreDeclared(t *testing.T) {
	for _, root := range []string{
		filepath.Join("plugins", "claude-code", "memforge"),
		filepath.Join("plugins", "codex", "memforge"),
	} {
		for _, platform := range []string{"darwin-arm64", "darwin-amd64", "linux-arm64", "linux-amd64", "windows-arm64", "windows-amd64"} {
			path := filepath.Join(root, "bin", platform, "README.md")
			if read(t, path) == "" {
				t.Fatalf("missing runtime platform placeholder: %s", path)
			}
		}
	}
}
```

This checks placeholders in source. CI release packaging will replace placeholders with binaries in release artifacts.

- [ ] **Step 2: Run harness to verify it fails**

Run:

```bash
make test-harness
```

Expected: FAIL because platform placeholder directories do not exist yet.

- [ ] **Step 3: Add platform placeholder files**

Create these files with content `Runtime binary is injected by release packaging. Do not edit this placeholder.`:

```txt
plugins/claude-code/memforge/bin/darwin-arm64/README.md
plugins/claude-code/memforge/bin/darwin-amd64/README.md
plugins/claude-code/memforge/bin/linux-arm64/README.md
plugins/claude-code/memforge/bin/linux-amd64/README.md
plugins/claude-code/memforge/bin/windows-arm64/README.md
plugins/claude-code/memforge/bin/windows-amd64/README.md
plugins/codex/memforge/bin/darwin-arm64/README.md
plugins/codex/memforge/bin/darwin-amd64/README.md
plugins/codex/memforge/bin/linux-arm64/README.md
plugins/codex/memforge/bin/linux-amd64/README.md
plugins/codex/memforge/bin/windows-arm64/README.md
plugins/codex/memforge/bin/windows-amd64/README.md
```

- [ ] **Step 4: Create packaging script**

Create `tools/package_plugins.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST="${DIST:-$ROOT/dist}"
VERSION="${VERSION:-$(cat "$ROOT/VERSION")}"

platforms=(
  darwin-arm64
  darwin-amd64
  linux-arm64
  linux-amd64
  windows-arm64
  windows-amd64
)

rm -rf "$DIST/plugins"
mkdir -p "$DIST/plugins/claude-code" "$DIST/plugins/codex"

copy_plugin() {
  local source="$1"
  local target="$2"
  mkdir -p "$(dirname "$target")"
  cp -R "$source" "$target"
}

copy_plugin "$ROOT/plugins/claude-code/memforge" "$DIST/plugins/claude-code/memforge"
copy_plugin "$ROOT/plugins/codex/memforge" "$DIST/plugins/codex/memforge"

for platform in "${platforms[@]}"; do
  source_name="memforge-$platform"
  if [[ "$platform" == windows-* ]]; then
    source_name="$source_name.exe"
    target_name="memforge.exe"
  else
    target_name="memforge"
  fi
  source_path="$DIST/bin/$source_name"
  if [[ ! -f "$source_path" ]]; then
    echo "missing built binary: $source_path" >&2
    exit 2
  fi
  for host in claude-code codex; do
    target_dir="$DIST/plugins/$host/memforge/bin/$platform"
    rm -rf "$target_dir"
    mkdir -p "$target_dir"
    cp "$source_path" "$target_dir/$target_name"
    chmod +x "$target_dir/$target_name" || true
  done
done

(
  cd "$DIST/plugins/claude-code"
  zip -qr "../memforge-claude-code-plugin-v$VERSION.zip" memforge
)
(
  cd "$DIST/plugins/codex"
  zip -qr "../memforge-codex-plugin-v$VERSION.zip" memforge
)

echo "$DIST/plugins/memforge-claude-code-plugin-v$VERSION.zip"
echo "$DIST/plugins/memforge-codex-plugin-v$VERSION.zip"
```

- [ ] **Step 5: Add Makefile targets**

Modify `.PHONY` line to include:

```make
package-plugins smoke-plugin-runtime
```

Add targets:

```make
package-plugins:
	./tools/package_plugins.sh

smoke-plugin-runtime:
	MEMFORGE_HOME=$$(mktemp -d) node plugins/claude-code/memforge/bin/memforge-mcp-launcher.js <<<'{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

Note: `smoke-plugin-runtime` only works after a local binary is staged into the current platform directory. CI will smoke packaged artifacts after staging binaries.

- [ ] **Step 6: Run harness**

Run:

```bash
make test-harness
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add tools/package_plugins.sh Makefile plugins/claude-code/memforge/bin plugins/codex/memforge/bin harness/plugin_distribution_test.go
git commit -m "🏗️ build(plugin): package bundled runtimes"
```

## Task 4: Add release workflow for binary and plugin artifacts

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create release workflow**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"
  workflow_dispatch:

permissions:
  contents: write

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goos: darwin
            goarch: arm64
            suffix: darwin-arm64
            exe: memforge
          - goos: darwin
            goarch: amd64
            suffix: darwin-amd64
            exe: memforge
          - goos: linux
            goarch: arm64
            suffix: linux-arm64
            exe: memforge
          - goos: linux
            goarch: amd64
            suffix: linux-amd64
            exe: memforge
          - goos: windows
            goarch: arm64
            suffix: windows-arm64
            exe: memforge.exe
          - goos: windows
            goarch: amd64
            suffix: windows-amd64
            exe: memforge.exe
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: "1.26.3"
          cache: true
      - run: mkdir -p dist/bin
      - run: GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} CGO_ENABLED=0 go build -o dist/bin/${{ matrix.exe }} ./cmd/memforge
      - run: mv dist/bin/${{ matrix.exe }} dist/bin/memforge-${{ matrix.suffix }}${{ matrix.goos == 'windows' && '.exe' || '' }}
      - uses: actions/upload-artifact@v6
        with:
          name: memforge-${{ matrix.suffix }}
          path: dist/bin/memforge-${{ matrix.suffix }}*

  package:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: "1.26.3"
          cache: true
      - uses: actions/download-artifact@v6
        with:
          path: dist/artifacts
      - run: mkdir -p dist/bin && find dist/artifacts -type f -name 'memforge-*' -exec cp {} dist/bin/ \;
      - run: make validate
      - run: make test-harness
      - run: ./tools/package_plugins.sh
      - run: |
          unzip -q dist/plugins/memforge-claude-code-plugin-v$(cat VERSION).zip -d dist/smoke-claude
          MEMFORGE_HOME=$(mktemp -d) node dist/smoke-claude/memforge/bin/memforge-mcp-launcher.js <<'EOF'
          {"jsonrpc":"2.0","id":1,"method":"tools/list"}
          EOF
      - uses: softprops/action-gh-release@v2
        if: startsWith(github.ref, 'refs/tags/')
        with:
          files: |
            dist/bin/memforge-*
            dist/plugins/memforge-claude-code-plugin-v*.zip
            dist/plugins/memforge-codex-plugin-v*.zip
```

- [ ] **Step 2: Run YAML smoke locally where possible**

Run:

```bash
make validate
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "👷 ci(release): publish plugin artifacts"
```

## Task 5: Update docs for user install and developer install separation

**Files:**
- Modify: `docs/plugin-distribution.md`
- Modify: `docs/zh-CN/plugin-distribution.md`
- Modify: `plugins/claude-code/memforge/README.md`
- Modify: `plugins/codex/memforge/README.md`
- Modify: `README.md`
- Modify: `README.zh-CN.md`

- [ ] **Step 1: Update Claude Code docs**

In `docs/plugin-distribution.md`, replace the Claude Code section text that says the plugin expects `memforge` on `PATH` with:

```markdown
Marketplace users do not need to install the CLI separately. Release packages include platform-specific `memforge` binaries under the plugin `bin/` directory, and `.mcp.json` starts the bundled runtime through `bin/memforge-mcp-launcher.js`.

Developer smoke from a checkout can still use:

```bash
make build
go install ./cmd/memforge
claude --plugin-dir ./plugins/claude-code/memforge
```
```

Use the equivalent zh-CN wording in `docs/zh-CN/plugin-distribution.md`.

- [ ] **Step 2: Update plugin READMEs**

For `plugins/claude-code/memforge/README.md`, make the first install path marketplace-first:

```markdown
## User install

Install from a Claude Code marketplace. The release package includes the `memforge` runtime; users do not need to run `go install`.

## Local development

From a source checkout:

```bash
make build
go install ./cmd/memforge
claude --plugin-dir ./plugins/claude-code/memforge
```
```

For `plugins/codex/memforge/README.md`, document the current Codex CLI boundary and bundled runtime package.

- [ ] **Step 3: Update top-level README docs list if needed**

Ensure both README files still point to `docs/plugin-distribution.md` and `docs/zh-CN/plugin-distribution.md`.

- [ ] **Step 4: Run docs/harness checks**

Run:

```bash
make test-harness
make validate
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
git add README.md README.zh-CN.md docs/plugin-distribution.md docs/zh-CN/plugin-distribution.md plugins/claude-code/memforge/README.md plugins/codex/memforge/README.md
git commit -m "📝 docs(plugin): document marketplace-first install"
```

## Task 6: Local end-to-end smoke for current platform

**Files:**
- No source files required unless smoke exposes a bug.

- [ ] **Step 1: Stage current platform binary into plugin directories**

On Apple Silicon, run:

```bash
make build
mkdir -p plugins/claude-code/memforge/bin/darwin-arm64 plugins/codex/memforge/bin/darwin-arm64
cp memforge plugins/claude-code/memforge/bin/darwin-arm64/memforge
cp memforge plugins/codex/memforge/bin/darwin-arm64/memforge
```

On Intel macOS, use `darwin-amd64`. On Linux, use `linux-amd64` or `linux-arm64`.

- [ ] **Step 2: Verify bundled launcher lists tools**

Run:

```bash
MEMFORGE_HOME=$(mktemp -d) MEMFORGE_PLUGIN_ROOT=$PWD/plugins/claude-code/memforge node plugins/claude-code/memforge/bin/memforge-mcp-launcher.js <<'EOF'
{"jsonrpc":"2.0","id":1,"method":"tools/list"}
EOF
```

Expected: JSON result includes `search_memory`, `save_memory`, and `get_project_context`.

- [ ] **Step 3: Verify Claude Code plugin use without PATH memforge**

Run with a PATH that does not include `$HOME/go/bin`:

```bash
PATH="/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin" MEMFORGE_HOME=$(mktemp -d) claude -p --permission-mode bypassPermissions --plugin-dir "$PWD/plugins/claude-code/memforge" --output-format text 'Use the memforge MCP tool to save a decision memory titled "Bundled Claude plugin smoke" with content "Claude Code can use bundled memforge runtime." tagged claude-code and smoke. Then search memory for "Bundled Claude plugin smoke" and answer exactly FOUND if you find it, otherwise NOT FOUND.'
```

Expected: `FOUND`.

- [ ] **Step 4: Verify Codex exec use with bundled launcher**

Run:

```bash
PATH="/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin" MEMFORGE_HOME=$(mktemp -d) codex exec --ephemeral -C "$PWD" -s workspace-write -c 'mcp_servers.memforge.command="node"' -c "mcp_servers.memforge.args=[\"$PWD/plugins/codex/memforge/bin/memforge-mcp-launcher.js\"]" -c 'mcp_servers.memforge.env.MEMFORGE_PLUGIN_ROOT="'$PWD'/plugins/codex/memforge"' -c 'mcp_servers.memforge.default_tools_approval_mode="approve"' 'Use the memforge MCP server directly. Save a decision memory with title "Bundled Codex plugin smoke" and content "Codex can use bundled memforge runtime." tagged codex and smoke. Then search memforge for "Bundled Codex plugin smoke". Report only FOUND if the saved memory appears, otherwise NOT FOUND.'
```

Expected: `FOUND`.

- [ ] **Step 5: Clean local staged binaries before commit if they are not release artifacts**

Run:

```bash
rm -f plugins/claude-code/memforge/bin/darwin-arm64/memforge plugins/codex/memforge/bin/darwin-arm64/memforge memforge
```

Adjust paths for the current platform.

- [ ] **Step 6: Final validation**

Run:

```bash
make validate
make test-harness
```

Expected: both PASS.

## Self-review

- Spec coverage: Claude Code marketplace install without preinstalled CLI is covered by Tasks 1, 3, 4, and 6. Codex marketplace/MCP usage is covered by Tasks 2, 3, 4, and 6. CI/CD automation is covered by Task 4. Docs are covered by Task 5.
- Placeholder scan: Source tree uses explicit README placeholders for platform runtime directories; release packaging replaces them with binaries. No implementation step is left vague.
- Type consistency: MCP config helpers use `default_tools_approval_mode`, `MEMFORGE_PLUGIN_ROOT`, and `${CLAUDE_PLUGIN_ROOT}` consistently across tasks.
