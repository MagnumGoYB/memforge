# Harness Engineering

[中文](zh-CN/harness-engineering.md)

A harness combines feedforward guides that steer an agent before it edits with feedback sensors that catch drift after it edits.

For `memforge`, the harness is intentionally lightweight: Go tests, a Makefile, PR metadata validation, CI gates, and concise agent guides. It protects the offline, repository-non-polluting memory contract without adding background services.

## Feedforward Guides

- `AGENTS.md` and `AGENTS.zh-CN.md`: repository mission, coding constraints, validation matrix, privacy boundaries, and handoff rules.
- `README.md`: user-facing CLI usage and install path.
- `CONTRIBUTING.md`: contributor validation and offline-first rules.
- `Makefile`: canonical local commands: `make setup`, `make check`, `make test`, `make test-packages`, `make test-harness`, `make vet`, `make build`, `make validate`, `make validate-pr-body`, and `make commitlint`.
- `tools/commitlint` and `.githooks/commit-msg`: repository-native Go commit-message validation for `{emoji} {type}{scope}: {subject}` without Node/npm tooling. Each commit type has exactly one allowed emoji: `✨ feat`, `🐛 fix`, `📝 docs`, `👷 ci`, `💄 style`, `♻️ refactor`, `🔖 release`, `⚡️ perf`, `✅ test`, `🔧 chore`, and `🏗️ build`. `make setup` enables the hook for local commits.
- `.github/pull_request_template.md`: repeatable PR checklist for requirement classification, acceptance criteria, test evidence, validation, rollback, and residual risk.
- Release decision policy: engineering/process-only changes do not require a software release; feature and bugfix changes require a follow-up release or explicit deferral.
- `.github/workflows/ci.yml`: hosted validation matching the local gates.

## Feedback Sensors

- `make test`: unit and integration coverage for project detection, memory model, markdown I/O, SQLite/FTS, compiler, and CLI command surfaces.
- `make test-packages PKGS="./internal/index ./internal/compiler"`: targeted package tests with the same repository-local Go caches as the full test target.
- `make test-harness`: repository-structure sensors for agent docs, Makefile commands, CI gates, PR template, offline/privacy constraints, and proper noun casing in documentation (`docs/casing-rules.txt`).
- `make vet`: static analysis.
- `make build`: single-binary build check.
- `make validate-pr-body`: executable PR body metadata gate.
- `make setup`: one-time local setup that runs `git config core.hooksPath .githooks`.
- `make commitlint COMMIT_MSG_FILE=<commit-msg-file>`: executable single commit-message gate, wired through `.githooks/commit-msg` after setup.
- `make commitlint-range COMMIT_RANGE=<base..head>`: executable PR-range commit-message gate, mirrored by PR CI for every commit in the PR.
- `.cache/memforge/`: repository-local, git-ignored Go build/module cache used by Makefile targets so agent verification stays bound to this checkout instead of ad hoc `/tmp` paths.
- Agents should not run raw `go test`, `go vet`, `go build`, or `go run` in sandboxed sessions. Use Makefile targets so Go never falls back to `~/Library/Caches/go-build`.

## Agent Workflow Contract

- Classify each request before editing: feature, bugfix, refactor, harness/tooling, or analysis-only.
- Lock observable acceptance criteria before implementation.
- Record the release decision before handoff.
- Add a failing test, harness sensor, or explicit manual verification checklist before changing behavior when practical.
- Keep edits scoped to the named files/directories.
- Map every acceptance criterion to evidence before handoff.
- Report skipped validation and residual risk.

## memforge-Specific Guardrails

- Do not add network upload, sync, telemetry, or remote reporting by default. The only command-start network call permitted is the GitHub release version check, which must remain skippable via `--no-version-check` or `MEMFORGE_NO_VERSION_CHECK=1`.
- Do not write into the user's repository. Memories live under `MEMFORGE_HOME` or `$XDG_DATA_HOME/memforge`.
- Do not auto-scan, hash, or fingerprint user source files.
- Markdown is the source of truth; SQLite must remain rebuildable via `memforge reindex`.
- LLM provider calls are opt-in. MVP commands (`init`, `remember`, `search`, `context`, `before`, `reindex`) must never trigger a network LLM call.
- Preserve stable CLI output for automation: `--format json` is the contract.

## Updating Harness

When changing harness, CI, PR workflow, or validation scripts:

1. Update the executable script/test/workflow.
2. Update `AGENTS.md` and `AGENTS.zh-CN.md` if agent behavior changes.
3. Update this document and `docs/zh-CN/harness-engineering.md`.
4. Run `make check`, `make test-harness`, and `make validate-pr-body` when PR metadata rules changed.

When changing commit workflow rules, update `tools/commitlint`, `.githooks/commit-msg`, `Makefile`, `.github/workflows/pr.yml`, `AGENTS.md`, `AGENTS.zh-CN.md`, this document, and `docs/zh-CN/harness-engineering.md`.

When changing release decision rules, update `tools/validate-pr-body`, `.github/pull_request_template.md`, `.github/pull_request_template.zh-CN.md`, `.github/workflows/pr-review.yml`, `AGENTS.md`, `AGENTS.zh-CN.md`, this document, `docs/zh-CN/harness-engineering.md`, and `docs/github-automation.md` / `docs/zh-CN/github-automation.md`.

## GitHub Automation Coverage

- `docs/github-automation.md` and `docs/zh-CN/github-automation.md` document the PR review checklist, Dependabot auto-merge, and CI flows.
- `.github/workflows/pr.yml` validates real PR metadata and every commit message in the PR.
- `.github/workflows/pr-review.yml` posts the review checklist with `issues: write` and `pull-requests: write` permissions so it can create or update the PR issue comment.
- `.github/CODEOWNERS` routes high-risk areas to maintainers; required repository review governance stays GitHub-native and excludes paid AI review automation.
- `.github/workflows/ci.yml` runs `make validate` and `make test-harness` on push and pull request.
