# MCP server

[中文](zh-CN/mcp.md)

`memforge mcp` runs MemForge's stdio MCP server for local project memory access. It uses the same storage rules as the CLI: memories stay under `MEMFORGE_HOME` or `$XDG_DATA_HOME/memforge`, markdown is canonical, and the server does not auto-scan the repository or call remote providers.

## Run

```bash
memforge --no-version-check mcp --root /path/to/project
```

The server speaks newline-delimited JSON-RPC over stdin/stdout.

## Tools

All tools accept optional `project_root` as an absolute repository/workspace path. When omitted, the MCP server uses the project root resolved at server startup. Plugins should pass `project_root` because hosts may keep MCP server processes running from plugin cache directories.

### `search_memory`

Input schema:

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["query"],
  "properties": {
    "project_root": { "type": "string" },
    "query": { "type": "string" },
    "kinds": { "type": "array", "items": { "type": "string" } },
    "limit": { "type": "integer", "minimum": 0 },
    "hybrid": { "type": "boolean" }
  }
}
```

Returns ranked matches from the SQLite FTS index.

### `compile_context`

Input schema:

```json
{
  "type": "object",
	  "additionalProperties": false,
  "properties": {
    "project_root": { "type": "string" },
    "budget": { "type": "integer", "minimum": 0 },
    "kinds": { "type": "array", "items": { "type": "string" } }
  }
}
```

Returns agent-ready markdown context compiled from stored memories. When `budget` is omitted or `0`, the server uses project configuration (`.memoryrc` or user config) and then the built-in default.

The response also includes `estimated_tokens` and `usage.estimated_tokens`, which report MemForge's local estimate for the compiled context payload. This is useful for plugin-side observability and budget tuning, but it is not a full host-model billing record.

### `save_memory`

Input schema:

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["kind", "title", "content"],
  "properties": {
    "project_root": { "type": "string" },
    "kind": { "type": "string" },
    "title": { "type": "string" },
    "content": { "type": "string" },
    "tags": { "type": "array", "items": { "type": "string" } }
  }
}
```

Persists a new project memory through the same markdown-first path as `remember`. Agents should prefer `upsert_project_memory` when they may be updating an existing memory.

### `upsert_project_memory`

Input schema:

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["kind", "title", "content"],
  "properties": {
    "project_root": { "type": "string" },
    "kind": { "type": "string" },
    "title": { "type": "string" },
    "content": { "type": "string" },
    "tags": { "type": "array", "items": { "type": "string" } }
  }
}
```

Creates or updates a stable project memory by `kind` and normalized `title`. It returns `action: "created"` or `action: "updated"` with the memory id.

Enabled Claude Code and Codex plugins may use this tool during an active thread when the agent determines a durable project memory should be created or revised. The tool is still bounded by MCP tool approval and local storage rules: it does not auto-scan source files and it does not call remote providers.

### `check_update`

Input schema:

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {}
}
```

Checks the latest published MemForge release and returns `current`, `latest`, `has_update`, and `update_url`. This tool is explicit and may perform a GitHub release metadata request unless `MEMFORGE_VERSION_CHECK_LATEST` or cached release metadata is available.

### `list_constraints`

Input schema:

```json
{
  "type": "object",
	  "additionalProperties": false,
  "properties": {
    "project_root": { "type": "string" },
    "limit": { "type": "integer", "minimum": 0 }
  }
}
```

Returns stored `constraint` memories.

### `get_project_context`

Input schema:

```json
{
  "type": "object",
	  "additionalProperties": false,
  "properties": {
    "project_root": { "type": "string" },
    "task": { "type": "string" },
    "budget": { "type": "integer", "minimum": 0 }
  }
}
```

When `task` is empty, this behaves like `compile_context`. When `task` is present, it uses the same task-conditioned selection strategy as `before`. When `budget` is omitted or `0`, project configuration supplies the default.

Like `compile_context`, the response includes `estimated_tokens` and `usage.estimated_tokens` for the compiled context returned by MemForge.
