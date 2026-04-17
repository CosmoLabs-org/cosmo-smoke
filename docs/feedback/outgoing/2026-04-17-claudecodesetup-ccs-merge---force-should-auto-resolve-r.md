---
id: FB-507
title: ccs merge --force should auto-resolve .review.json and .ccsession.json conflicts
type: bug
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-17T14:42:04.972005-03:00"
updated: "2026-04-17T14:42:04.972005-03:00"
suggested_conversion: bug
converted_to: null
related_issues: []
brainstorm_ref: null
session: 2027
suggested_workflow: []
response:
  acknowledged: null
  acknowledged_by: null
  started: null
  implemented: null
  rejected: null
  rejection_reason: null
  notes: ""
---

# FB-507: ccs merge --force should auto-resolve .review.json and .ccsession.json conflicts

## Summary
Every Agent tool worktree merge (isolation:worktree, model:sonnet) hits the same two mechanical failures:
1. .review.json conflict — both master and agent branch have fresh review timestamps
2. 'worktree has uncommitted changes' — .ccsession.json modified post-last-commit

Both resolutions are always identical: git checkout --theirs .review.json, then git checkout .ccsession.json before retrying. Pattern is fully reproducible across 3+ merges in a single session.

## Repro
Run ccs merge <agent-worktree> --force --yes after any Agent tool (isolation:worktree) dispatch.

## Why It Matters
These are mechanical, zero-decision steps that every Claude session following the merge SOP rediscovers. Wasted tool calls every time.

## Proposed Solution
In ccs merge --force, auto-resolve these two well-known files before reporting conflict:
- .review.json: take incoming (reviewed branch wins)
- .ccsession.json: reset to base (session-tracking state, not feature work)
Both decisions are deterministic and should be encoded into the merge flow.

