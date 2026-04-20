---
id: FB-584
title: Edit tool fails to match tab-indented Go code
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-19T21:17:59.714228-03:00"
updated: "2026-04-19T21:17:59.714228-03:00"
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

# FB-584: Edit tool fails to match tab-indented Go code

## Problem description
The Edit tool cannot match strings in tab-indented Go files. When copying exact content from Read output (which shows tabs as spaces in the display), the match fails with \"String to replace not found.\" This happened ~5 consecutive times on cmd/run.go.

## Current vs expected
Current — attempted to replace a block in cmd/run.go with exact content from Read output:
\`\`\`
Edit: old_string=\"}\n\nfunc runWatch(configDir...\"
Error: String to replace not found
\`\`\`
Multiple attempts with varying indentation all failed. The file uses tabs (confirmed via xxd showing 0x09 bytes).

Expected — the Edit tool should match the exact content from Read output, including tab characters.

## Why it matters
Go projects use tabs by convention. Every Go file edit in this project will hit this. The workaround (sed -i) introduces its own problem: macOS sed converts \t to literal \"t\", requiring a second fix step.

## Priority justification
Medium. Affects every Go project session. Workaround exists but costs 3-4 extra tool calls per failed edit, which adds up.

## Reproduction steps
1. Open a Go file that uses tab indentation (e.g., cmd/run.go in cosmo-smoke)
2. Read a section of the file
3. Attempt an Edit with old_string matching the exact content shown
4. Observe \"String to replace not found\" error
5. Confirm the content exists via grep or xxd

## Affected files
- The Edit tool itself (Claude Code built-in)
- Any tab-indented Go file, e.g. cmd/run.go, internal/runner/runner.go

## Suggested implementation
The Edit tool should handle tab characters in old_string/new_string correctly. Two options:
1. Preserve tab characters as-is when matching (don't normalize to spaces)
2. Provide a tab-aware mode or allow \\t escapes in match strings

## Investigation Already Done
- Confirmed the file uses tabs via xxd (0x09 bytes)
- Confirmed the match string looks correct visually
- Tried 5+ variations of indentation (spaces, mixed, etc.)
- All failed — the Edit tool cannot match tab-indented blocks in this file
- Workaround: use sed -i for insertion, then Edit for simple one-line fixes

