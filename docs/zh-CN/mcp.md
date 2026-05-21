# MCP server

[English](../mcp.md)

`memforge mcp` 启动 MemForge 用于本地项目记忆访问的 stdio MCP server。它复用 CLI 的同一套存储规则：记忆只保存在 `MEMFORGE_HOME` 或 `$XDG_DATA_HOME/memforge` 下，markdown 是 canonical store，server 不会自动扫描仓库，也不会调用远程 provider。

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
    "limit": { "type": "integer", "minimum": 0 },
    "hybrid": { "type": "boolean" }
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

从已保存记忆编译 agent-ready markdown context。`budget` 省略或为 `0` 时，server 会使用项目配置（`.memoryrc` 或用户配置），再回退到内置默认值。

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

通过与 `remember` 相同的 markdown-first 路径持久化一条新的 project memory。当可能更新已有 memory 时，agent 应优先使用 `upsert_project_memory`。

### `upsert_project_memory`

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

按 `kind` 和归一化后的 `title` 创建或更新一条稳定 project memory。返回值包含 memory id，以及 `action: "created"` 或 `action: "updated"`。

enabled Claude Code 与 Codex plugin 可以在 active thread 中由 agent 判断是否调用该工具创建或修订稳定 project memory。该工具仍受 MCP tool approval 与本地存储规则约束：不会自动扫描源代码文件，也不会调用远程 provider。

### `check_update`

输入 schema：

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {}
}
```

检查最新发布的 MemForge release，并返回 `current`、`latest`、`has_update` 和 `update_url`。该工具需要显式调用；除非设置了 `MEMFORGE_VERSION_CHECK_LATEST` 或已有缓存的 release metadata，否则可能发起一次 GitHub release metadata 请求。

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

`task` 为空时行为类似 `compile_context`；提供 `task` 时复用 `before` 的任务条件选择策略。`budget` 省略或为 `0` 时使用项目配置提供的默认值。
