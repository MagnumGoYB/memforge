---
name: memforge-memory
description: Use when you need local project memory search, context compilation, or automatic project memory maintenance through MemForge.
---

Use the MemForge MCP server tools when available:

- `search_memory` for targeted retrieval.
- `compile_context` or `get_project_context` before coding.
- `upsert_project_memory` during or near the end of a thread when the current work reveals a stable project decision, constraint, convention, bug root cause, API contract, or agent instruction worth remembering.
- `save_memory` only when you intentionally need a new standalone memory and have checked for duplicates first.

Always include `project_root` with the absolute path of the user's current workspace/repository when calling MemForge MCP tools. Codex may keep plugin MCP servers running from the plugin cache directory across turns, so relying on the server process working directory can attach memories to the wrong project.

When this plugin is enabled, you may decide whether to create or update project memories without asking for separate confirmation. Keep writes high-signal: do not persist temporary status, raw logs, secrets, credentials, API keys, personal data, guesses, or one-off command output.

Do not auto-scan repositories, upload data, or call external LLM providers.
