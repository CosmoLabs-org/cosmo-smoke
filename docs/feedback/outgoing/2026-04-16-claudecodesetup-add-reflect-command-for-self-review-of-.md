---
id: FB-467
title: Add /reflect command for self-review of tools and SOPs used during session
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-16T09:38:10.521957-03:00"
updated: "2026-04-16T09:38:10.521957-03:00"
suggested_conversion: feature
converted_to: null
related_issues: []
brainstorm_ref: null
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

# FB-467: Add /reflect command for self-review of tools and SOPs used during session

## Problem
During sessions, Claude uses many CCS tools, SOPs, skills, commands, and agents. Sometimes these have gaps, friction, missing features, or architectural problems. Currently there's no systematic way to capture these observations — they get lost when the session ends.

Today's example: Used `/makefile` which ran `ccs makefile init`. The generated Makefile was missing Go ldflags for version injection. This gap was only discovered because the user explicitly asked "did you notice any problems?" After prompting, detailed feedback (FB-466) was filed. Without that prompt, the observation would have been lost.

## Current state
- No command exists for "reflect on tools you just used"
- Session-end Phase 3c (Idea Capture) focuses on code observations, not tool/SOP feedback
- Tool friction goes unnoticed unless user asks

## Proposed solution
Create `/reflect` (or `/feedback-ccs` or `/tool-review`) command that:

1. **Prompts self-review questions:**
   - "Which CCS tools, SOPs, skills, commands, or agents did you use this session?"
   - "Were there any gaps, friction, missing features, or problems?"
   - "Did any tool produce unexpected or suboptimal output?"
   - "Did you have to work around any limitations?"

2. **If issues found, sends detailed feedback to CCS including:**
   - Problem description
   - Current vs Expected behavior
   - Why it matters
   - Priority justification
   - Reproduction steps
   - Affected files (if known)
   - Suggested implementation (if obvious)

3. **Accepts honest "nothing to report"** — not every session surfaces issues

## Integration with session-end

Add as **Phase 3d: Tool Reflection** in full/complete/release modes:

```
Phase 3d: Tool Reflection
━━━━━━━━━━━━━━━━━━━━━━━━
Review tools, SOPs, and skills used this session.
Any gaps, friction, or problems worth feeding back to CCS?
```

**Gate:** Skip in quick/lean/commit-only modes. Active in full/complete/release/continue.

## Priority justification
**High** — This is a compounding improvement. Every session becomes an opportunity to improve tooling. The more CCS is used, the better it gets. Without this, tool friction accumulates silently and users must manually prompt for reflection.

## Reproduction of the gap
Any session where tools have friction — the friction goes unnoticed unless user explicitly asks.

## Affected files
- `skills/reflect/skill.md` or `commands/reflect/` — new command
- `skills/session-end/skill.md` — add Phase 3d
- `docs/standards/feedback-quality.md` — document required sections

## Suggested implementation
1. Create `/reflect` skill with self-review prompts
2. Add Phase 3d to session-end SOP (after 3c Idea Capture)
3. Document feedback completeness requirements
4. Make Phase 3d announce: "Reflecting on tools used..." and either file feedback or report "No tool issues observed"

## Why this matters
CCS is a living system. The best feedback comes from actual usage, not hypothetical planning. Claude notices friction that users might not articulate. This creates a virtuous cycle: use tools → notice gaps → file feedback → improve tools → repeat.

