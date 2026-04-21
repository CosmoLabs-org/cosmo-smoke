---
id: FB-627
title: TaskList empty at workcheck time — verification pipeline blind spot
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-21T19:23:58.120669-03:00"
updated: "2026-04-21T19:23:58.120669-03:00"
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

# FB-627: TaskList empty at workcheck time — verification pipeline blind spot

## Problem
TaskCreate creates in-memory tasks that don't survive context compaction or session boundaries. By the time /workcheck runs (often late in session), TaskList returns empty even though 14 tasks were created and completed. The workcheck skill relies on TaskList to cross-reference task completion against commits — if tasks are gone, verification is incomplete.

## Current vs Expected

**Current:**
- Created 14 tasks via TaskCreate during session
- All marked completed as work finished
- Ran /workcheck late in session
- TaskList returned: \"No tasks found\"
- CCS workcheck JSON showed 13 tasks from a DIFFERENT session (CCS's own tracking), not the current session's TaskCreate tasks
- Workcheck Step 2 says to persist to .claude/task-log.jsonl, but by then the data is already gone

**Expected:**
- TaskList should still show the current session's tasks, OR
- Tasks should be auto-persisted as they're created so workcheck can recover them

## Why It Matters
The /workcheck skill's Step 4 (\"Verify Tasks\") and Step 5 (\"cross-reference TaskList against commits\") are fundamentally broken if TaskList is empty. This session had 14 completed tasks with zero evidence at verification time. The gap means workcheck can never reliably answer \"did we finish what we started?\" for task-tracked work.

This will bite every long session that uses TaskCreate for tracking.

## Priority Justification
Medium — the /workcheck skill is the quality gate at session end. If it can't see completed tasks, it can't catch dropped work. The workaround (CCS workcheck JSON) catches some things but misses TaskCreate-only items.

## Reproduction Steps
1. Start a session, do significant work
2. Create 10+ tasks via TaskCreate, mark them completed as you go
3. Work for 30+ minutes (enough for context to shift/compact)
4. Run /workcheck
5. Call TaskList — observe \"No tasks found\"
6. CCS workcheck JSON may show old tasks from its own system, but not the TaskCreate tasks

## Affected Files
- The /workcheck skill definition (wherever TaskList is called in Step 1 and Step 4)
- Potentially: a new persistence layer for TaskCreate tasks

## Suggested Implementation
Two options:

**Option A: Auto-persist on create.** Every TaskCreate call appends to `.claude/task-log.jsonl`. Workcheck reads from this file instead of (or in addition to) in-memory TaskList. Survives compaction.

**Option B: Workcheck reads CCS task data.** The CCS workcheck JSON already has a task tracking system. Make the /workcheck skill prefer this over in-memory TaskList, since it's more durable.

Option A is simpler and doesn't depend on CCS. The workcheck skill could add a Step 0: \"read .claude/task-log.jsonl if it exists, merge with TaskList.\"

