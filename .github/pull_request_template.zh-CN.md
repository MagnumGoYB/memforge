## Summary

- 简要说明本次变更。
- 说明 offline / local-first 行为是否变化。

## Requirement Classification

- Type: feature | bugfix | refactor | harness/tooling | analysis-only
- User-visible outcome:
- Target platform(s): darwin | linux | windows | repository-only
- Scope assumptions:

## Acceptance Criteria

- [ ] 实现前已写明验收标准。
- [ ] 每条验收标准在可行时映射到单元测试。
- [ ] 每条验收标准在相关时映射到 CLI/manual/platform 证据。
- [ ] 覆盖至少一个失败或边界场景，或说明原因。

## Changed Areas

- 列出本次变更涉及的文件或目录。

## Release Decision

- Release required after merge: feature 和 bugfix 变更应提示或继续进入发版流程，除非用户已明确批准延期。
- Release not required: harness、docs、CI 或 workflow guardrails 等工程/流程优化。
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
