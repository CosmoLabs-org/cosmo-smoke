---
id: FB-565
title: roadmap update fails on YAML with duplicate keys — stale listing
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-19T13:45:55.518681-03:00"
updated: "2026-04-19T13:45:55.518681-03:00"
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

# FB-565: roadmap update fails on YAML with duplicate keys — stale listing

## Problem Description
\`ccs roadmap update ROAD-035 --status completed\` fails with a YAML parse error because ROAD-035.yaml has duplicate \`reporter\` keys under the \`implementation\` mapping. This prevents status updates, which makes \`ccs roadmap\` listing show stale data (items as \"captured\" when their actual status is \"completed\").

## Current vs Expected Behavior

Current:
\`\`\`
$ ccs roadmap update ROAD-035 --status completed
Exit code 1: failed to parse roadmap item ROAD-035: yaml: line 20: mapping values are not allowed in this context

$ ccs roadmap --all
  💡 ROAD-035   Export smoke results as OTel telemetry  # shows captured, wrong

$ grep status docs/roadmap/items/ROAD-035.yaml
status: completed  # file is correct
\`\`\`

Expected: \`ccs roadmap update\` should either handle duplicate keys gracefully (last-write-wins) or warn and skip the problematic field while still updating status.

## Why It Matters
Roadmap status drift causes confusion about what work is done. In this session, 4 completed OTel items (ROAD-035-038) all showed as \"captured\" in the listing because the YAML for ROAD-035 has a parse error. I had to fall back to \`grep\` on individual YAML files to verify actual status.

## Priority Justification
Medium — causes confusion during session-end and workcheck, but doesn't break builds or tests. Affects any project where roadmap items are edited manually and might have duplicate keys.

## Reproduction Steps
1. Create a roadmap YAML with duplicate keys under a mapping (e.g., two \`reporter:\` lines under \`implementation:\`)
2. Run \`ccs roadmap update ROAD-xxx --status completed\`
3. Observe parse error
4. Run \`ccs roadmap\` — item shows stale status

## Affected Files
- cosmo-smoke: \`docs/roadmap/items/ROAD-035.yaml\` lines 16-17 (duplicate \`reporter\` keys)
- CCS: \`internal/roadmap/storage.go\` or equivalent YAML parsing code — uses strict YAML parsing that rejects duplicates

## Suggested Implementation
Option A: Use \`yaml.v3\` with \`KnownFields(true)\` or a lenient decoder that takes last-value-wins on duplicate keys (this is standard YAML behavior).
Option B: Catch the parse error in \`roadmap update\` and provide a actionable message: \"ROAD-035.yaml has duplicate keys at lines 16-17. Fix the YAML or use --force to skip validation.\"
Option C: Add a \`ccs roadmap doctor\` command that scans all items for YAML issues and reports them.

## Session Context
cosmo-smoke session 2026-04-19. ROAD-035 through ROAD-038 were completed in v0.8.0 but showed as \"captured\" in \`ccs roadmap\` output. Had to manually verify via grep.

