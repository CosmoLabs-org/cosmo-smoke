---
id: FB-544
title: ccs idea show <ID> should display idea details, not list all ideas
type: feature
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-18T23:24:59.36125-03:00"
updated: "2026-04-18T23:24:59.36125-03:00"
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

# FB-544: ccs idea show <ID> should display idea details, not list all ideas

## Problem
`ccs idea show IDEA-MO53CPHE` outputs the full idea listing table instead of showing the specific idea's details (title, description, tags, status, origin).

## Current vs Expected
Current:
```
ccs idea show IDEA-MO53CPHE
💡 **5 ideas**
ID             DATE         STATUS     SOURCE  TITLE
--             ----         ------     ------  -----
IDEA-MO53CPHE  2026-04-18   🌰 seed     agent   Make run field optional...
...all 5 ideas listed...
```

Expected: Show the full content of the specific idea file (equivalent to `cat docs/ideas/<slug>.md`).

## Why It Matters
When triaging ideas or deciding whether to promote, you need the full description and context. Currently must fall back to reading the file directly via Read tool or cat, defeating the purpose of the command.

## Priority Justification
Medium — the command exists but does something unexpected. Every session that uses `ccs idea show` hits this.

## Reproduction Steps
1. Run `ccs idea add 'test idea'`
2. Note the ID from output
3. Run `ccs idea show <ID>`
4. Observe it lists all ideas instead of showing the specific one

## Affected Files
`tools/ccsession/cmd/idea.go` — show subcommand likely missing ID-based lookup logic

## Suggested Implementation
Resolve the ID to the idea file path (using the same lookup as `ccs idea promote`), read it, and display its full content. Follow the same pattern as `ccs issues show <ID>` which correctly displays a single issue.

