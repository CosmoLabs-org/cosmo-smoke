---
id: FB-506
title: Auto-spawned road-<id>-<slug> worktrees conflict with Agent-tool Sonnet dispatch
type: bug
status: pending
priority: high
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-17T14:41:55.282777-03:00"
updated: "2026-04-17T14:41:55.282777-03:00"
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

# FB-506: Auto-spawned road-<id>-<slug> worktrees conflict with Agent-tool Sonnet dispatch

## Summary
When a TaskList item with a ROAD-xxx subject is marked in_progress, some automation (possibly a SessionStart hook, GOrchestra, or a skill side-effect) silently auto-creates a worktree named 'road-<id>-<slug>'. If a Sonnet agent was already dispatched via Agent tool (isolation:worktree) for the same work, two parallel worktrees race on identical files — completely surprising both user and Claude.

## Repro
1. Mark a task in_progress with subject 'ROAD-003: Watch mode' in TaskList
2. Simultaneously dispatch Agent tool with isolation:worktree for the same feature
3. Observe: road-003-watch-mode/ and agent-<hash>/ both exist, both targeting same work

## Why It Matters
Both patterns (TaskList with ROAD-xxx + Agent isolation:worktree dispatch) are explicitly endorsed in CLAUDE.md. They should not conflict. During a cosmo-smoke session this burned ~15 minutes: killed two Sonnet dispatches, cherry-picked the commit from the auto-spawned worktree, cleaned up orphans.

## Proposed Solution
Option A — Document which hook/skill/binary triggers auto-spawn and how to opt out (in CLAUDE.md superpowers rules).
Option B — Before Agent tool isolation:worktree dispatch, check for an existing task-slug worktree and warn.
Option C — Emit a visible announcement (🌿 Auto-spawning road-003-watch-mode...) at spawn time so it's not a silent surprise.

