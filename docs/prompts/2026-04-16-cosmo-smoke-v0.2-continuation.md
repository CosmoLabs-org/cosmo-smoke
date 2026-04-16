---
title: "cosmo-smoke v0.2 — Continue Feature Development"
created: 2026-04-16
status: PENDING
priority: high
branch: master
origin: "/continuation-prompt"
tags: [continuation, v0.2, features]
goals_total: 8
goals_completed: 0
carried_over_from: null
carried_over_goals: 3
related_prompts:
  - docs/prompts/2026-04-16-session-handoff.md
  - docs/prompts/2026-04-16-glm-dispatch-assertions.yaml
---

# cosmo-smoke v0.2 — Continue Feature Development

## File Scope
```yaml
files_modified:
  - internal/schema/schema.go
  - internal/runner/runner.go
  - internal/runner/assertion.go
  - internal/runner/assertion_test.go
  - internal/reporter/terminal.go
  - internal/reporter/json.go
  - cmd/run.go
files_created:
  - internal/reporter/tap.go
  - internal/reporter/tap_test.go
```

## Context

cosmo-smoke is gaining serious momentum. This session was highly productive: we implemented `smoke serve` (the Goss-killer feature), created 7 comprehensive example configs, added port_listening assertions, and captured extensive research from Grok identifying our competitive positioning against Goss (5.9k stars, stalled development).

Version was corrected from premature 1.0.0 to proper 0.1.0 beta. We now have 9 assertion types and are building toward v0.2.0 with HTTP assertions, config inheritance, and full CI integration.

4 GLM agents were dispatched but 3 timed out — their work is archived with recovery patches. The port_listening assertion was successfully merged through the quality gate.

## GLM Dispatch Rules

When goals involve dispatching subagents:

1. **ALWAYS** use `ccs glm-agent exec` for GLM agents (routes through queue with retry logic)
2. **NEVER** use Agent tool with `model:sonnet` or `model:haiku` for GLM work (bypasses queue, risks 429 rate limits)
3. Agent tool with `model:opus` is fine for Opus subagents
4. For parallel work: use `/glm-sprint` or `ccs glm-agent exec-batch`

## What Got Done

**Core Features Built:**
- `smoke serve` command — HTTP health endpoint for Docker/K8s probes (cmd/serve.go)
- port_listening assertion — TCP/UDP port checks (FEAT-004 merged via quality gate)
- stderr_matches + env_exists assertions — 9 total assertion types now
- JUnit XML output format — CI integration ready
- GitHub Actions workflows — ci.yml + reusable smoke.yml

**Documentation & Examples:**
- 7 example configs: go-api, node-fullstack, python-fastapi, docker-compose, rust-cli, monorepo, kubernetes
- Comprehensive research doc from Grok competitive analysis
- CCS integration vision document
- 13 ideas + 21 roadmap items captured

**Version & Process:**
- Fixed version: v1.0.0 → v0.1.0 (proper beta progression)
- Sent 4 feedback items to CCS (FB-453/454/455/456) about global integration

## Goals

### [ ] 1. Recover and complete timed-out GLM work
Recovery patches saved in:
- `GOrchestra/sessions/_glm-agent-0002-*/recovery.patch` — Process running assertion (FEAT-005)
- `GOrchestra/sessions/_glm-agent-0003-*/recovery.patch` — TAP output format (FEAT-006)  
- `GOrchestra/sessions/_glm-agent-0004-*/recovery.patch` — allow_failure flag (FEAT-007)

Apply patches or re-dispatch with enriched prompts.

### [ ] 2. Push to GitHub
Run `/move-to-github` to create CosmoLabs-org/cosmo-smoke and push all work.

### [ ] 3. Implement HTTP endpoint assertions (ROAD-006)
The most requested feature. Add:
- `http` assertion type with status, headers, body, JSONPath matching
- Support for GET/POST methods
- Timeout configuration

### [ ] 4. Add config inheritance (ROAD-020)
Goss-style config composition:
- `includes:` to import other .smoke.yaml files
- `--vars file.yaml` for variable injection
- Go template support `{{ .Env.FOO }}`

### [ ] 5. Implement JSON field assertions (ROAD-007)
Parse stdout as JSON and assert on fields:
- `json_field: '.version'` with `equals:`, `contains:`, `matches:`
- Use gjson or similar for JSONPath

### [ ] 6. Add auto-generate magic (ROAD-023)
`smoke init --from-running` to inspect running container and generate .smoke.yaml automatically. This is Goss's beloved `autoadd` feature.

### [ ] 7. Update CLAUDE.md assertion table
Add the new assertion types (env_exists, stderr_matches, port_listening) to the documentation table.

### [ ] 8. Run full test suite and self-smoke
Verify everything works: `go test ./...` + `smoke run`

## Carry-Over Tasks

From timed-out GLM agents (work partially done, recovery patches available):
- [ ] FEAT-005: Process running assertion (was: timeout)
- [ ] FEAT-006: TAP output format (was: timeout)
- [ ] FEAT-007: allow_failure for flaky tests (was: timeout)

## Where We're Headed

**v0.2.0 Target:**
- HTTP + JSON assertions (covers 90% of API smoke testing)
- Config inheritance (DRY configs across environments)
- TAP + Prometheus output formats

**Strategic Position:**
Goss development has stalled (18+ months no major release). We're positioned to be the modern successor with:
- Universal language support (not Linux-centric)
- Multi-language auto-scaffolding
- Container-native design (`smoke serve`)
- Active development momentum

**CCS Integration:**
4 feedback items sent to CCS. Next step is CCS implementing `ccs smoke` wrapper and `/project-init` integration.

## Priority Order

1. **Push to GitHub** — make the repo public, enable collaboration
2. **Recover GLM work** — complete the timed-out features (process, TAP, allow_failure)
3. **HTTP assertions** — highest user value, core missing feature
4. **Config inheritance** — essential for portfolio-scale adoption
5. **Auto-generate** — the "magic" that wins users from Goss
