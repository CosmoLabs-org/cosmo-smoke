---
id: FB-536
title: commit-batch JSON should accept single object
type: feature
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-18T22:39:54.111329-03:00"
updated: "2026-04-18T22:39:54.111329-03:00"
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

# FB-536: commit-batch JSON should accept single object

## Summary
`ccs commit-batch --json` requires a JSON array even for a single commit. Passing a single object produces: `json: cannot unmarshal object into Go value of type []cmd.CommitPlanEntry`

## Motivation
Hit this twice in one session. Every user will hit this at least once. The error message doesn't explain the fix.

## Proposed Solution
Attempt `[]CommitPlanEntry` first. On error, attempt `CommitPlanEntry` and wrap in slice. Or detect leading `{` and auto-wrap.

## Reproduction
```bash
echo '{"files":["README.md"],"message":"docs: test"}' | ccs commit-batch --json
# Error: invalid JSON plan: json: cannot unmarshal object into Go value of type []cmd.CommitPlanEntry
```

## Affected Files
`tools/ccsession/cmd/commit_batch.go` - JSON unmarshal target is `[]CommitPlanEntry`

## Priority: medium (low friction but high frequency)

