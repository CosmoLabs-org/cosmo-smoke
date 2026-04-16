---
id: FB-455
title: Full CCS integration with standalone cosmo-smoke
type: feature
status: pending
priority: medium
complexity: complex
from_project: cosmo-smoke
from_path: /Users/gab/PROJECTS/cosmo-smoke
to_project: ClaudeCodeSetup
to_target: project
created: "2026-04-15T17:44:56.149962-03:00"
updated: "2026-04-15T17:44:56.149962-03:00"
suggested_conversion: feature
converted_to: null
related_issues: []
brainstorm_ref: null
suggested_workflow:
  - brainstorming
  - plan-mode
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

# FB-455: Full CCS integration with standalone cosmo-smoke

cosmo-smoke must remain standalone (own repo, own binary, zero CCS deps). CCS provides integration layer.

**Integration Points (in priority order):**

1. **ccs smoke** — Wrapper command
   - 'ccs smoke' → runs 'smoke run' in project root
   - 'ccs smoke init' → runs 'smoke init'
   - Passes through all flags

2. **/project-init** — Scaffolding
   - Detect project type, offer to run 'smoke init'
   - Add .smoke.yaml to new projects automatically

3. **/health-check** — Validation
   - Check .smoke.yaml exists
   - Run 'smoke run --dry-run' to validate config
   - Flag missing smoke tests as health issue

4. **/audit** — Compliance
   - 'Missing smoke tests' as audit finding
   - Track smoke test coverage across portfolio

5. **SessionStart hook** — Optional auto-run
   - Config flag: 'autoSmoke: true'
   - Run smoke tests at session start, warn if failing

6. **/session-end** — Pre-commit gate
   - Run smoke before committing (optional)
   - Block commit if smoke fails

7. **superpowers:systematic-debugging** — Integration
   - 'Run smoke first' as step 0 in debug protocol
   - If smoke fails, that's the bug

**Key constraint:** smoke binary stays 100% independent. No imports from CCS. CCS wraps/calls it.

## Suggested Workflow

1. brainstorming
2. plan-mode
3. implementation

