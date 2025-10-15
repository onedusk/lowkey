# Agent-Managed Code Review System

This proposal defines how repository-aware agents own the entire review loop by wiring into existing Claude hook infrastructure.

## For Agents: Operating Playbook

**You are the primary reviewer.** Follow these directives every time a PR event lands.

1. **Bootstrap Context**
   - Load author and reviewer guides via a new `.claude/hooks/doc_loader.py` hook (mirrors current context system).
   - Cache `docs/eng-practices/review/developer/*.md` and `docs/eng-practices/review/reviewer/*.md` for rubric prompts.
2. **Preflight Checks** (`.claude/hooks/review_preflight.py`)
   - Run `make test`; attach stdout/stderr failures to the PR thread.
   - Enforce CL description rules (imperative first line, "what/why", tag length) before allowing deeper analysis.
3. **Policy Review Loop** (`.claude/hooks/review_cycle.py`)
   - Score the diff against lenses: Design, Functionality, Complexity, Tests, Style, Docs.
   - Emit structured comments with severities `Blocker|Required|Nit`, quoting the relevant guideline section.
   - Auto-apply trivial fixes (formatting, import order) and push patches when possible.
4. **Escalation Rules**
   - After two failed cycles or >1K LOC touched, write a rollup note and ping maintainers.
   - Tag `urgent:` bypasses only the preflight gate; still log findings for audit.
5. **Post-Review Sync** (`.claude/hooks/git_sync.py`)
   - Stage agent-authored patches, ensuring sensitive files remain untouched.
   - Update `.claude/context/current-focus.md` with review status.

## For Maintainers: System Overview

### Hook Topology

- `review_preflight.py` – registered on `PreToolUse` to block non-compliant PRs.
- `review_cycle.py` – tied to `PostToolUse` so every edit triggers re-evaluation.
- `doc_loader.py` – `SessionStart` loader that injects eng-practice excerpts alongside existing context.
- `git_sync.py` – unchanged; continues staging auto-fixes for visibility.

### Rollout Plan

1. **Week 1** – Implement hooks, run in shadow mode on live PRs, store logs under `.claude/context/review-audit/`.
2. **Week 2** – Flip to enforcing mode for standard PRs; schedule a brown-bag to demo agent output.
3. **Week 3** – Add escalation dashboard summarizing agent resolution time and open blockers.

### Success Metrics

- ≥95% of PRs close without human edits.
- Median agent response <15 minutes; 90th percentile escalation <4 hours.
- <5% of human escalations cite unclear agent feedback.

### Risk Mitigation

- **False Positives** – calibrate rubric weights using five historical PRs before enforcement.
- **Knowledge Drift** – monthly prompt review tied to ADR updates.
- **Emergency Fixes** – manual override documented in onboarding; requires maintainer sign-off post-merge.

