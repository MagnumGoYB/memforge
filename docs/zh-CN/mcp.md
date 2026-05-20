# MCP server

[English](../mcp.md)

`memforge mcp` 启动一个用于本地项目记忆访问的 stdio MCP server。它复用 CLI 的同一套存储规则：记忆只保存在 `MEMFORGE_HOME` 或 `$XDG_DATA_HOME/memforge` 下，markdown 是 canonical store，server 不会自动扫描仓库，也不会调用远程 provider。

## 运行

```bash
memforge --no-version-check mcp --root /path/to/project
```

server 通过 stdin/stdout 使用 newline-delimited JSON-RPC。

## Tools

### `search_memory`

输入 schema：

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

从 SQLite FTS index 返回排序后的匹配结果。

### `compile_context`

输入 schema：

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

从已保存记忆编译 agent-ready markdown context。

### `save_memory`

输入 schema：

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

通过与 `remember` 相同的 markdown-first 路径持久化一条人工确认的 memory。

### `list_constraints`

输入 schema：

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "limit": { "type": "integer", "minimum": 0 }
  }
}
```

返回已保存的 `constraint` memories。

### `get_project_context`

输入 schema：

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

`task` 为空时行为类似 `compile_context`；提供 `task` 时复用 `before` 的任务条件选择策略。
