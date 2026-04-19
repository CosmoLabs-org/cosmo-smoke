---
id: FB-549
title: ccs version --bump highlights overlap detection too aggressive — rejects valid descriptions containing phrases that match existing changelog entries
type: bug
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-19T03:33:54.570948-03:00"
updated: "2026-04-19T03:33:54.570948-03:00"
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

# FB-549: ccs version --bump highlights overlap detection too aggressive — rejects valid descriptions containing phrases that match existing changelog entries

Summary: ccs version --bump highlights validation rejects valid descriptions that happen to match changelog entries. A narrative sentence containing 'OpenTelemetry trace correlation' was rejected because the changelog already has an entry with the same phrase. The overlap detection treats any substring match as a duplicate, even when the user is describing the same feature in prose.

Motivation: Forces workaround of shortening or rewording descriptions to avoid false positives. Degrades the quality of commit/changelog messages since users write around the validator instead of describing accurately.

Proposed solution: Use exact entry-level matching (full changelog entry text) instead of substring overlap. Only flag as duplicate when the proposed description is substantially identical to an existing entry, not when it merely contains a shared phrase.

