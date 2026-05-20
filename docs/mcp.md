# MCP server

[中文](zh-CN/mcp.md)

`memforge mcp` runs a stdio MCP server for local project memory access. It uses the same storage rules as the CLI: memories stay under `MEMFORGE_HOME` or `$XDG_DATA_HOME/memforge`, markdown is canonical, and the server does not auto-scan the repository or call remote providers.

## Run

```bash
memforge --no-version-check mcp --root /path/to/project
```

The server speaks newline-delimited JSON-RPC over stdin/stdout.

## Tools

### `search_memory`

Input schema:

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["query"],
  "properties": {
    "query": { "type": "string" },
    "kinds": { "type": "array", "items": { "type": "string" } },
    "limit": { "type": "integer", "minimum": 0 }
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
    "budget": { "type": "integer", "minimum": 0 },
    "kinds": { "type": "array", "items": { "type": "string" } }
  }
}
```

Returns agent-ready markdown context compiled from stored memories.

### `save_memory`

Input schema:

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["kind", "title", "content"],
  "properties": {
    "kind": { "type": "string" },
    "title": { "type": "string" },
    "content": { "type": "string" },
    "tags": { "type": "array", "items": { "type": "string" } }
  }
}
```

Persists a human-confirmed memory through the same markdown-first path as `remember`.

### `list_constraints`

Input schema:

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
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
    "task": { "type": "string" },
    "budget": { "type": "integer", "minimum": 0 }
  }
}
```

When `task` is empty, this behaves like `compile_context`. When `task` is present, it uses the same task-conditioned selection strategy as `before`.
