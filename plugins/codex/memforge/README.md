# memforge Codex plugin

This plugin connects Codex to `memforge` through its stdio MCP server.

> **Official Codex plugin store:** OpenAI's self-serve plugin publishing is not yet available (marked "coming soon" as of May 2026). MemForge is not listed in the Codex official plugin browser. When public publishing opens and MemForge is accepted, users will be able to install directly via `/plugins` without manual setup.

## Packaged plugin install

Use a local/private marketplace or Codex host plugin package flow where supported. Packaged plugin bundles include platform-specific `memforge` runtime binaries, so users should not need to run `go install` or put a separate `memforge` CLI on `PATH` for packaged installs.

The plugin manifest is at `.codex-plugin/plugin.json`; MCP configuration is in `.mcp.json`. The MCP server starts through `bin/memforge-mcp-launcher.js`, which selects the bundled runtime for the current platform and runs `memforge --no-version-check mcp` from inside the plugin package.

Optionally set a storage root:

```bash
export MEMFORGE_HOME="$HOME/.local/share/memforge"
```

## Codex CLI development smoke

Codex CLI 0.132 exposes marketplace management and plugin install/remove commands. The extra `memforge-codex-marketplace` entry comes from adding this repository as a local/private marketplace for development validation.

Build the packaged bundle before using that marketplace entry:

```bash
make package-plugins
codex plugin marketplace add "$PWD"
codex plugin add memforge@memforge-codex-marketplace
```

For day-to-day CLI development smoke, the simpler fallback is to register the MCP server directly:

```bash
go install ./cmd/memforge
codex mcp add memforge -- memforge --no-version-check mcp
```

Direct `codex mcp add` avoids carrying an extra local marketplace entry, but it uses a `memforge` binary on `PATH` and is not the packaged plugin runtime path. The Codex MCP config uses `default_tools_approval_mode: approve` so non-interactive `codex exec` can complete memforge MCP tool calls.
