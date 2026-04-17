---
id: FB-508
title: Post-commit hook creates chase-your-tail loop
type: bug
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-17T14:42:12.106127-03:00"
updated: "2026-04-17T14:42:12.106127-03:00"
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

# FB-508: Post-commit hook creates chase-your-tail loop

## Summary
Each ccs commit-batch triggers post-commit hooks that update intel/architecture.json and issues YAML timestamps. Those updates then show as uncommitted changes requiring another commit, which triggers another hook, etc.

## Repro (from cosmo-smoke v0.3 session)
commit → hook updates intel/architecture.json → uncommitted changes → try merge → 'uncommitted changes' error → commit metadata → hook again → repeat.

## Why It Matters
During multi-worktree sessions this manifests as multiple forced rounds of ccs commit-batch just to clear hook-generated noise. It's confusing and wasteful.

## Proposed Solution
Option A — Batch hook-generated updates into a periodic background task rather than per-commit.
Option B — Hook should NOT touch files if the delta is only a timestamp change (no structural change).

