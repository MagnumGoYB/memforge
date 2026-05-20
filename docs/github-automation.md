# GitHub Automation

[中文](zh-CN/github-automation.md)

This repository uses GitHub-native automation for pull requests, review prompts, Dependabot auto-merge, bug reports, and CI.

## Pull Request Flow

- `.github/pull_request_template.md` requires requirement classification, acceptance criteria, test evidence, validation, rollback, and residual risk.
- `.github/workflows/pr.yml` validates the real pull request body with `make validate-pr-body` and validates every commit message in the PR range with `make commitlint-range`.
- PR metadata must include a release decision: feature and bugfix changes require a follow-up release or an explicit user-approved deferral, while engineering-process-only changes mark release not required.
- `.github/workflows/ci.yml` runs local validation and harness gates on pushes and pull requests.

## Review Flow

- `.github/workflows/pr-review.yml` posts a checklist comment on new or updated pull requests.
- The checklist workflow runs with `issues: write` and `pull-requests: write` so `actions/github-script` can create or update the PR issue comment under branch protection.
- The checklist reminds reviewers to inspect offline/privacy boundaries, repository-non-pollution, markdown-as-source-of-truth, CLI output stability, release decision, and release impact.
- `.github/CODEOWNERS` requests review for core areas such as `internal/`, `harness/`, `tools/`, and GitHub workflows.
- Repository review governance is GitHub-native only. Do not add paid AI review automation back into required repository workflows.

## Bugfix Flow

- `.github/ISSUE_TEMPLATE/bug_report.yml` captures CLI command, OS, version, expected behavior, actual behavior, sanitized logs, and validation evidence.
- Bugfix PRs must include a failing test, fixture, harness sensor, or explicit manual reproduction before the fix when practical.
- The PR template requires failure or edge coverage and rollback notes.

## Dependabot Flow

- `.github/dependabot.yml` opens weekly grouped PRs for Go modules and GitHub Actions.
- `.github/workflows/dependabot-auto-merge.yml` enables GitHub auto-merge only for non-major Dependabot updates.

## CI Flow

- `.github/workflows/ci.yml` runs `make validate` and `make test-harness` on Ubuntu using the pinned Go toolchain.
- Workflows use Node 24 action majors such as `actions/checkout@v6`, `actions/setup-go@v6`, and `actions/github-script@v8`.

## Release Flow

Release automation is intentionally out of scope for v0.1. When release infrastructure is added, document it here and mirror the changes to `docs/zh-CN/github-automation.md`.

Release decision policy stays in effect even without release automation: feature and bugfix changes prompt for or continue into the release flow unless the user explicitly defers; engineering- or process-only changes mark release not required.
