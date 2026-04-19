---
id: FB-556
title: session-end-pre change_profile misses session-wide scope when source changes are committed mid-session
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-19T05:47:34.427742-03:00"
updated: "2026-04-19T05:47:34.427742-03:00"
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

# FB-556: session-end-pre change_profile misses session-wide scope when source changes are committed mid-session

## Problem

`ccs session-end-pre --json` returns `change_profile.category` based solely on uncommitted changes. When source code changes are committed mid-session (e.g., by project-upgrade), session-end-pre classifies the session as `config-only` and all quality gates (simplify, verify, review) are skipped — even if the session included substantial source changes.

## Current vs Expected

**Current:**
```json
{
  "change_profile": {
    "category": "config-only",
    "go_packages_touched": [],
    "has_untested_source": false,
    "cli_changed": false,
    "uncommitted_files": ["GOrchestra/intel/architecture.json", "GOrchestra/intel/status.json"]
  }
}
```

**Expected:** session-end-pre should also consider commits made during the current session (since session start or since last session-end). A session that added 5 new Go files, 25 tests, and modified 8 source files should not be classified as `config-only`.

## Why It Matters

Quality gates exist to catch issues before they ship. When session-end-pre reports `config-only`, the session-end SOP skips simplify (1.5), verification (1.75), and code review (1.8) — the three most valuable quality checks. This happened in a session that added 526 lines of new Go code with 25 tests.

## Priority Justification

High — quality gates are the primary value of session-end. When they're bypassed on substantial source sessions, the safety net is silently disabled.

## Reproduction Steps

1. Start a session, implement significant source changes
2. Run project-upgrade or any process that commits the source changes mid-session
3. Run `ccs session-end-pre --json`
4. Observe `change_profile.category: config-only` (only uncommitted metadata files)
5. All quality gates skip

## Affected Files

- `tools/ccsession/cmd/session_end_pre.go` — the change_profile logic
- `~/.claude/rules/quality-gate.md` — the SOP that gates on change_profile.category

## Suggested Implementation

Two approaches:

**A) Session-scope flag:** Add `--session` flag to session-end-pre that also analyzes commits since session start (detect via session transcript or git log since last session-end commit). Merge the uncommitted profile with the session-committed profile, taking the union of go_packages_touched and the more impactful category.

**B) Dual profile:** Return both `change_profile` (uncommitted only) and `session_profile` (full session scope). The SOP uses `session_profile` for quality gate decisions and `change_profile` for commit staging.

