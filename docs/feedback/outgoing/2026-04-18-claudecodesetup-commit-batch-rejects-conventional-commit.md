---
id: FB-529
title: commit-batch rejects conventional commit multi-scope with misleading error
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-18T18:26:38.099008-03:00"
updated: "2026-04-18T18:26:38.099008-03:00"
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

# FB-529: commit-batch rejects conventional commit multi-scope with misleading error

## Problem
ccs commit-batch rejects conventional commit messages with comma-separated scopes (e.g., feat(schema,runner):) with a misleading error message that suggests double-prefixing the type.

## Current vs Expected

**Current behavior:**
Input message: "feat(schema,runner): add skip_if conditional execution and env config merge"
Error: "missing conventional commit type prefix. Got: \"feat(schema,runner): ...\" Try: \"feat: feat(schema,runner): ...\""

**Expected behavior:**
Either accept comma-separated scopes (valid per conventional commits spec) or give an actionable error: "comma in scope not supported — use single scope"

## Why It Matters
Conventional commits explicitly support multi-scope like feat(scope1,scope2):. When Claude writes commits for cross-cutting changes, multi-scope is natural. The misleading "Try:" suggestion wastes a commit cycle and confuses the workflow.

## Priority
Medium — happens every time a change spans two packages. Workaround is trivial (flatten scope) but the misleading error message wastes debugging time.

## Reproduction
1. Stage changes across two packages (e.g., internal/schema/ + internal/runner/)
2. Run: printf '%s' '[{"files":["file1.go","file2.go"],"message":"feat(schema,runner): description"}]' | ccs commit-batch --json
3. Observe error suggesting "feat: feat(schema,runner): ..." as fix

## Affected Files
tools/ccsession/cmd/commit_batch.go (scope validation regex/parser)

## Suggested Implementation
Either:
(a) Accept comma-separated scopes in the validator regex, OR
(b) Reject with clear message: "multi-scope (commas) not supported in scope field — use single scope"

Option (b) is lower effort and still an improvement over the current misleading "Try:" suggestion.

