---
id: FB-509
title: ccs commit-batch needs --worktree flag
type: feature
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-17T14:42:18.985846-03:00"
updated: "2026-04-17T14:42:18.985846-03:00"
suggested_conversion: feature
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

# FB-509: ccs commit-batch needs --worktree flag

## Summary
ccs commit-batch has no --worktree flag. In multi-worktree sessions, committing files in agent worktrees from the main session requires a subshell workaround: (cd /path && ccs commit-batch). The cd-guard hook blocks direct cd, making this non-obvious.

## Repro
Any parallel-agent session where you need to commit from 5+ different worktrees.

## Proposed Solution
Add --worktree <path> flag to ccs commit-batch that chdir's internally before committing. Alternative: auto-detect the repo from the commit plan file paths.

