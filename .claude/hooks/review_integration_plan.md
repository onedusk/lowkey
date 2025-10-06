# Agent Review Hook Integration Plan

This note outlines how the proposed agent-managed review flow will hook into the existing Claude automation without disrupting the current Oober and documentation systems.

## Target Hook Events

- **SessionStart** – extend `doc_loader.py` to pull engineering-practices excerpts and expose them as additional context under a new `review_guides` section.
- **PreToolUse (Edit|MultiEdit)** – insert `review_preflight.py` before the Oober hooks so the agent can validate CL metadata, ensure `make test` readiness, and block non-compliant edits.
- **PostToolUse (Edit|MultiEdit|Write)** – run `review_cycle.py` after Oober post-edit and context reminders to score diffs, emit structured findings, and queue auto-fixes before `git_sync.py` stages results.
- **UserPromptSubmit** – optional future hook `review_prompt_policy.py` to enforce PR template fields on manual agent prompts.

## Execution Order Concept

1. `review_preflight.py` (new) – validates and logs; may deny tool call.
2. `oober/pre_edit.py` – keeps bulk-edit suggestion flow.
3. Tool executes (Edit/MultiEdit).
4. `oober/post_edit.py` – suggests oober if patterns match.
5. `review_cycle.py` (new) – analyzes diff/tests, writes findings to audit log.
6. `context_reminder.py`
7. `git_sync.py`

## Required Configuration Changes

- Add new PreToolUse matcher entry for `review_preflight.py` ahead of the existing Oober matcher in `.claude/settings.json`.
- Add new PostToolUse matcher entry for `review_cycle.py` with timeout ≥10s to allow analysis.
- Create `.claude/context/review-audit/` for structured JSON logs keyed by session + file.

## Next Steps

- Implement `review_preflight.py` in logging mode only.
- Define audit log schema (`timestamp`, `tool_name`, `decision`, `messages`).
- Dry-run hooks locally and capture stdout/stderr to confirm they integrate cleanly.
