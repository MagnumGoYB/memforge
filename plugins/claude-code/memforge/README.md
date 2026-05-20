# memforge Claude Code plugin

This plugin connects Claude Code to `memforge` through its stdio MCP server.

## Marketplace or release install

Marketplace/release installation is the normal user path. Release plugin packages include platform-specific `memforge` runtime binaries, so users do not run `go install` and do not need a separate `memforge` CLI on `PATH` before installing the plugin.

The MCP server starts through the bundled Node launcher configured in `.mcp.json`:

```json
{
  "command": "node",
  "args": ["${CLAUDE_PLUGIN_ROOT}/bin/memforge-mcp-launcher.js"],
  "env": {
    "MEMFORGE_PLUGIN_ROOT": "${CLAUDE_PLUGIN_ROOT}"
  }
}
```

The launcher selects the current platform's bundled runtime and starts `memforge --no-version-check mcp` from inside the plugin package.

Set a local storage root if desired:

```bash
export MEMFORGE_HOME="$HOME/.local/share/memforge"
```

## Local development smoke

From this repository, source checkout developers can run:

```bash
make build
go install ./cmd/memforge
claude --plugin-dir ./plugins/claude-code/memforge
```

Then in Claude Code, reload plugins and verify the `memforge` MCP tools are available.

This local source-build flow is for development and debugging only. It is not the normal marketplace/release install path.
