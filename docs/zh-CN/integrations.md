# Agent integrations

[English](../integrations.md)

`memforge` 面向通过 CLI 或 MCP server 的 agent 调用。

## Claude Code

推荐本地命令：

```bash
memforge --no-version-check before --format json "Task description"
memforge --no-version-check search --hybrid --format json "repository framework"
memforge --no-version-check after --adapter claude-code --from session.jsonl --format json
```

只有在人工 review candidate memories 后，才使用 `after --approve all`。

## Codex 与 Cursor

对于 JSONL 日志，使用显式 adapter 名称：

```bash
memforge --no-version-check after --adapter codex --from session.jsonl --format json
memforge --no-version-check after --adapter cursor --from session.jsonl --format json
```

adapter 会从常见的 `content`、`text`、`message`、`input`、`output` 字段抽取文本。它们不会自动扫描 session 目录。

## MCP

当宿主支持 stdio MCP server 时使用 `memforge mcp`。工具 schema 见 `docs/mcp.md`。
