# AGENTS

[中文](AGENTS.zh-CN.md)

This file is the execution guide for AI coding agents working in this repository. User-facing progress and handoff updates should default to zh-CN unless the user asks otherwise. Code, command names, test names, upstream docs, and quoted source text stay as-is.

## 1) Project Mission

`memforge` is a local-first project memory layer for AI coding agents (Codex, Claude Code, Cursor, Gemini CLI).
It persists structured project memories on the user's machine, indexes them with SQLite + FTS5, and compiles them into agent-ready context within a token budget.
Although the CLI is human-usable, product and engineering decisions must prioritize reliable invocation by AI agents and automation.

Current codebase shape:

- Go 1.26.3+
- Single-binary CLI: `cmd/memforge`
- CLI commands: `internal/cli`
- Project detection: `internal/project`
- Memory model and markdown I/O: `internal/memory`
- SQLite + FTS5 index: `internal/index`
- Context compiler: `internal/compiler`
- Config (Viper): `internal/config`
- Build metadata: `internal/buildinfo`
- Harness sensors: `harness`

Iteration notes:

- Versioned iteration handoff notes live in `docs/iteration-notes/` with zh-CN mirrors in `docs/zh-CN/iteration-notes/`.
- Before continuing any milestone in plan.md, read the corresponding iteration note when one exists.

Out of scope unless explicitly requested:

- Network upload, remote sync, telemetry, or cloud reporting
- Auto-scanning, hashing, or fingerprinting the user's source files
- LLM provider calls during `init`, `remember`, `search`, `context`, `before`, or `reindex`
- Persistent background services or daemons
- Embedding-based search (planned for v0.4, opt-in)

## 2) Dev Commands

- Format and static checks: `make check`
- Enable local git hooks: `make setup`
- Full tests: `make test`
- Targeted package tests: `make test-packages PKGS="./internal/index ./internal/compiler"`
- Harness-only tests: `make test-harness`
- Vet: `make vet`
- Build: `make build`
- Run CLI with project-local Go cache: `make run ARGS="search auth"`
- Full local validation: `make validate`
- PR metadata check: `make validate-pr-body`
- Commit message check: `make commitlint COMMIT_MSG_FILE=<commit-msg-file>`
- Avoid raw `go test`, `go vet`, `go build`, or `go run` in sandboxed agent sessions; use the Makefile targets so Go build/cache output stays under `.cache/memforge`. For targeted package tests, use `make test-packages PKGS="./internal/index ./internal/compiler"` instead of raw `go test ./internal/index ./internal/compiler`.

## 3) Iteration Self-Constraint Protocol

Before every product or harness iteration, do a short internal review:

1. Requirement classification
- Classify the request as feature, bugfix, refactor, harness/tooling, or analysis-only.
- State the user-visible outcome, target platform, and affected area.
- State the release decision: engineering/process-only optimization does not require a software release; feature and bugfix work must prompt for or continue into the release flow unless the user explicitly defers it.
- If the request is ambiguous, write the safest concrete assumption before editing. Ask only when a wrong assumption would create meaningful data or product risk.

2. Stack fit
- Prefer the Go standard library and existing package boundaries.
- New dependencies need a concrete reason and must explain impact on binary size, offline behavior, and supply-chain risk.
- Local memory storage must stay markdown-canonical. SQLite is an index, not a source of truth.

3. Acceptance and regression plan
- Write observable acceptance criteria before implementation.
- Cover at least one failure or edge case: missing storage directory, malformed markdown frontmatter, duplicate ULID, oversized content, FTS trigger replay after reindex, project hash collisions across forks.
- Map each criterion to unit tests, harness tests, CLI smoke, manual evidence, or explicit not-applicable rationale.

4. Scope guard
- Name the files or directories expected to change.
- Avoid unrelated refactors, formatting churn, dependency upgrades, or release workflow changes.
- Keep harness, CI, docs, and production code responsibilities separated.

5. Handoff readiness
- Run the smallest targeted check while iterating.
- Run the validation matrix in section 6 before handoff based on changed area.
- Summarize residual risks and any validation that could not be run.

## 4) Coding Rules

