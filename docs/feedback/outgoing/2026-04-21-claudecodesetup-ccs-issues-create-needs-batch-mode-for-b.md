---
id: FB-628
title: ccs issues create needs batch mode for bulk operations
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-21T19:25:34.252735-03:00"
updated: "2026-04-21T19:25:34.252735-03:00"
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

# FB-628: ccs issues create needs batch mode for bulk operations

## Problem
Creating 22 feature issues required 22 separate \`ccs issues create feat\` calls in a bash loop. Each call produces multiple lines of output. For portfolio-scale work (CosmoLabs has ~95 projects), bulk issue creation is a real workflow need.

## Current vs Expected

**Current:**
\`\`\`bash
for type in java ruby php deno ...; do
  ccs issues create feat "\${type} project detection" --description "..." --bare
done
# 22 calls, 66+ lines of output, slow
\`\`\`

**Expected:**
\`\`\`bash
# Option A: from file
ccs issues create-batch --from-file batch.yaml

# Option B: stdin
cat batch.yaml | ccs issues create-batch

# Option C: multi-title
ccs issues create feat "Title 1" "Title 2" "Title 3" --description "shared desc" --bare
\`\`\`

## Why It Matters
CosmoLabs operates ~95 projects. Any time we do portfolio-wide work (adding project types, standardizing configs, bulk feature requests), we create issues in batches. This session needed 22 issues. Future sessions could need 50+. One-at-a-time creation is the bottleneck.

## Priority Justification
Medium — the loop works, but a batch command would reduce a 2-minute operation to 2 seconds and produce cleaner output. The pattern will repeat whenever we do portfolio-scale work.

## Reproduction Steps
1. Need to create 5+ issues at once
2. Run \`ccs issues create\` in a loop
3. Observe: verbose output, slow, no summary at end

## Affected Files
- \`tools/ccsession/cmd/issues.go\` — the issues create command handler

## Suggested Implementation
Add a \`--from-file\` flag to \`ccs issues create\` that reads a YAML array:

\`\`\`yaml
# batch-issues.yaml
items:
  - title: "Java/Maven project detection"
    type: feat
    description: "Detect pom.xml..."
    bare: true
  - title: "Ruby project detection"
    type: feat
    description: "Detect Gemfile..."
    bare: true
\`\`\`

\`ccs issues create --from-file batch.yaml\` creates all items, prints a summary:

\`\`\`
Created 22 issues: FEAT-014..FEAT-035
\`\`\`

Alternative simpler approach: accept multiple \`--title\` flags in a single call.

## Session Context
This session on cosmo-smoke created 22 FEAT issues (FEAT-014 through FEAT-035) for new project type detectors. Used a bash for-loop with individual \`ccs issues create feat\` calls. Each call took ~0.5s and printed 3 lines of output. Total: 22 calls, ~11 seconds, 66 lines of output. A batch command would have been 1 call, 1 line of output.

