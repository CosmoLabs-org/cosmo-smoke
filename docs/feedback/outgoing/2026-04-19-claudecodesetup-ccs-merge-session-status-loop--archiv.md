---
id: FB-572
title: ccs merge session status loop — archive step resets finished to working
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-19T15:44:28.817617-03:00"
updated: "2026-04-19T15:44:28.817617-03:00"
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

# FB-572: ccs merge session status loop — archive step resets finished to working

## Problem
ccs merge auto-commits metadata (archiving), which rewrites session.json status from "finished" back to "working"/"conflict". The merge gate then refuses because status != "finished". This creates an infinite loop: finish → merge → auto-commit resets status → merge refuses.

## Current vs Expected
Current: After `ccs session finish <name>` sets status=finished, running `ccs merge <name>` auto-commits archive artifacts which set status=working/conflict, then the merge gate fails with "session.status=working (expected finished)".

Expected: ccs merge should either:
1. Not overwrite session status during archiving, OR
2. Set status=finished after archiving completes, before the merge gate check

## Why It Matters
Any worktree with metadata changes cannot be merged via ccs merge. This blocks the entire worktree workflow — the primary development pattern in this project.

## Priority
High — this is a blocking bug in the core merge workflow. Currently happening with the mcp-extension worktree.

## Reproduction
1. ccs spawn test-branch
2. Make changes, commit in worktree
3. ccs session finish test-branch
4. ccs merge test-branch --skip-review --reason "test"
5. Observe: "Auto-committed N metadata file(s)" → session.json rewritten → "not ready to merge"

## Affected Files
GOrchestra/sessions/<name>/session.json — status field overwritten during archive step
ccs merge command — archive + gate ordering

## Suggested Implementation
In the merge flow, either:
- Move the session status gate check BEFORE the archive auto-commit step, OR
- After archiving, re-set status="finished" in session.json before the gate check, OR
- Skip the session gate when --skip-review is provided (already bypassing review)

