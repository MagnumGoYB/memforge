# Memory 格式

[English](../memory-format.md)

MemForge 将项目记忆保存为按 kind 分组的 markdown block。Markdown 是 canonical store；SQLite 是可由 `memforge reindex` 重建的派生索引。

## Kind 文件映射

记忆存放在：

```txt
$MEMFORGE_HOME/projects/{project_id}/memories/
```

当前文件映射：

- `manual` → `manual.md`
- `constraint` → `constraints.md`
- `convention` → `conventions.md`
- `decision` → `decisions.md`
- `bugfix` → `bugfixes.md`
- `api-contract` → `api-contracts.md`
- `agent-instruction` → `agent-instructions.md`

## Canonical block 形态

每条 memory 以一个 append-only block 保存：

```md
<!-- memforge:memory id=01JV... kind=decision -->
---
title: "Repository layer must remain framework-agnostic"
tags: ["architecture", "repository"]
source: "cli"
confidence: 1
created_at: 2026-05-19T10:00:00Z
updated_at: 2026-05-19T10:00:00Z
---

Body in markdown.

<!-- /memforge:memory -->
```

规则：

- `id` 使用 ULID。
- marker 中的 `kind` 必须与目标文件的 kind 一致。
- `title`、`tags`、`confidence`、`created_at`、`updated_at` 为必填 frontmatter 字段。
- `source` 为可选字段。
- Body 保持 markdown 原文，仅在创建时做首尾 trim。

## Source-of-truth 规则

- `remember` 先 append markdown block。
- 然后用同一条 record upsert SQLite `memories` 行。
- `reindex` 完全从 markdown block 重建 SQLite。
- 即使 SQLite 被删除或失同步，markdown 仍然是权威来源。

## 命令语义

- `remember` 将一条 block 写入匹配的 kind 文件。
- `search` 查询 SQLite FTS5 并返回排序后的匹配结果。
- `context` 从已保存记忆编译按组输出的 markdown。
- `before` 合并高优先级记忆与任务相关匹配。
- `after` 从显式提供的 session 文件提取 candidate memories，报告 duplicates 与 merge proposals，并且只为显式确认的 candidates 写入 markdown。
- MCP `upsert_project_memory` 按 `kind` 和归一化后的 `title` 创建或更新一条稳定 memory，然后从 markdown 重建 SQLite index。
- `reindex` 重新解析全部 markdown block 并重建 SQLite 索引。

## Confidence decay

Markdown 中保存的 `confidence` 仍然是 canonical base value。`search` 与 `context` 在读取/排序时计算 effective confidence，因此 decay 本身不会重写 memory 文件。
