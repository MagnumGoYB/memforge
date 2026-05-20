# memforge

[中文](README.zh-CN.md)

![memforge overview](README-200k.webp)

`memforge` is a local-first project memory layer for AI coding agents such as Claude Code, Codex, Cursor, and Gemini CLI.

It stores structured project memories on the user's machine, indexes them with SQLite + FTS5, and compiles agent-ready context within a token budget. Markdown remains the source of truth; the SQLite index is rebuildable.

## What it is

`memforge` is a local project context compiler for AI coding workflows.

It is designed to help agents and developers keep durable project memory without polluting the working repository.

## What it is not

- Not a generic knowledge base
- Not a chat history archive
- Not a cloud memory service
- Not a repository auto-scanner
- Not an always-on background daemon

## Current status

This repository now has a working local-first MVP command path.

Implemented now:

- `version`
- `help`
- `init`
- `remember`
- `search`
- `context`
- `before`
- `after`
- `reindex`
- `mcp`
- `diff-summary`
- hybrid `search --hybrid`
- session adapters for `after --adapter`
- `debug paths`
- governance, harness, and GitHub workflow guardrails

## Installation

### Agent plugin installs

For Claude Code, the normal user path is a marketplace or release plugin install. The plugin package includes platform-specific `memforge` runtime binaries and starts MCP through `bin/memforge-mcp-launcher.js`, so users do not run `go install` first and do not need a separate `memforge` CLI on `PATH`.

Codex packaged plugin bundles include the same runtime, but public self-service publishing is still limited; use local/private marketplace or Codex host plugin package support where available. See `docs/plugin-distribution.md` for the current Claude Code and Codex distribution details.

### CLI and source development installs

Requirements for CLI/source builds:

- Go 1.26.3 or newer

Install the latest published module version:

```bash
go install github.com/MagnumGOYB/memforge/cmd/memforge@latest
```

Install from a local checkout:

```bash
git clone https://github.com/MagnumGOYB/memforge.git
cd memforge
make build
go install ./cmd/memforge
```

For agent and automation usage, set a local storage root explicitly:

```bash
export MEMFORGE_HOME="$HOME/.local/share/memforge"
```

## Quick start

Build the CLI:

```bash
make build
```

Run the current commands:

```bash
make run ARGS="version"
MEMFORGE_HOME=/absolute/path make run ARGS="init --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="remember --kind decision --title 'Repository layer is framework-agnostic' --format json --no-version-check 'Body'"
MEMFORGE_HOME=/absolute/path make run ARGS="search --format json --no-version-check 'repository framework'"
MEMFORGE_HOME=/absolute/path make run ARGS="context --budget 3000 --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="before --budget 3000 --format json --no-version-check 'Refactor repository layer'"
MEMFORGE_HOME=/absolute/path make run ARGS="after --from /absolute/path/session.md --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="after --from /absolute/path/session.md --approve all --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="reindex --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="mcp --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="search --hybrid --format json --no-version-check 'repository framework'"
MEMFORGE_HOME=/absolute/path make run ARGS="after --adapter claude-code --from /absolute/path/session.jsonl --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="diff-summary --from /absolute/path/numstat.txt --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="debug paths --format json --no-version-check"
```

## AI agent contract

For automation and agent-driven usage, prefer:

```bash
memforge --no-version-check <command> --format json
```

Rules:

- JSON payloads belong on stdout.
- Human-readable warnings belong on stderr.
- `--no-version-check` is part of the automation contract.

## Local-first and privacy boundaries

- Memories do not live in the user repository.
- Storage resolves under `MEMFORGE_HOME` or `$XDG_DATA_HOME/memforge`.
- Default behavior is local-first and offline.
- Markdown is canonical storage.
- SQLite is a rebuildable index layer.
- MVP commands do not make opt-in LLM/provider calls.
- `after` is proposal-first: it extracts candidate memories from an explicit session file and persists only when `--approve` is provided.
- Provider-backed extraction is opt-in and scoped to `after`; other commands stay local/offline.
- Hybrid search is explicit via `search --hybrid` and uses local deterministic embeddings by default.
- Session adapters and diff summaries are local transformations over explicitly supplied files or local git output.

## Development

Use the repository Makefile targets so Go caches stay inside `.cache/memforge` during local and agent-driven validation:

```bash
make setup
make check
make test
make test-packages PKGS="./internal/index ./internal/compiler"
make test-harness
make vet
make build
make validate
make validate-pr-body
```

## Open source project docs

- Contributor guide: `CONTRIBUTING.md` / `CONTRIBUTING.zh-CN.md`
- Agent execution guide: `AGENTS.md` / `AGENTS.zh-CN.md`
- Claude Code entrypoint: `CLAUDE.md` / `CLAUDE.zh-CN.md`
- Harness engineering: `docs/harness-engineering.md` / `docs/zh-CN/harness-engineering.md`
- GitHub automation: `docs/github-automation.md` / `docs/zh-CN/github-automation.md`
- Memory format: `docs/memory-format.md` / `docs/zh-CN/memory-format.md`
- MCP server: `docs/mcp.md` / `docs/zh-CN/mcp.md`
- Agent integrations: `docs/integrations.md` / `docs/zh-CN/integrations.md`
- Plugin distribution: `docs/plugin-distribution.md` / `docs/zh-CN/plugin-distribution.md`
- Plan: `plan.md` / `plan.zh-CN.md`

## Module path

```bash
go install github.com/MagnumGOYB/memforge/cmd/memforge@latest
```

For local development from this checkout:

```bash
go install ./cmd/memforge
```
