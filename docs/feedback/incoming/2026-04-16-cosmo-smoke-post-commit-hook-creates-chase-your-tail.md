---
id: FB-003
title: Post-commit hook creates chase-your-tail loop
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: cosmo-smoke
to_target: self
created: "2026-04-16T17:12:47.323845-03:00"
updated: "2026-04-16T17:12:47.323845-03:00"
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

# FB-003: Post-commit hook creates chase-your-tail loop

Each ccs commit-batch triggers post-commit hooks that update intel/architecture.json + issues YAML timestamps, which then show as uncommitted changes requiring ANOTHER commit, triggering ANOTHER hook update, etc. During v0.3 session this manifested as: commit → hook → timestamp update → try merge → 'uncommitted changes' error → commit metadata → hook again. Workaround: multiple rounds of ccs commit-batch. Suggest: (a) batch hook-generated updates into a periodic background task rather than per-commit, or (b) hook should NOT touch files if they are only timestamp deltas.

