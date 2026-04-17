---
id: FB-002
title: ccs commit-batch needs --worktree flag
type: idea
status: implemented
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: cosmo-smoke
to_target: self
created: "2026-04-16T17:12:43.299638-03:00"
updated: "2026-04-17T14:42:30.779253-03:00"
suggested_conversion: feature
converted_to: null
related_issues: []
brainstorm_ref: null
session: 2026
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

# FB-002: ccs commit-batch needs --worktree flag

During v0.3 parallel-agent session, needed to commit files in 5+ different agent worktrees from the main session. ccs commit-batch has no --worktree flag, so had to use subshell: (cd /path && cat plan | ccs commit-batch). This works but: (a) the cd-guard blocks direct cd, (b) the hack is non-obvious. Suggest: add --worktree <path> flag that chdirs internally. Alternative: detect repo from commit plan file paths. Repro: any multi-worktree merge workflow.

