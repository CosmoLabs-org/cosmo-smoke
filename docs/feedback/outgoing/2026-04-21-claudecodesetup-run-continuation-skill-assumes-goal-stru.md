---
id: FB-624
title: run-continuation skill assumes goal-structured prompts — handoff prompts fall through
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-21T17:38:15.506687-03:00"
updated: "2026-04-21T17:38:15.506687-03:00"
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

# FB-624: run-continuation skill assumes goal-structured prompts — handoff prompts fall through

## Problem
The \`/run-continuation\` skill\'s invariant I2 (\"every ### [ ] G-NN gets ONE TaskCreate\") and its Step 4 (\"Parse the prompt's ## Goals section\") assume all continuation prompts have formal \`## Goals\` with numbered items. Handoff/status prompts use \`## Outstanding Items\` and \`## Potential Next Steps\` instead — no goal markers.

When a handoff prompt is loaded, the skill has no guidance for what to do. I had to improvise with AskUserQuestion to pick a direction.

## Current vs Expected

**Current behavior:**
\`ccs prompts load-context\` succeeds. Skill says \"parse ## Goals\" — no ## Goals exists. Agent is left without a clear workflow: create tasks? ask the user? just execute?

**Expected:**
Skill should handle at least two prompt shapes:
1. Goal-structured prompts → I2 applies, create tasks per goal
2. Handoff/status prompts → offer the items as options via AskUserQuestion, create tasks for chosen direction

## Why It Matters
Handoff prompts are a common pattern (session-end generates them). Every time one is loaded via /run-continuation, the skill's workflow breaks at Step 4. This is the second time I've hit this (first was in ClaudeCodeSetup project).

## Priority Justification
Medium — affects a common workflow (session handoff → next session resume). The fix is a small addition to the skill's Step 4, not a redesign.

## Reproduction Steps
1. Have a handoff prompt (e.g., \`docs/prompts/2026-04-21-session-handoff.md\`) with ## Outstanding Items and ## Potential Next Steps but no ## Goals
2. Run \`/run-continuation\`
3. Skill loads successfully via load-context
4. Step 4 finds no ## Goals — no clear next step in the skill workflow

## Affected Files
- \`~/.claude/skills/run-continuation/skill.md\` (or wherever the /run-continuation skill is defined) — Step 4 needs a fallback for non-goal prompts

## Suggested Implementation
Add to Step 4 of the skill:

\`\`\`markdown
#### Step 4a: Goal-structured prompts
If ## Goals exists with ### [ ] G-NN items → create one TaskCreate per goal (current behavior).

#### Step 4b: Handoff/status prompts
If no ## Goals but ## Outstanding Items or ## Potential Next Steps exist:
1. Collect items into an AskUserQuestion picker
2. User selects direction
3. Create tasks for chosen items
4. Proceed to Step 5
\`\`\`

