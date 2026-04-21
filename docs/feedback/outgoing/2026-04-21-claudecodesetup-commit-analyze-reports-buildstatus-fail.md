---
id: FB-610
title: commit-analyze reports build_status fail for //go:build ignore files
type: idea
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-21T02:17:27.797596-03:00"
updated: "2026-04-21T02:17:27.797596-03:00"
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

# FB-610: commit-analyze reports build_status fail for //go:build ignore files

## Problem Description

\`ccs commit-analyze --json\` reports \`build_status: \"fail\"\` when the working tree contains Go files with \`//go:build ignore\` tags in subdirectories (e.g., GOrchestra agent archives). The Go compiler outputs \"build constraints exclude all Go files\" warnings for these packages, which commit-analyze interprets as a build failure.

## Current vs Expected

**Actual:**
\`\`\`
\$ ccs commit-analyze --json 2>&1
{
  \"quality\": {
    \"build_status\": \"fail\",
    \"build_error\": \"package .../0010-schema-extra-tests/files/cmd: build constraints exclude all Go files in ...\"
  }
}
\`\`\`

**Expected:**
\`build_status\` should be \"pass\" (or at least \"warn\") when the only \"errors\" are build constraint exclusions in non-main packages. The actual main package builds fine:
\`\`\`
\$ go build -o /dev/null ./cmd/
Go build: Success
\`\`\`

## Why It Matters

Every session that includes GOrchestra archive files in the working tree gets a false build failure. This triggers unnecessary investigation and can cause session-end pipelines to behave differently than expected. It also reduces trust in the quality gate — if build_status is wrong, what else might be wrong?

Affects any session where GOrchestra agent files are modified (currently 25+ files with \`//go:build ignore\`).

## Priority Justification

Medium. It doesn't block work (you can ignore it once you know it's a false positive), but it adds noise to every commit-analyze run and makes the quality gate output unreliable. New users will waste time investigating a non-existent build breakage.

## Reproduction Steps

1. Have any Go file with \`//go:build ignore\` tag in a subdirectory (e.g., GOrchestra/glm-agents/XXXX/files/...)
2. Run: \`ccs commit-analyze --json\`
3. Observe \`build_status: \"fail\"\` with build constraint exclusion messages

## Affected Files

- \`tools/ccsession/cmd/commit_analyze.go\` (or equivalent — the file that runs \`go build\` and parses output)
- The build check likely runs \`go build ./...\` which includes all packages including archived ones

## Suggested Implementation

Two approaches (either or both):

1. **Filter build constraint messages**: When parsing Go build output, treat \"build constraints exclude all Go files\" as non-fatal. These are informational, not errors.
2. **Scope the build check**: Run \`go build ./cmd/ ./internal/...\` instead of \`go build ./...\` to exclude archived test directories under GOrchestra/.
3. **Pattern exclusion**: Skip build checks for paths matching \`GOrchestra/glm-agents/*/files/\` since these are archived agent worktrees with intentional build ignores.

