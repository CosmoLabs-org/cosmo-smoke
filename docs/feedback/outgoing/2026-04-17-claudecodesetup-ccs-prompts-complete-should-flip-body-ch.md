---
id: FB-510
title: ccs prompts complete should flip body checkboxes and workcheck should respect frontmatter status
type: bug
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-17T14:42:26.696142-03:00"
updated: "2026-04-17T14:42:26.696142-03:00"
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

# FB-510: ccs prompts complete should flip body checkboxes and workcheck should respect frontmatter status

## Summary
ccs prompts complete marks frontmatter status: COMPLETED but leaves '### [ ]' checkboxes in the body unchecked. ccs workcheck reads the checkboxes, so after completing a prompt it still shows 0/N goals completed.

## Repro
1. Run: ccs prompts complete <prompt-file>
2. Run: ccs workcheck --json
3. Observe: goals_completed: 0, despite frontmatter showing COMPLETED

Workaround used: sed -i '' 's/### \[ \]/### [x]/g' <file>

## Proposed Solution
Option A — ccs prompts complete should auto-flip all '### [ ]' → '### [x]' in the body.
Option B — ccs workcheck should prefer frontmatter status: COMPLETED over body checkbox parsing.

