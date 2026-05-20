# MemForge Coding Plan

[中文](plan.zh-CN.md)

## 1. Project Goal

MemForge is a local-first project memory layer for AI coding agents.

Primary goals:

- Provide long-term project memory for Codex / Claude Code / Cursor / Gemini CLI
- Avoid polluting user repositories
- Support manual memory persistence and retrieval
- Compile relevant context for AI coding workflows
- Keep token cost low and retrieval precision high
- Remain fully local and human-controllable

---

## 2. Core Product Positioning

MemForge is NOT:

- A generic knowledge base
- A chat history archive
- A personal AI memory system
- A general-purpose RAG platform

MemForge IS:

> A local project context compiler for AI coding agents.

---

## 3. Technical Stack

```txt
Language: Go 1.26.3+
CLI: github.com/spf13/cobra
Config: github.com/spf13/viper (file + env + flags)
Database: modernc.org/sqlite (pure-Go SQLite + FTS5)
Storage: Markdown (source of truth) + SQLite (index)
ID: github.com/oklog/ulid/v2
Tokenizer: github.com/pkoukk/tiktoken-go (cl100k_base; offline fallback: chars/4)
MCP: stdio MCP Server (v0.3)
LLM Providers: pluggable; OpenAI-compatible / Anthropic / Ollama (v0.2+)
```

Dependency policy:

- Production code must justify every new dependency in the PR. Binary size, offline behavior, and supply-chain surface are part of the review.
- Tools, harness, and CI must depend on the Go standard library only unless a concrete reason is recorded.

---

## 4. AI Engineering Governance

This project ships with the same governance shape as `aitok`, denoised for memforge semantics. Agents must follow the protocol in `AGENTS.md` and the harness/PR rules in `docs/harness-engineering.md`.

Privacy and scope guardrails:

- No network upload, no remote sync, no telemetry by default. The only command-start network behavior is the low-frequency GitHub release metadata version check, which must remain skippable via `--no-version-check` or `MEMFORGE_NO_VERSION_CHECK=1`.
- Memories live exclusively in the user's local data directory. The CLI must not write into the working repository, and must not auto-scan, hash, or fingerprint repository file contents.
- LLM provider calls are opt-in and never run as a side effect of `init`, `remember`, `search`, `context`, or `before`. Memory payloads are sent to a provider only when the user invokes a command that explicitly requires it.
- Production packages must not import `harness/` or `tools/`.

Iteration self-constraint protocol (mirrors aitok §3, denoised):

1. Requirement classification — feature / bugfix / refactor / harness-tooling / analysis-only; state user-visible outcome, target platform, release decision.
2. Stack fit — prefer the Go standard library and existing package boundaries; new dependencies need a concrete reason and an impact statement.
3. Acceptance and regression plan — observable criteria before implementation; at least one failure or edge case covered.
4. Scope guard — name the files or directories expected to change; no unrelated refactors or formatting churn.
5. Handoff readiness — smallest targeted check while iterating; the regression matrix in §16 before handoff.

Release decision policy: engineering- or process-only changes do not require a software release; feature and bugfix changes prompt or continue into the release flow unless the user explicitly defers.

---

## 5. MVP Scope (v0.1)

The first version only focuses on the core workflow:

```txt
init
remember
search
context
before
reindex
```

Not included in v0.1:

```txt
Embedding search
Complex UI
Automatic repository scanning
Automatic memory persistence
Git diff summarization
Session auto extraction
MCP server
LLM provider calls
```

---

## 6. Project Structure

```txt
memforge/
  cmd/
    memforge/
      main.go              # CLI entrypoint, wires Cobra commands
  internal/
    cli/                   # one file per command: init/remember/search/context/before/reindex
    project/               # detect.go, hash.go, meta.go
    memory/                # model.go, markdown.go, store.go, validator.go
    index/                 # sqlite.go, migrations.go, fts.go, search.go
    compiler/              # compiler.go, ranker.go, budget.go, tokenizer.go
    config/                # config.go (Viper wrapper, precedence)
    buildinfo/             # version metadata wired via -ldflags
  tools/
    commitlint/            # Go-native commit message validator
    validate-pr-body/      # Go-native PR template validator
    version/               # reads VERSION, validates tags
  harness/
    architecture_test.go   # structure/contract sensors only
  docs/
    harness-engineering.md
    github-automation.md
    iteration-notes/
    zh-CN/
  .githooks/
    commit-msg
  .github/
    pull_request_template.md
    workflows/
  AGENTS.md, CLAUDE.md, README.md, CONTRIBUTING.md   (+ zh-CN mirrors)
  Makefile, VERSION, go.mod, .gitignore
```

