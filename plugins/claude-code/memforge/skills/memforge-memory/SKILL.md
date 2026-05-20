---
name: memforge-memory
description: Use when you need to search, compile, or automatically maintain local project memories through MemForge.
---

Use the MemForge MCP server tools when available:

- `search_memory` for targeted memory retrieval during coding.
- `compile_context` or `get_project_context` before starting a task.
- `upsert_project_memory` during or near the end of a thread when the current work reveals a stable project decision, constraint, convention, bug root cause, API contract, or agent instruction worth remembering.
- `save_memory` only when you intentionally need a new standalone memory and have checked for duplicates first.

When this plugin is enabled, you may decide whether to create or update project memories without asking for separate confirmation. Keep writes high-signal: do not persist temporary status, raw logs, secrets, credentials, API keys, personal data, guesses, or one-off command output.

Memories are local-first. Do not auto-scan repositories, upload data, or call external LLM providers.
