# Agent integrations

[English](../integrations.md)

MemForge 面向通过 CLI 或 MCP server 的 agent 调用。

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

已安装的 Claude Code 与 Codex plugin 可以在 active thread 中，由 agent 判断是否调用 `upsert_project_memory` 保存或更新稳定 project memory。这不同于 CLI `after`：`after` 仍是 proposal-first，并且持久化前需要显式 approval。

## Shared memory acceptance

Claude Code 和 Codex 只有在使用同一个绝对路径 `MEMFORGE_HOME`，并解析到同一个项目根目录时，才会共享 memforge 记忆。验收时显式设置 storage root 和 `--root`，避免结果依赖任一宿主的当前工作目录或私有记忆系统。

```bash
export MEMFORGE_HOME="/absolute/path/to/memforge-home"

# 在 Claude Code 中执行，或用 shell 代表 Claude Code：
memforge --no-version-check --root /absolute/path/to/project remember \
  --kind decision \
  --title "Shared memory smoke" \
  --format json \
  "Claude Code and Codex can read this memory through the same memforge store."

# 在 Codex 中执行，或用 shell 代表 Codex：
memforge --no-version-check --root /absolute/path/to/project search \
  --format json \
  "Shared memory smoke"
```

当第二条命令能从同一个 `MEMFORGE_HOME` 和同一个项目根目录返回已保存标题时，验收通过。验证两个已安装插件时，再反向执行一次 smoke：通过 Codex 写入，然后通过 Claude Code 搜索。如果任一侧使用了不同的 storage root、项目根目录或项目标识，记忆会按设计隔离，而不是共享。
