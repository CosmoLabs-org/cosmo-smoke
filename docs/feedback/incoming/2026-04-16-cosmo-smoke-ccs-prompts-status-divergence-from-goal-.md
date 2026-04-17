---
id: FB-004
title: ccs prompts status divergence from goal checkboxes
type: idea
status: implemented
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: cosmo-smoke
to_target: self
created: "2026-04-16T17:12:51.901448-03:00"
updated: "2026-04-17T14:42:30.825244-03:00"
suggested_conversion: feature
converted_to: null
related_issues: []
brainstorm_ref: null
session: 2026
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

# FB-004: ccs prompts status divergence from goal checkboxes

ccs prompts complete marks frontmatter status: COMPLETED, but ccs workcheck reads the '### [ ]' markdown checkboxes in the body. After running ccs prompts complete on the v0.2 continuation prompt, workcheck still showed 0/8 goals because body checkboxes were unchecked. Had to manually sed -i '' 's/### \[ \]/### [x]/g' the file. Suggest: (a) ccs prompts complete should auto-flip all '### [ ]' → '### [x]' in the body, or (b) workcheck should prefer frontmatter status over body parsing. Repro: complete a prompt via ccs prompts complete, run ccs workcheck --json → goals_completed: 0.

