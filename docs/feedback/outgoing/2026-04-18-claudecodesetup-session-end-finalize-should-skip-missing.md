---
id: FB-537
title: session-end-finalize should skip missing release-context dir
type: feature
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-18T22:40:04.889135-03:00"
updated: "2026-04-18T22:40:04.889135-03:00"
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

# FB-537: session-end-finalize should skip missing release-context dir

## Summary
`ccs session-end-finalize` fails with: `fatal: pathspec '.claude/release-context/' did not match any files`. This directory doesn't exist in projects that aren't ClaudeCodeSetup.

## Motivation
Blocks session-end Phase 6 in every non-ClaudeCodeSetup project. The failure is noisy even though the rest of finalize partially completes.

## Proposed Solution
Add existence check before `git add` for release-context directory. If missing, skip silently.

## Reproduction
```bash
ccs session-end-finalize
# stage-artifacts step fails with pathspec error
```

## Affected Files
`tools/ccsession/internal/sessionend/finalize.go` (or wherever stage-artifacts lives)

## Priority: medium (session-end still partially completes, but the failure is noisy)

