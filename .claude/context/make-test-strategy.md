# `make test` Execution Strategy for Agent Review Hooks

## Objectives
- Provide a deterministic pathway for hooks to run `make test` during review without blocking legitimate work indefinitely.
- Respect sandbox and permission constraints while capturing actionable telemetry for reviewers.
- Document rollback procedures so enforcement changes can be disabled quickly if instability occurs.

## Execution Model
- **Trigger**: `review_preflight.py` (or subsequent enforcement layer) invokes a dedicated Bash helper `scripts/run_make_test.sh` when edits affect Go code (`*.go`, `go.mod`, `go.sum`) or test directories.
- **Environment**: Run with `CLAUDE_NO_PROXY=1` and minimal environment to avoid picking up user-specific configs. Use repo root as working directory.
- **Timeouts**: Default 120s timeout; overridable via `.claude/settings.local.json` for local experimentation.
- **Output Capture**: Redirect stdout/stderr to `.claude/context/review-audit/make-test.log`, storing exit code and duration alongside the audit JSON for the corresponding hook event.

## Permission Considerations
- Hooks run under the developer's credentials; they already have filesystem access. No elevation is attempted.
- If sandbox denies execution (write restrictions, missing binaries), log the failure and surface a non-blocking warning to the developer. Enforcement should only occur once successful test execution has been observed in the current session.

## Rollback Plan
1. Toggle enforcement off by removing the `make test` invocation in `review_preflight.py` (keep logging in place) and reload hooks.
2. Clear pending audit files (`.claude/context/review-audit/*`) if they grow unexpectedly large.
3. Communicate via `README.md` / Slack that hooks are in observation mode until issues are resolved.

## Open Questions
- Should we batch test runs across multiple files in a single edit session? (To investigate after initial telemetry.)
- Do we need language-specific fallbacks for non-Go changes? Possibly integrate with existing CI status checks rather than always running `make test`.

