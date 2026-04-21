---
id: FB-629
title: /reflect skill skips too aggressively — needs better judgment on what improves workflow
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-21T19:28:07.517731-03:00"
updated: "2026-04-21T19:28:07.517731-03:00"
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

# FB-629: /reflect skill skips too aggressively — needs better judgment on what improves workflow

## Problem
The /reflect skill's seven-section quality bar causes false negatives. I skipped 2 observations this session that the user had to push back on — both turned out to be real, actionable improvements (FB-627: TaskList persistence, FB-628: batch issue creation). The skill's filter is too strict and its default posture is "skip unless perfect."

## Current vs Expected

**Current behavior:**
- Seven-section quality bar: if any section feels thin, skip the observation
- Default to "No actionable friction" rather than erring on filing
- Category labels like "platform issue" or "thin nice-to-have" are used to justify skipping
- User has to push back to get things filed

**Expected behavior:**
- Default to filing if the observation would improve daily workflow
- The question should be: "Would fixing this make our next session smoother?" not "Can I fill all 7 sections perfectly?"
- Every workaround (python3, sed, bash loop) should auto-qualify — workarounds ARE friction
- If the user notices something enough to question the skip, it should have been filed

## Why It Matters
Reflect is the self-improvement loop. If it filters out real improvements, the system stagnates. This session: 2 real improvements almost didn't get filed. The user caught both. Next time, the user might not be watching as closely and we lose the signal.

The filter is doing the opposite of its job — it's preventing improvement rather than ensuring quality.

## Priority Justification
High — reflect is the session-over-session improvement mechanism. If it's broken, nothing improves.

## Reproduction Steps
1. Have a productive session with minor friction (workarounds, missing features)
2. Run /reflect
3. Observe: agent skips observations using labels like "platform issue" or "thin"
4. User pushes back
5. Observation turns out to be a real, actionable improvement

## Affected Files
- `~/.claude/skills/reflect/skill.md` — the /reflect skill definition

## Suggested Implementation

### Change 1: Replace 7-section gate with workflow-impact test
Current: "Can you fill all 7 sections? If not, skip."
New: "Would fixing this make the next session smoother? If yes, file it. The 7 sections are a guide for quality, not a gate for permission."

### Change 2: Auto-qualify workarounds
Any time the agent uses python3, sed, bash loops, or manual workarounds to route around a missing tool/feature/command, that observation auto-qualifies for filing. The workaround IS the evidence.

### Change 3: Add "workflow improvement" category
Not everything is a bug. Add a tier system:
- **Bug**: Something broken or misleading (7-section gate applies)
- **Workflow improvement**: Something that could be better (lower bar — just needs problem + suggested fix)
- **Observation**: Something noticed but not actionable (skip is OK here)

### Change 4: Anti-pattern — "Nothing to report" should be rare
If a session did non-trivial work and used any workarounds, "Nothing to report" is suspicious. Add a check: "Did you use any workarounds this session? If yes, at least one of those should become feedback."

## Session Context
cosmo-smoke session where I:
1. Used python3/sed 3 times to work around Edit tool failures → skipped as "platform issue"
2. Created 22 issues one-at-a-time → skipped as "thin nice-to-have"
3. TaskList was empty at workcheck → skipped as "no actionable fix"

User pushed back on #2 and #3. Both became real FBs (FB-627, FB-628). #1 was genuinely platform-level but I should have at least noted it as workflow friction rather than dismissing it entirely.

The pattern: I was optimizing for "don't file noise" when I should optimize for "don't miss signal."