---

## 7. Local Storage Layout

Default storage location resolution order:

```txt
1. MEMFORGE_HOME (env, absolute path)
2. $XDG_DATA_HOME/memforge
3. ~/.local/share/memforge   (macOS and Linux fallback)
```

Layout under that root:

```txt
projects/
  {project_hash}/
    meta.json
    schema_version
    memories/
      manual.md
      constraints.md
      conventions.md
      decisions.md
      bugfixes.md
      api-contracts.md
      agent-instructions.md
    index.sqlite
    cache/
```

Project hash is deterministic:

```txt
1. Try `git config --get remote.origin.url`. If present, canonicalize (lowercase scheme + host, strip credentials, strip trailing `.git`).
2. Else use the git toplevel absolute path.
3. Else use the canonical absolute path of the current working directory.
4. SHA256 the chosen identifier; encode as lowercase hex; keep the first 16 hex chars.
```

Optional project-level config: `.memoryrc` (TOML). The file is optional. Project config never contains memory payloads; it only contains agent-facing knobs (budgets, kind weights, default tags).

Config precedence: `flag > env (MEMFORGE_*) > project .memoryrc > user $XDG_CONFIG_HOME/memforge/config.toml > defaults`.

SQLite operational rules:

- Open with `PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; PRAGMA busy_timeout=5000`.
- Schema version is recorded in `schema_version` (single-row TEXT). Migrations run on every CLI start; `init` and `reindex` must remain safe to re-run.

---

## 8. Memory Model

```go
type MemoryKind string

const (
  KindManual           MemoryKind = "manual"
  KindConstraint       MemoryKind = "constraint"
  KindConvention       MemoryKind = "convention"
  KindDecision         MemoryKind = "decision"
  KindBugfix           MemoryKind = "bugfix"
  KindAPIContract      MemoryKind = "api-contract"
  KindAgentInstruction MemoryKind = "agent-instruction"
)
```

Kind semantics (must be documented in `docs/memory-format.md`):

- `manual` — high-priority free-form notes from the user.
- `constraint` — invariants the agent must obey ("X must remain framework-agnostic").
- `convention` — repository style rules.
- `decision` — recorded architectural decisions.
- `bugfix` — lessons learned from a specific incident.
- `api-contract` — stable interface or schema commitments.
- `agent-instruction` — operating instructions for AI agents; lower priority than `manual` and used for workflow hints.

Memory struct:

```go
type Memory struct {
  ID         string     // ULID, time-sortable, 26 chars
  ProjectID  string
  Kind       MemoryKind
  Title      string
  Content    string
  Tags       []string
  Source     string     // free-text, e.g. "cli", "session:2026-05-19", "import"
  Confidence float64    // 0.0–1.0
  UsageCount int
  CreatedAt  time.Time
  UpdatedAt  time.Time
}
```

Markdown file format (one memory per fenced block, kind-grouped file):

```md
<!-- memforge:memory id=01HQR... kind=decision -->
---
title: Repository layer must remain framework-agnostic
tags: [architecture, repository]
source: cli
confidence: 1.0
created_at: 2026-05-19T10:00:00Z
updated_at: 2026-05-19T10:00:00Z
---

Body in markdown.

<!-- /memforge:memory -->
```

Source-of-truth rule: markdown is canonical. SQLite is rebuildable via `memforge reindex` and must never carry information that is not in the markdown frontmatter or body.

Write atomicity: `remember` writes the markdown block first (fsync the directory), then upserts the SQLite row in the same transaction as the FTS trigger fan-out. If the SQLite step fails, the next CLI invocation reconciles by re-reading the markdown.

---

## 9. SQLite Schema

