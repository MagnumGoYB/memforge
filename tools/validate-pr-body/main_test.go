package main

import "testing"

func TestValidateBodyAcceptsCompleteTemplate(t *testing.T) {
	body := `## Summary

- Add init command that creates the project storage directory.

## Requirement Classification

- Type: feature
- User-visible outcome: ` + "`memforge init`" + ` creates the storage layout.
- Target platform(s): CLI on darwin and linux.
- Scope assumptions: project hash uses git remote when available.

## Acceptance Criteria

- [x] Criteria are written before implementation starts.
- [x] Each criterion maps to unit test coverage where practical.
- [x] CLI evidence covers the happy path.
- [x] At least one failure or edge case is covered.

## Changed Areas

- internal/project, internal/cli/init.go, harness/architecture_test.go

## Release Decision

- Release required after merge: feature work, will tag v0.2.0 follow-up release.

## TDD / Test Evidence

- Test/sensor added before implementation: yes, failing test in internal/cli.
- Unit tests: internal/cli/init_test.go covers happy and edge.
- CLI/manual/platform evidence: ran ` + "`make run ARGS=init`" + ` locally.
- Failure and edge cases covered: missing git remote, non-empty existing directory.
- Acceptance criteria evidence map: criterion 1 -> init_test happy path.

## Validation

- [x] make check
- [x] make test
- [x] make test-harness
- [x] make build
- Skipped validation and reason: none skipped.

## Risk and Rollback

- Risk: low; new feature with no existing behavior change.
- Rollback: revert commit; storage layout is additive, offline, local-first.
- Residual risk: none observed.
`
	if errs := validateBody(body); len(errs) != 0 {
		t.Fatalf("expected pass, got: %v", errs)
	}
}

func TestValidateBodyFlagsMissingSections(t *testing.T) {
	if errs := validateBody(""); len(errs) == 0 {
		t.Fatal("expected empty body to fail")
	}
}
