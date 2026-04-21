---
id: FB-613
title: /project-upgrade should include project health checks
type: improvement
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-21T15:04:45.787356-03:00"
updated: "2026-04-21T15:04:45.787356-03:00"
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

# FB-613: /project-upgrade should include project health checks

Currently /project-upgrade focuses on version mismatches and structural upgrades but misses project hygiene issues that accumulate over time.

What it should surface:

1. Stale issue/idea statuses — FEAT-013 was open but fully implemented (802 tests, 12 commits). No automated check caught this.
2. Missing session summaries — versions with release notes but no docs/sessions/ file. We found 3 gaps (v0.2.0, v0.6.0, v0.9.0) manually via /triage.
3. Harvested ideas — IDEA-MO1FC22M was still seed status after full implementation. Upgrade audit should flag ideas whose titles match completed issue/roadmap items.
4. Uncommitted metadata — session-end docs, prompts, transcripts pile up. Upgrade should warn when >5 uncommitted doc files exist.

These are low-effort, high-signal checks. Suggest adding a health pass before the structural upgrade checks so users clean house before upgrading.

Discovered during cosmo-smoke session: /triage found all of these, /project-upgrade found none.