```sql
CREATE TABLE schema_version (version TEXT NOT NULL);

CREATE TABLE memories (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  tags_json TEXT NOT NULL DEFAULT '[]',
  tags_flat TEXT NOT NULL DEFAULT '',   -- space-separated mirror of tags_json for FTS
  source TEXT,
  confidence REAL NOT NULL DEFAULT 1.0,
  usage_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX memories_kind_idx ON memories(kind);
CREATE INDEX memories_updated_idx ON memories(updated_at);

CREATE VIRTUAL TABLE memories_fts USING fts5(
  title,
  content,
  tags_flat,
  content='memories',
  content_rowid='rowid',
  tokenize = 'unicode61 remove_diacritics 2'
);

CREATE TRIGGER memories_ai AFTER INSERT ON memories BEGIN
  INSERT INTO memories_fts(rowid, title, content, tags_flat)
  VALUES (new.rowid, new.title, new.content, new.tags_flat);
END;

CREATE TRIGGER memories_ad AFTER DELETE ON memories BEGIN
  INSERT INTO memories_fts(memories_fts, rowid, title, content, tags_flat)
  VALUES ('delete', old.rowid, old.title, old.content, old.tags_flat);
END;

CREATE TRIGGER memories_au AFTER UPDATE ON memories BEGIN
  INSERT INTO memories_fts(memories_fts, rowid, title, content, tags_flat)
  VALUES ('delete', old.rowid, old.title, old.content, old.tags_flat);
  INSERT INTO memories_fts(rowid, title, content, tags_flat)
  VALUES (new.rowid, new.title, new.content, new.tags_flat);
END;
```

Tags are persisted twice: `tags_json` for round-tripping and `tags_flat` as a tokenizer-friendly mirror that participates in FTS5 scoring. `tags_flat` is rebuilt on every upsert from `tags_json`.

---

## 10. CLI Commands

### Common flags

Every command supports:

- `--format text|json` (default `text`; `json` is the automation contract).
- `--no-version-check` and respects `MEMFORGE_NO_VERSION_CHECK=1`.
- Non-zero exit codes are part of the contract:
  - `0` success
  - `1` user-visible error (missing project, validation failed)
  - `2` invalid invocation (bad flags)
  - `3` internal error (filesystem, SQLite)

JSON payload on stdout must remain a complete machine-readable JSON object. Human-readable warnings and version prompts belong on stderr.

### init

```bash
memforge init [--root PATH]
```

Responsibilities:

```txt
Detect current project (git root or cwd)
Compute project hash (§7)
Create storage directory and meta.json
Create markdown memory files (kind-grouped, empty)
Open SQLite, apply migrations, initialize FTS5
```

Idempotent: re-running `init` must not destroy existing memories. It only creates missing files and migrates schema.

### remember

```bash
memforge remember \
  --kind decision \
  --title "Repository layer is framework-agnostic" \
  --tag architecture --tag repository \
  --source cli \
  --confidence 1.0 \
  "Body text or - to read from stdin or --from FILE to read from a file"
```

Content sources (mutually exclusive):

```txt
positional argument
--from FILE
-          (read from stdin)
```

Responsibilities:

```txt
Validate kind, tags, content
Generate ULID
Append markdown block to the matching kind file
Upsert SQLite row inside a single transaction
Print JSON {id, kind, title} on --format json
```

### search

```bash
memforge search [--kind k1,k2] [--tag t1] [--limit 20] [--format json] QUERY
```

Behavior:

```txt
FTS5 MATCH with prefix tokens
Apply ranker (§11)
Return Title, Kind, Tags, Snippet, Score
```

### context

```bash
memforge context --budget 3000 [--format json] [--kinds k1,k2]
```

Responsibilities:

```txt
Load all memories of selected kinds (default: all)
Rank inside each kind (§11)
Apply token budget across kinds proportionally to kind weights
Generate agent-ready markdown context grouped by kind
```

### before

```bash
memforge before [--budget 3000] [--format json] "TASK DESCRIPTION"
```

`before` is task-conditioned. It differs from `context` in that:

```txt
1. Search relevant memories using TASK DESCRIPTION as the query
2. Merge task-matched memories with high-priority memories
   (manual + constraint always included up to a sub-budget)
3. Emit a markdown payload that names the task as the first heading
```

### reindex

```bash
memforge reindex
```

Rebuilds `index.sqlite` from markdown. Used for recovery after a failed write or after manual markdown edits.

---

## 11. Context Compiler

Objective:

> Generate the highest-value project context within a limited token budget.

