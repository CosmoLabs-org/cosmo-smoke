---
id: FB-456
title: Study cosmo-smoke for universal smoke testing integration
type: feature
status: pending
priority: medium
complexity: medium
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-15T17:45:49.905216-03:00"
updated: "2026-04-15T17:45:49.905216-03:00"
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

# FB-456: Study cosmo-smoke for universal smoke testing integration

cosmo-smoke (CosmoLabs-org/cosmo-smoke) is a standalone universal smoke test tool. CCS should study how to integrate it across all workflows.

**Study areas:**
1. How /project-init can auto-scaffold .smoke.yaml
2. How /health-check can verify smoke tests pass
3. How SessionStart can optionally run smoke tests
4. How Dependencies SOP can generate smoke prerequisites
5. How Credentials SOP can generate env_exists assertions
6. How /audit can flag projects missing smoke tests

**Existing roadmap items in cosmo-smoke:**
- ROAD-013: Dependency version assertions (links to Dependencies SOP)
- ROAD-014: Credential smoke tests (links to Credentials SOP)
- ROAD-015: Docker smoke tests
- ROAD-016: Database connectivity
- ROAD-017: Multi-environment configs
- ROAD-018: Service dependency checks

**Action:** Review these expansions and plan CCS-side integration points.

## Suggested Workflow

1. brainstorming
2. implementation

