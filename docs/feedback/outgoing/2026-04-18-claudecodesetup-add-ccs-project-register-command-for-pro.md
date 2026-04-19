---
id: FB-545
title: Add ccs project register command for project-registry.yaml
type: feature
status: pending
priority: medium
complexity: ""
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-18T23:24:59.616138-03:00"
updated: "2026-04-18T23:24:59.616138-03:00"
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

# FB-545: Add ccs project register command for project-registry.yaml

## Problem
No CLI command exists to register a project in `~/.claude/project-registry.yaml`. The only way is manually editing the YAML file.

## Current vs Expected
Current:
```bash
# Must manually edit YAML:
cat >> ~/.claude/project-registry.yaml << 'EOF'
ClaudeCodeSetup:
  path: /Users/gab/PROJECTS/ClaudeCodeSetup
  aliases: [claudecodesetup, ccs-setup]
EOF
```

Expected:
```bash
ccs project register /Users/gab/PROJECTS/ClaudeCodeSetup --alias claudecodesetup --alias ccs-setup
# Auto-infers name from directory name if not provided
```

## Why It Matters
Every new project requires manual YAML editing. Cross-project feedback (`ccs feedback send <target>`) fails with 'unknown target' if the project isn't registered. This is the first friction point when onboarding a new project.

## Priority Justification
Medium — low frequency (once per project) but high friction when it blocks cross-project workflows.

## Reproduction Steps
1. Create a new project with CLAUDE.md
2. Try `ccs feedback send NewProject 'test'`
3. Get: `failed to resolve target 'NewProject': unknown target`
4. Must manually add to ~/.claude/project-registry.yaml

## Affected Files
`tools/ccsession/cmd/project.go` — new subcommand needed
`~/.claude/project-registry.yaml` — target file to modify

## Suggested Implementation
Add `ccs project register <path>` subcommand. Auto-detect project name from directory basename. Accept optional `--alias` flags. Validate path exists and contains CLAUDE.md or .smoke.yaml. Append to registry YAML preserving existing entries.