- Memories live exclusively under `MEMFORGE_HOME` or `$XDG_DATA_HOME/memforge` (fallback `~/.local/share/memforge`). The CLI must never write inside the user's repository.
- Markdown is the source of truth for memories. SQLite is rebuildable via `memforge reindex` and must not carry data that is not in the markdown frontmatter or body.
- Do not add usage-data network transmission. The only command-start network behavior is the low-frequency GitHub release metadata version check, which must remain skippable with `--no-version-check` or `MEMFORGE_NO_VERSION_CHECK=1`.
- LLM provider integration is opt-in. The MVP commands (`init`, `remember`, `search`, `context`, `before`, `reindex`) must never trigger a network LLM call.
- CLI output must stay stable; JSON field changes need tests.
- AI agents should treat `--format json` plus `--no-version-check` as the primary automation contract.
- For JSON commands, stdout must remain a complete machine-readable JSON payload. Human-readable warnings and version prompts belong on stderr or in the returned error path.
- Exit codes are part of the contract: `0` success, `1` user-visible error, `2` invalid invocation, `3` internal error.
- Markdown/table reports should remain readable and scriptable.
- Production packages must not import `harness/` or `tools/`.
- Project hash construction is deterministic and documented in `internal/project/hash.go`; do not change it without bumping the storage schema version.

## 5) Harness Maintenance Rules

- When a repeated issue comes from missing context, update `AGENTS.md` / `AGENTS.zh-CN.md` or docs.
- When an issue can be detected deterministically, add or refine a `harness` test.
- When changing harness, CI, PR workflow, or validation scripts, update:
  - `docs/harness-engineering.md`
  - `docs/zh-CN/harness-engineering.md`
  - `.github/pull_request_template.md` when PR rules change
- Harness tests should check repository structure, workflow constraints, and docs/script alignment. They should not test product business logic.
- Proper noun casing in documentation is enforced by `harness/casing_test.go` using the word list in `docs/casing-rules.txt`. When adding a new proper noun to docs, add its canonical form and wrong variants to the rules file.

## 6) Regression Verification Matrix

- Documentation or harness only: `make check`, `make test-harness`, and `make validate-pr-body` when PR rules changed.
- Memory model or markdown parser: `make test`, including malformed frontmatter, missing fields, duplicate IDs, oversized content.
- Index/SQLite: `make test`, covering migration replay, WAL recovery, FTS trigger fan-out, reindex correctness.
- Compiler/ranker: `make test`, covering empty store, single kind, oversize content, tokenizer fallback path.
- CLI: `make test`, `make build`, and when useful `make run ARGS="search ..."` smoke.
- Dependencies, CI, or release config: `make validate`, with binary-size/offline/supply-chain impact noted.

## 7) CI and Submission Protocol

- CI must run `make validate` and `make test-harness`.
- PRs must include Summary, Requirement Classification, Acceptance Criteria, Changed Areas, Release Decision, TDD / Test Evidence, Validation, and Risk and Rollback.
- Feature and bugfix PRs must mark release required after merge or identify the explicit user-approved deferral. Harness/tooling, docs, CI, and other engineering-process-only PRs should mark release not required.
- Commit messages must match the repository Go commitlint format: `{emoji} {type}{scope}: {subject}`, and the emoji must match the commit type. Current mapping: `✨ feat`, `🐛 fix`, `📝 docs`, `👷 ci`, `💄 style`, `♻️ refactor`, `🔖 release`, `⚡️ perf`, `✅ test`, `🔧 chore`, `🏗️ build`. Run `make setup` once to enable `.githooks/commit-msg`, or run `make commitlint COMMIT_MSG_FILE=<commit-msg-file>` directly. PR CI validates every commit message in the PR range.
- Do not claim completion after skipping local validation.
- Stage/commit only files that belong to the current iteration; do not include unrelated dirty files.

## 8) Open Source Documentation and GitHub Automation

- Every public guide or policy document needs a zh-CN counterpart, for example `README.md` + `README.zh-CN.md` and `CONTRIBUTING.md` + `CONTRIBUTING.zh-CN.md`.
- Keep GitHub PR, review, and automation workflow changes documented in `docs/github-automation.md` and `docs/zh-CN/github-automation.md`.
- Repository review governance is GitHub-native only. Do not introduce paid AI review automation into required repository workflows, docs, or harness rules unless the user explicitly asks for that policy change.
