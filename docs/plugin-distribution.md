# Plugin distribution

[中文](zh-CN/plugin-distribution.md)

This repository ships plugin packages for Claude Code and Codex. The release packages are designed for normal users to install from a marketplace or release bundle without first installing the `memforge` CLI on `PATH`.

## Claude Code

Claude Code supports plugin marketplaces. This repository includes:

```txt
plugins/claude-code/memforge/.claude-plugin/plugin.json
plugins/claude-code/memforge/.mcp.json
.claude-plugin/marketplace.json
```

### Normal user install

The normal user path is a marketplace or release plugin install:

```txt
/plugin marketplace add <marketplace-or-release-catalog>
/plugin install memforge@<marketplace-name>
/reload-plugins
```

Release plugin packages include platform-specific `memforge` runtime binaries under the plugin's `bin/<platform>/` directory. Users do not need to run `go install` or put a separately installed `memforge` binary on `PATH` before installing the Claude Code marketplace package.

The plugin MCP configuration starts the server through the bundled Node launcher:

```json
{
  "command": "node",
  "args": ["${CLAUDE_PLUGIN_ROOT}/bin/memforge-mcp-launcher.js"],
  "env": {
    "MEMFORGE_PLUGIN_ROOT": "${CLAUDE_PLUGIN_ROOT}"
  }
}
```

The launcher resolves the current platform, selects the bundled runtime, and starts the stdio MCP server from that runtime.

### Local development smoke

Source checkout developers can still smoke the plugin directly from this repository:

```bash
make build
go install ./cmd/memforge
claude --plugin-dir ./plugins/claude-code/memforge
```

This source-checkout flow is for local development and debugging only. It is not the normal user install path and should not be required for marketplace/release installs.

## Codex

> **Official Codex plugin store status:** OpenAI's self-serve plugin publishing is not yet available (marked "coming soon" in OpenAI docs as of May 2026). MemForge is not listed in the Codex official plugin browser (`/plugins`). When public publishing opens and MemForge is accepted, users will be able to install directly from the Codex plugin browser without any manual setup.

Codex supports plugin manifests and marketplace/catalog installation flows. This repository includes a local/private distribution package:

```txt
dist/plugins/codex/memforge/.codex-plugin/plugin.json
dist/plugins/codex/memforge/.mcp.json
.agents/plugins/marketplace.json
```

Packaged Codex plugin bundles include the same platform-specific `memforge` runtimes and use `bin/memforge-mcp-launcher.js` through the MCP configuration. Where a Codex host supports local/private marketplace or plugin package installation, users should not need to preinstall the `memforge` CLI on `PATH`.

Codex CLI 0.132 exposes marketplace management and plugin install/remove commands. The `memforge-codex-marketplace` entry points at the packaged bundle under `dist/plugins/codex/memforge`, so run `make package-plugins` before adding or refreshing the marketplace from a source checkout:

```bash
make package-plugins
codex plugin marketplace add "$PWD"
codex plugin add memforge@memforge-codex-marketplace
```

For day-to-day CLI smoke, the simpler fallback is direct MCP registration:

```bash
go install ./cmd/memforge
codex mcp add memforge -- memforge --no-version-check mcp
```

Direct `codex mcp add` avoids carrying an extra local marketplace entry, but it is a development/debugging fallback rather than the packaged plugin runtime path. It uses the `memforge` binary on `PATH`; packaged Codex plugin bundles use the bundled launcher instead. The Codex MCP config sets `default_tools_approval_mode` to `approve` so non-interactive `codex exec` can complete memforge MCP tool calls.

## Release and CI packaging

The GitHub release workflow builds multi-platform `memforge` binaries, runs repository and harness validation, packages the Claude Code and Codex plugin zips with `tools/package_plugins.sh`, smokes the bundled Claude runtime through the Node launcher, and uploads both standalone binaries and plugin zip assets to the release.

## Real smoke target

Repository validation checks plugin manifests, launcher configuration, packaging, and release workflow structure. Runtime smoke should verify:

1. A marketplace/release plugin can start without a separate `memforge` binary on `PATH`.
2. `bin/memforge-mcp-launcher.js` selects the bundled runtime for the current platform.
3. The bundled runtime responds to MCP `tools/list`.
4. `save_memory` can persist a memory.
5. `search_memory` can retrieve it.
