# memforge Codex plugin

This plugin connects Codex to `memforge` through its stdio MCP server.

## Packaged plugin install

Codex public self-service plugin publishing is still limited, so use a local/private marketplace or Codex host plugin package flow where supported. Packaged plugin bundles include platform-specific `memforge` runtime binaries, so users should not need to run `go install` or put a separate `memforge` CLI on `PATH` for packaged installs.

The plugin manifest is at `.codex-plugin/plugin.json`; MCP configuration is in `.mcp.json`. The MCP server starts through `bin/memforge-mcp-launcher.js`, which selects the bundled runtime for the current platform and runs `memforge --no-version-check mcp` from inside the plugin package.

Optionally set a storage root:

```bash
export MEMFORGE_HOME="$HOME/.local/share/memforge"
```

## Codex CLI development smoke

Codex CLI 0.130 exposes marketplace management but no standalone `plugin install/list/details` subcommands. For CLI smoke usage during development or debugging, add the local marketplace:

```bash
codex plugin marketplace add "$PWD"
```

If the host plugin flow is unavailable, register the MCP server explicitly as a development/debugging fallback:

```bash
codex mcp add memforge -- memforge --no-version-check mcp
```

That explicit MCP registration path uses a `memforge` binary on `PATH` and is not the packaged plugin runtime path. The Codex MCP config uses `default_tools_approval_mode: approve` so non-interactive `codex exec` can complete memforge MCP tool calls.
