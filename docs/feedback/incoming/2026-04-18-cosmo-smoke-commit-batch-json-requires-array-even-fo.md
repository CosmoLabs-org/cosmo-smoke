---
id: FB-007
title: commit-batch JSON requires array even for single commit
type: idea
status: acknowledged
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: cosmo-smoke
to_target: self
created: "2026-04-18T22:35:54.527387-03:00"
updated: "2026-04-18T22:40:11.828687-03:00"
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

# FB-007: commit-batch JSON requires array even for single commit

## Problem
ccs commit-batch --json requires a JSON array even for a single commit. Passing a single object (without [] wrapper) produces: 'invalid JSON plan: json: cannot unmarshal object into Go value of type []cmd.CommitPlanEntry'

## Current vs Expected
Current: echo '{"files":["a.go"],"message":"fix: x"}' | ccs commit-batch --json
-> Error: invalid JSON plan: json: cannot unmarshal object into Go value of type []cmd.CommitPlanEntry

Expected: Either accept single object, or improve error message to say wrap in array brackets

## Why It Matters
Hit this twice in one session. Every user will hit this at least once.

## Priority Justification
Low friction but high frequency

## Reproduction Steps
echo '{"files":["README.md"],"message":"docs: test"}' | ccs commit-batch --json

## Affected Files
tools/ccsession/cmd/commit_batch.go - JSON unmarshal target is []CommitPlanEntry

## Suggested Implementation
Attempt []CommitPlanEntry first. On error, attempt CommitPlanEntry and wrap in slice. Or detect leading curly brace and auto-wrap.

