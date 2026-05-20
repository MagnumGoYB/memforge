## Summary

- Briefly describe the change.
- State whether offline/local-first behavior changed.

## Requirement Classification

- Type: feature | bugfix | refactor | harness/tooling | analysis-only
- User-visible outcome:
- Target platform(s): darwin | linux | windows | repository-only
- Scope assumptions:

## Acceptance Criteria

- [ ] Criteria are written before implementation starts.
- [ ] Each criterion maps to unit test coverage where practical.
- [ ] Each criterion maps to CLI/manual/platform evidence where relevant.
- [ ] At least one failure or edge case is covered, or a reason is documented.

## Changed Areas

- List the files or directories changed.

## Release Decision

- Release required after merge: feature and bugfix changes should prompt for or continue into the release flow, unless the user explicitly approved deferral.
- Release not required: engineering/process-only changes such as harness, docs, CI, or workflow guardrails.
- Decision for this PR:

## TDD / Test Evidence

- Test/sensor added before implementation:
- Unit tests:
- CLI/manual/platform evidence:
- Failure and edge cases covered:
- Acceptance criteria evidence map:

## Validation

- [ ] `make check`
- [ ] `make test`
- [ ] `make test-harness`
- [ ] `make build`
- [ ] Commit message follows `make commitlint COMMIT_MSG_FILE=<commit-msg-file>`
- Skipped validation and reason:

## Risk and Rollback

- Risk:
- Rollback:
- Residual risk:
