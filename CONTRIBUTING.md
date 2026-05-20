# Contributing

[中文](CONTRIBUTING.zh-CN.md)

Thanks for contributing to `memforge`. This repository is a local-first Go CLI and memory layer for AI coding agents, so every contribution must preserve offline/privacy boundaries, stable automation behavior, and repository-non-pollution guarantees.

## Required Checks

Run before submitting changes:

```bash
make check
make test
make test-harness
make build
```

When PR metadata, CI, workflows, or repository guardrails change, also run:

```bash
make validate-pr-body
```

For broader changes, prefer:

```bash
make validate
```

## Contribution Constraints

- Every behavior change needs tests, harness coverage, or explicit manual evidence.
- Parser and markdown changes must include malformed or missing input cases where relevant.
- Keep the tool offline by default.
- Do not add upload, sync, telemetry, or background services unless explicitly requested.
- Do not write memories inside the user's repository.
- Markdown remains the source of truth; SQLite must stay rebuildable via `memforge reindex`.
- CLI automation behavior must stay stable, especially JSON output, stderr/stdout separation, and exit codes.

## Repository Workflow

- Use `make setup` once to enable the repository `commit-msg` hook.
- Commit messages must follow the repository format validated by `make commitlint COMMIT_MSG_FILE=<commit-msg-file>`.
- Pull requests must fill the required sections in `.github/pull_request_template.md`.
- Feature and bugfix changes must mark release required after merge or explicitly document user-approved deferral.

## Updating Governance

When changing harness, CI, PR workflows, or validation scripts, update the matching governance docs:

- `AGENTS.md` / `AGENTS.zh-CN.md` when agent instructions change
- `docs/harness-engineering.md` / `docs/zh-CN/harness-engineering.md`
- `docs/github-automation.md` / `docs/zh-CN/github-automation.md`
- `.github/pull_request_template.md` / `.github/pull_request_template.zh-CN.md` when PR metadata rules change
