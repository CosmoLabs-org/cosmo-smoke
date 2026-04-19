---
id: FB-008
title: session-end-finalize fails on missing .claude/release-context/
type: idea
status: acknowledged
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: cosmo-smoke
to_target: self
created: "2026-04-18T22:36:05.856799-03:00"
updated: "2026-04-18T22:40:11.922813-03:00"
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

# FB-008: session-end-finalize fails on missing .claude/release-context/

## Problem
ccs session-end-finalize fails with: fatal: pathspec '.claude/release-context/' did not match any files. This directory does not exist in cosmo-smoke (it's a ClaudeCodeSetup concept). The finalize command assumes it exists.

## Current vs Expected
Current: ccs session-end-finalize -> stage-artifacts step fails with pathspec error
Expected: Skip release-context staging if directory doesn't exist, or create it

## Why It Matters
Blocks session-end Phase 6 in projects that aren't ClaudeCodeSetup. Every cosmo-smoke session-end hits this.

## Priority Justification
Medium — session-end still partially completes (mark-phase-6 succeeds), but the failure is noisy

## Reproduction Steps
ccs session-end-finalize

## Affected Files
tools/ccsession/internal/sessionend/finalize.go (or wherever stage-artifacts lives)

## Suggested Implementation
Add existence check before git add for release-context directory. If missing, skip silently.