Token counting: tiktoken `cl100k_base`. If the tokenizer module is unavailable, fall back to `len(utf8.RuneCountInString(content))/4 + 8` per memory and emit a stderr warning.

Ranking formula (normalized so individual terms cannot dominate):

```txt
score =
    0.40 * bm25_norm        // FTS5 BM25, min-max normalized within the candidate set
  + 0.20 * kind_weight_norm // kind weight / 100
  + 0.15 * recency_score    // 0.5 ^ (days_since_updated / 30)
  + 0.15 * usage_score      // min(1, log10(1 + usage_count) / 2)
  + 0.10 * confidence       // 0.0–1.0
```

Default kind weights (configurable in `.memoryrc`):

```txt
manual:            100
constraint:         90
convention:         80
decision:           70
bugfix:             60
api-contract:       55
agent-instruction:  50
```

Budget allocation:

```txt
1. Compute per-kind quota = (kind_weight / sum_kind_weights) * total_budget
2. Fill quotas greedily by descending score
3. Redistribute unused quota to the next-highest-weight kind
4. Always include manual + constraint up to 30% of the budget before greedy fill
```

---

## 12. Development Phases

Every phase must follow the iteration self-constraint protocol (§4) and the regression verification matrix (§16). Tests are required before code unless the phase is documentation-only.

### Phase 1 — Project Bootstrap

Tasks:

- Initialize Go module
- Wire Cobra root command, version sub-command via `internal/buildinfo`
- Wire Viper config with the precedence in §7
- Activate the Makefile, .githooks, and harness sensors that ship in this iteration

Acceptance:

```bash
make build
./memforge --version
./memforge help
make test-harness
```

### Phase 2 — Project Detection and Storage Bootstrap

Tasks:

- Implement `internal/project` (detect, hash, meta)
- Implement `internal/index/sqlite.go` (migrations, WAL, busy_timeout)
- Implement `internal/memory/markdown.go` (frontmatter parser/writer)
- `memforge init` writes `meta.json`, kind-grouped markdown files, `index.sqlite`

Acceptance:

```bash
memforge init
test -f "$(memforge debug paths --format json | jq -r .meta)"
```

### Phase 3 — remember

Tasks:

- Implement `--kind`, `--title`, `--tag`, `--source`, `--confidence`, `--from`, stdin
- Markdown write → SQLite upsert atomicity
- JSON output contract

Acceptance:

```bash
memforge remember --kind decision --title "..." "Body"
memforge remember --kind constraint --from notes.md
echo "Body via stdin" | memforge remember --kind manual --title "Pipe" -
```

### Phase 4 — search

Tasks:

- FTS5 MATCH builder with prefix tokens
- Ranking and snippet generation
- JSON and text formats

Acceptance:

```bash
memforge search "repository framework"
memforge search --format json --kind constraint,decision "auth"
```

### Phase 5 — context

Tasks:

- Tokenizer wiring (tiktoken with chars/4 fallback)
- Budget allocator across kinds
- Markdown context emitter grouped by kind

Acceptance:

```bash
memforge context --budget 3000
memforge context --budget 1500 --kinds constraint,decision --format json
```

### Phase 6 — before

Tasks:

- Task-conditioned retrieval
- Merge high-priority (manual + constraint) sub-budget with task-matched results
- Markdown payload with the task as the leading section

Acceptance:

```bash
memforge before "Refactor auth middleware"
```

### Phase 7 — reindex and recovery

Tasks:

- Parse all markdown blocks and rebuild SQLite from scratch
- Detect and report orphans (rows without a backing markdown block) and ghosts (markdown blocks without a row)

Acceptance:

```bash
rm "$STORAGE/projects/$HASH/index.sqlite"
memforge reindex
memforge search "repository framework"   # still works
```

---

## 13. v0.2 Roadmap

```txt
after command (candidate memory extraction from a session log)
LLM provider abstraction (OpenAI-compatible, Anthropic, Ollama)
Memory deduplication and merge proposals
Confidence decay over time
```

Example:

```bash
memforge after --from session.md
```

Default behavior:

```txt
Generate candidate memories
Require human confirmation
Persist approved memories
```

LLM calls are opt-in. They never run during `init`, `remember`, `search`, `context`, or `before`.

---

## 14. v0.3 Roadmap

