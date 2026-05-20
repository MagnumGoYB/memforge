# Agent integrations

[中文](zh-CN/integrations.md)

MemForge is designed for agent invocation through the CLI or MCP server.

## Claude Code

Recommended local commands:

```bash
memforge --no-version-check before --format json "Task description"
memforge --no-version-check search --hybrid --format json "repository framework"
memforge --no-version-check after --adapter claude-code --from session.jsonl --format json
```

Use `after --approve all` only after the candidate memories have been reviewed.

## Codex and Cursor

For JSONL logs, use the explicit adapter names:

```bash
memforge --no-version-check after --adapter codex --from session.jsonl --format json
memforge --no-version-check after --adapter cursor --from session.jsonl --format json
```

The adapters extract text from common `content`, `text`, `message`, `input`, and `output` fields. They do not auto-scan session directories.

## MCP

Use `memforge mcp` when the host supports stdio MCP servers. See `docs/mcp.md` for tool schemas.
