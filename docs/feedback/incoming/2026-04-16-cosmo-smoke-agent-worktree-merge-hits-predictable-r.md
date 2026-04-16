---
id: FB-006
title: Agent worktree merge hits predictable .review.json + .ccsession.json friction
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: cosmo-smoke
to_target: self
created: "2026-04-16T20:43:28.041057-03:00"
updated: "2026-04-16T20:43:28.041057-03:00"
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

# FB-006: Agent worktree merge hits predictable .review.json + .ccsession.json friction

What happened: Every time I merge a worktree created by the Agent tool (isolation:worktree, model:sonnet), the 'ccs merge <name> --force --yes' flow fails with the same two issues:

1. '.review.json' conflicts because both master and the agent branch have a freshly-written review timestamp. Resolution: git checkout --theirs .review.json && git add .review.json
2. 'worktree X has N uncommitted changes' because '.ccsession.json' was modified post-last-commit. Resolution: git -C <worktree> checkout .ccsession.json before retrying merge

Happened 3x in one session on this project for: agent-a1e42bae (retry), agent-a281a237 (postgres/mysql), agent-aa6c4029 (docker, pre-merge). Pattern is fully reproducible.

Why it matters: This is mechanical friction on every merge. The resolution is always identical (take incoming .review.json, discard .ccsession.json). Any Claude session following the SOP will hit these same steps. That's wasted tool calls every time.

Proposed solution: In 'ccs merge --force', auto-resolve these two well-known files before reporting a conflict:
- .review.json: take incoming (the branch being merged was reviewed more recently)
- .ccsession.json: reset to base (it's session-tracking state, not feature work)

Both decisions are deterministic and documented — they should be encoded into the merge flow, not rediscovered every time.

Priority justification: Medium — not a blocker, but every multi-agent session pays this tax multiple times. An hour saved across all future sessions trivially pays for a one-time fix in ccs merge.

