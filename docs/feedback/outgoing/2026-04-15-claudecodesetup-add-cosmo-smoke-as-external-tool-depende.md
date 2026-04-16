---
id: FB-453
title: Add cosmo-smoke as external tool dependency
type: feature
status: pending
priority: medium
complexity: simple
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-15T17:42:58.993137-03:00"
updated: "2026-04-15T17:42:58.993137-03:00"
suggested_conversion: feature
converted_to: null
related_issues: []
brainstorm_ref: null
suggested_workflow:
  - implementation
response:
  acknowledged: null
  acknowledged_by: null
  started: null
  implemented: null
  rejected: null
  rejection_reason: null
  notes: ""
---

# FB-453: Add cosmo-smoke as external tool dependency

cosmo-smoke is now a standalone universal smoke test runner at CosmoLabs-org/cosmo-smoke.

**Motivation:** Standardize smoke testing across all 95 CosmoLabs projects with a single tool.

**Solution:**
1. Add .smoke.yaml to ClaudeCodeSetup for self-testing (build tools, verify ccs --help)
2. Document in docs/dependencies or README as external CosmoLabs tool
3. Consider 'ccs smoke' wrapper that delegates to ~/bin/smoke

Binary: ~/bin/smoke | Config: .smoke.yaml | Repo: CosmoLabs-org/cosmo-smoke

## Suggested Workflow

1. implementation