Add MCP server support (stdio).

Tools (each must publish a JSON schema in `docs/mcp.md`):

```txt
search_memory       { query: string, kinds?: string[], limit?: int, hybrid?: bool }
compile_context     { budget?: int, kinds?: string[] }
save_memory         { kind: string, title: string, content: string, tags?: string[] }
upsert_project_memory { kind: string, title: string, content: string, tags?: string[] }
list_constraints    { limit?: int }
get_project_context { task?: string, budget?: int }
```

Agent workflow:

```txt
Task Start    -> compile_context | get_project_context
During Coding -> search_memory
After Coding  -> save_memory for explicit new memories, or upsert_project_memory for stable agent-selected memory maintenance
```

The MCP server must reuse the same CLI guardrails: no remote network, no auto-scan, JSON-only on stdout. Enabled plugins may let the agent decide to call `upsert_project_memory` during a thread, while CLI `after` remains proposal-first and human-approved.

---

## 15. v0.4 Roadmap

```txt
Git diff summarization
Session extraction adapters (Codex, Claude Code, Cursor)
Hybrid search (BM25 + embeddings)
Embedding providers (local-first; opt-in remote)
Context budget optimizer
Claude Code integration docs
Codex integration docs
```

---

## 16. Regression Verification Matrix

- Documentation or harness only: `make check`, `make test-harness`, `make validate-pr-body` when PR rules changed.
- Memory model or markdown parser: `make test`, including malformed frontmatter, missing fields, duplicate IDs, oversized content.
- Index/SQLite: `make test`, covering migration replay, WAL recovery, FTS trigger fan-out, reindex correctness.
- Compiler/ranker: `make test`, covering empty store, single kind, oversize content, tokenizer fallback path.
- CLI: `make test`, `make build`, and `make run ARGS="search ..."` smoke where useful.
- Dependencies, CI, or release config: `make validate`, with binary-size/offline/supply-chain impact noted.

---

## 17. CI and Submission Protocol

- CI runs `make validate` and `make test-harness`.
- PRs must use `.github/pull_request_template.md` and include Summary, Requirement Classification, Acceptance Criteria, Changed Areas, Release Decision, TDD / Test Evidence, Validation, and Risk and Rollback.
- Feature and bugfix PRs mark release required after merge or identify the explicit user-approved deferral. Harness/tooling, docs, and CI changes mark release not required.
- Commit messages follow the Go commitlint format: `{emoji} {type}{scope}: {subject}`. Mapping: `✨ feat`, `🐛 fix`, `📝 docs`, `👷 ci`, `💄 style`, `♻️ refactor`, `🔖 release`, `⚡️ perf`, `✅ test`, `🔧 chore`, `🏗️ build`. Run `make setup` once to wire `.githooks/commit-msg`, or call `make commitlint COMMIT_MSG_FILE=<commit-msg-file>` directly. PR CI validates every commit message in the PR range.
- Do not claim completion after skipping local validation.
- Stage only files that belong to the current iteration; do not include unrelated dirty files.

---

## 18. Development Priority

```txt
1. Phase 1 — bootstrap and harness
2. Phase 2 — project detection and storage
3. Phase 3 — remember
4. Phase 4 — search
5. Phase 5 — context
6. Phase 6 — before
7. Phase 7 — reindex
8. MCP (v0.3)
9. LLM summarization (v0.2)
10. Embedding search (v0.4)
```

---

## 19. MVP Success Criteria

The following commands work reliably end-to-end:

```bash
memforge init
memforge remember --kind decision --title "Repository layer is framework-agnostic" "Body"
memforge search "repository framework"
memforge before "Refactor repository layer"
memforge reindex
```

And the system can:

```txt
Persist memories reliably (markdown is canonical, SQLite is rebuildable)
Retrieve relevant memories accurately under the §11 ranker
Avoid polluting repositories (storage stays under MEMFORGE_HOME / XDG_DATA_HOME)
Generate useful agent-ready context within a token budget
Recover from a deleted index.sqlite via reindex
```

---

## 20. Core Principles

```txt
Prioritize correctness over complexity
Prioritize local-first over cloud-first
Prioritize human control over full automation
Prioritize engineering memory over chat memory
Prioritize CLI over UI
Prioritize markdown source of truth over derived indexes
```
