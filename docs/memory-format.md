# Memory format

[中文](zh-CN/memory-format.md)

MemForge stores project memories as markdown blocks grouped by kind. Markdown is the canonical store; SQLite is a derived index rebuilt by `memforge reindex`.

## Kind files

Memories live under:

```txt
$MEMFORGE_HOME/projects/{project_id}/memories/
```

Current file mapping:

- `manual` → `manual.md`
- `constraint` → `constraints.md`
- `convention` → `conventions.md`
- `decision` → `decisions.md`
- `bugfix` → `bugfixes.md`
- `api-contract` → `api-contracts.md`
- `agent-instruction` → `agent-instructions.md`

Accepted kind aliases are normalized before storage: `note` stores as `manual`, `domain` stores as `convention`, and `workflow` stores as `agent-instruction`.

## Canonical block shape

Each memory is stored as one append-only block:

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

Rules:

- `id` is a ULID.
- `kind` in the marker must match the destination file kind.
- `title`, `tags`, `confidence`, `created_at`, and `updated_at` are required frontmatter fields.
- `source` is optional.
- Body content remains markdown and is preserved as-is except for surrounding trim during creation.

## Source-of-truth rules

- `remember` appends the markdown block first.
- The SQLite `memories` row is then upserted from the same record.
- `reindex` rebuilds SQLite entirely from markdown blocks.
- If SQLite is deleted or out of sync, markdown remains authoritative.

## Command expectations

- `remember` writes one block to the matching kind file.
- `search` queries SQLite FTS5 and returns ranked matches.
- `context` compiles grouped markdown from stored memories.
- `before` merges high-priority memories with task-conditioned matches.
- `after` extracts candidate memories from an explicit session file, reports duplicates and merge proposals, and writes markdown only for explicitly approved candidates.
- MCP `upsert_project_memory` creates or updates one stable memory by `kind` and normalized `title`, then rebuilds the SQLite index from markdown.
- `reindex` reparses all markdown blocks and rebuilds the SQLite index.

## Confidence decay

Stored `confidence` remains the canonical base value in markdown. Search and context compilation compute effective confidence at read/rank time, so decay never rewrites memory files by itself.
