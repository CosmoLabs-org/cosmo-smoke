---
id: FB-511
title: cosmohooks health falsely flags Edit+Write dual-registration as duplicates
type: bug
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-17T14:45:33.807686-03:00"
updated: "2026-04-17T14:45:33.807686-03:00"
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

# FB-511: cosmohooks health falsely flags Edit+Write dual-registration as duplicates

## Summary
cosmohooks health reports a warning for changelog-guard, version-registry-guard, and agent-config-guard being registered twice. But these are intentional dual-registrations: once on PreToolUse:Edit and once on PreToolUse:Write. The health check is comparing command names only, not (event, matcher, command) tuples.

## Repro
1. Run: cosmohooks health
2. Observe: ⚠️ settings-duplicates: duplicate hook entries: changelog-guard(2x), version-registry-guard(2x), agent-config-guard(2x)
3. Run: ccs hooks list | grep changelog-guard
4. Observe: two entries with DIFFERENT matchers (Edit vs Write) — not duplicates

## Why It Matters
The health check is a trust signal. A persistent false-positive warning trains users to ignore it, which means real duplicate bugs will be missed. This one has been showing up every session.

## Proposed Fix
In the duplicate detection logic, compare the full tuple (event, matcher, command) rather than command name alone. Two hooks with the same command but different matchers are valid and intentional — they guard both Edit and Write tool uses independently.

