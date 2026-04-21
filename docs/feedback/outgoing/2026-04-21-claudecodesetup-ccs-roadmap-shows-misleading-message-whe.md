---
id: FB-623
title: ccs roadmap shows misleading message when all items are completed
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-21T17:38:15.278158-03:00"
updated: "2026-04-21T17:38:15.278158-03:00"
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

# FB-623: ccs roadmap shows misleading message when all items are completed

## Problem
When all roadmap items are completed or promoted, \`ccs roadmap\` outputs:
\`\`\`
No roadmap items found
Capture an idea: ccs roadmap add "your idea here"
\`\`\`

This is misleading — 44 items exist in \`docs/roadmap/index.yaml\`. The command means \"no open/incomplete items\" but says \"no items found.\"

## Current vs Expected

**Current:**
\`\`\`bash
$ ccs roadmap
No roadmap items found

Capture an idea: ccs roadmap add "your idea here"
\`\`\`

**Expected:**
\`\`\`bash
$ ccs roadmap
✅ All roadmap items completed (44/44)

ROAD-001..ROAD-044 — all completed or promoted.
\`\`\`

Or at minimum: \"No open roadmap items\" instead of \"No roadmap items found.\"

## Why It Matters
Wasted 5 minutes investigating whether roadmap data was corrupted. The wording implies the index is empty, not that all work is done. Bites every time a project completes its roadmap.

## Priority Justification
Medium — cosmetic but causes false \"data loss\" panic. Quick fix (message wording only).

## Reproduction Steps
1. Have a project where all roadmap items have status: completed or status: promoted
2. Run \`ccs roadmap\`
3. Observe \"No roadmap items found\" — implies data is missing

## Affected Files
- \`tools/ccsession/cmd/roadmap.go\` (or wherever the roadmap list/show command renders output) — the \"No roadmap items found\" message text

## Suggested Implementation
Change the empty-result message to distinguish between \"index doesn't exist\" and \"all items are completed.\" Something like:
- No index file → \"No roadmap initialized. Run \`ccs roadmap init\`\"
- Index exists but all completed → \"All N roadmap items completed\"
- Index exists with open items → show them (current behavior)

