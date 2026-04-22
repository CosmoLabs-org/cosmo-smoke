---
id: FB-636
title: ccs commit-batch cannot force-add gitignored tracked files
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-22T02:22:56.833119-03:00"
updated: "2026-04-22T02:22:56.833119-03:00"
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

# FB-636: ccs commit-batch cannot force-add gitignored tracked files

## Problem
After v0.13.0 added a 71-pattern .gitignore, previously-tracked metadata files (.gorchestra/fingerprint-cache.json, GOrchestra/intel/*.json, docs/conversation-transcripts/) became gitignored. These files have uncommitted changes that need committing, but ccs commit-batch always runs git add without -f, so it fails on every attempt.

## Current vs Expected
Current behavior:
$ echo '[{"files": [".gorchestra/fingerprint-cache.json"], "message": "chore: sync metadata"}]' | ccs commit-batch --json --force
→ git add failed: The following paths are ignored by one of your .gitignore files: .gorchestra

Expected: ccs commit-batch --force should pass -f to git add (or have a dedicated --force-add flag) so gitignored-but-previously-tracked files can be committed.

## Why It Matters
The commit guard hook blocks raw git commit, routing through ccs commit-batch. When commit-batch can't handle gitignored files, there is no path to commit them. The only workaround is manually running git add -f + hoping the hook allows it (it doesn't — it blocks git commit too). Files are permanently stuck in dirty state.

This will recur whenever .gitignore patterns expand to cover previously-tracked paths.

## Priority Justification
Medium — affects post-release metadata sync. Not blocking code work, but creates permanent dirty state in git status that accumulates across sessions.

## Reproduction Steps
1. Add a path to .gitignore that was previously tracked
2. Modify that file
3. Try to commit it via ccs commit-batch
4. Observe: git add fails with gitignore error
5. Try --force flag: same error (flag is for filesize audit, not git add -f)

## Affected Files
.gorchestra/fingerprint-cache.json
GOrchestra/intel/architecture.json
GOrchestra/intel/status.json
docs/conversation-transcripts/2026-04-21_121103_f7ce16a6.md

Also: .version-registry.json was in the same group but wasn't gitignored.

## Suggested Implementation
Add a --force-add flag to ccs commit-batch that passes -f to the underlying git add call for each file group. Alternatively, detect when a file is in .gitignore but has a staged/modified status and auto-apply -f.

In the cosmo-smoke repo specifically, adding negation patterns to .gitignore would also fix it:
!.gorchestra/
!GOrchestra/intel/
!docs/conversation-transcripts/

