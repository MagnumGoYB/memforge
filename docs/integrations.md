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

Installed Claude Code and Codex plugins can call `upsert_project_memory` during an active thread when the agent decides a stable project memory should be saved or updated. This is different from CLI `after`: `after` remains proposal-first and requires explicit approval before persistence.

## Shared memory acceptance

Claude Code and Codex share memforge memories when they use the same absolute `MEMFORGE_HOME` and resolve the same project root. Validate sharing with an explicit storage root and `--root` so the test does not depend on either host's working directory or private memory system.

```bash
export MEMFORGE_HOME="/absolute/path/to/memforge-home"

# In Claude Code, or a shell standing in for Claude Code:
memforge --no-version-check --root /absolute/path/to/project remember \
  --kind decision \
  --title "Shared memory smoke" \
  --format json \
  "Claude Code and Codex can read this memory through the same memforge store."

# In Codex, or a shell standing in for Codex:
memforge --no-version-check --root /absolute/path/to/project search \
  --format json \
  "Shared memory smoke"
```

Acceptance passes when the second command returns the saved title from the same `MEMFORGE_HOME` and same project root. Repeat the smoke in the opposite direction when validating both installed plugins: save through Codex, then search through Claude Code. If either side uses a different storage root, project root, or project identifier, the memories are intentionally isolated rather than shared.
