---
id: FB-454
title: Study global cosmo-smoke integration for all CosmoLabs projects
type: feature
status: pending
priority: medium
complexity: medium
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-15T17:43:36.71657-03:00"
updated: "2026-04-15T17:43:36.71657-03:00"
suggested_conversion: feature
converted_to: null
related_issues: []
brainstorm_ref: null
suggested_workflow:
  - brainstorming
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

# FB-454: Study global cosmo-smoke integration for all CosmoLabs projects

cosmo-smoke is a portfolio-level tool (not CCS-specific) that should be integrated across all ~95 CosmoLabs projects.

**Motivation:** Every project needs smoke tests. Currently ad-hoc. cosmo-smoke standardizes this with a single binary + .smoke.yaml convention.

**Study Areas:**
1. **Project onboarding**: /project-init should offer to run 'smoke init' to scaffold .smoke.yaml
2. **CI integration**: GitHub Actions reusable workflow at CosmoLabs-org/cosmo-smoke/.github/workflows/smoke.yml
3. **CCS commands**: 'ccs smoke' wrapper that calls ~/bin/smoke with project context
4. **Session hooks**: SessionStart could run smoke tests automatically (optional)
5. **Health checks**: /health-check and /audit should verify .smoke.yaml exists and passes
6. **Documentation**: Add to docs/instructions/external-tools.md or similar

**Key principle:** smoke binary stays independent. CCS provides integration points, not ownership.

Binary: ~/bin/smoke | Repo: CosmoLabs-org/cosmo-smoke | First consumer: GoRalph

## Suggested Workflow

1. brainstorming
2. implementation

