---
id: FB-005
title: Auto-spawned road-<id>-<slug> worktrees conflict with Agent-tool Sonnet dispatch
type: idea
status: pending
priority: high
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: cosmo-smoke
to_target: self
created: "2026-04-16T20:43:17.618689-03:00"
updated: "2026-04-16T20:43:17.618689-03:00"
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

# FB-005: Auto-spawned road-<id>-<slug> worktrees conflict with Agent-tool Sonnet dispatch

What happened: During a session in /Users/gab/PROJECTS/cosmo-smoke, I marked TaskList items with ROAD-003/ROAD-015/ROAD-024 as in_progress. Some automation in the CCS environment (possibly a SessionStart hook, the goralph daemon, or a skill side-effect) auto-created a worktree named 'road-003-watch-mode' mirroring the task, and prompted the user to cd into it and run glm manually. Simultaneously I had already dispatched a Sonnet agent via the Agent tool with isolation:worktree, which created its own agent-<hash> worktree. Result: two parallel worktrees racing on the same work, user confused about which was which.

Why it matters: This is undocumented behavior that breaks the mental model of 'Agent tool with isolation:worktree handles its own worktree lifecycle'. The auto-spawned worktree surprised both me and the user. We burned ~15 minutes reconciling the state — killing my Sonnet dispatches, cherry-picking the commit from the auto-spawned worktree, and cleaning up orphans. Not a small friction cost for a standard task-tracking action.

Proposed solution: Option A — document the auto-worktree trigger prominently (which skill/hook/binary creates these, what tasks trigger it, how to opt out). Option B — add a pre-dispatch check in the Agent tool that detects an existing task-slug worktree and either skips auto-spawn or prompts the user. Option C — add a visible '🌿 Auto-spawning road-003-watch-mode worktree (triggered by: X)' announcement in the status line the moment the auto-spawn fires, so it's not a silent surprise.

Priority justification: High-friction DX bug that will recur every time someone uses TaskList with ROAD-xxx IDs while also using parallel Sonnet dispatch. Both patterns are explicitly endorsed in CLAUDE.md rules — they shouldn't fight each other.

Repro:
1. In any cosmo-smoke-like project, mark a task in_progress with 'ROAD-003: Watch mode' as subject
2. Simultaneously dispatch an Agent tool call with subagent_type:general-purpose, model:sonnet, isolation:worktree
3. Observe: two worktrees exist (road-003-watch-mode/ and agent-<hash>/), both targeting the same work.

