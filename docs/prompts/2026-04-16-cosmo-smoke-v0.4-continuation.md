---
title: "cosmo-smoke v0.4 — Continue from v0.3.0 Assertion Pack"
created: 2026-04-16
status: PENDING
priority: high
branch: master
origin: "/session-end"
tags: [continuation, v0.4, roadmap]
goals_total: 6
goals_completed: 0
carried_over_from: 2026-04-16-cosmo-smoke-v0.2-continuation.md
carried_over_goals: 0
related_prompts:
  - docs/prompts/2026-04-16-cosmo-smoke-v0.2-continuation.md
---

# cosmo-smoke v0.4 — Continue from v0.3.0 Assertion Pack

## Context

v0.3.0 shipped "The Assertion Pack" (15 assertion types, prometheus reporter, allow_failure flag) via parallel Sonnet agent dispatch with Opus quality gate. Test suite: 176 passing. Ready for v0.4.

The session proved a repeatable multi-agent workflow: Opus brief → Sonnet agents in worktrees → Opus review → merge with union-resolve. This pattern should be used for v0.4's assertion additions.

## GLM/Sonnet Dispatch Rules

1. Use `Agent` tool with `model: sonnet` and `isolation: worktree` for bounded implementation work
2. Always provide: exact file paths, exact patterns to follow (reference existing assertion like port_listening), test expectations, commit format
3. Opus reviews via `ccs verify-worktree <name> --approve` — score < 7/10 = fix in worktree before merge
4. Sequential merges with `ccs merge <name> --force` — expect conflicts on schema.go/assertion.go/runner.go, resolve by taking all additions

## Goals (6 total, from v0.4-tagged roadmap)

### [x] 1. Push v0.3.0 to GitHub
Run `git push origin master --tags` to publish on `CosmoLabs-org/cosmo-smoke`.

### [ ] 2. ROAD-024: Goss migration tool (priority 85)
Parse Goss YAML, emit .smoke.yaml. Goss dev stalled 18+ months — this captures their users. Effort: large. Needs: brainstorm first to spec mapping (Goss's `package:`, `service:`, `process:` → cosmo-smoke's assertions).

### [ ] 3. ROAD-012: Retry with backoff (priority 80)
`retry: {count: 3, backoff: 1s}` on test level. Pairs with allow_failure for flaky-network tests. Small effort. Good Sonnet candidate.

### [ ] 4. ROAD-015: Docker container smoke tests (priority 75)
Runner side of `smoke serve`: spin docker container, run smoke tests against it, report. Medium effort. Needs: docker CLI wrapping, test-server lifecycle.

### [ ] 5. ROAD-016 continuation: Postgres/MySQL connectivity (priority 70)
redis_ping + memcached_version shipped in v0.3. Remaining: postgres_ping, mysql_ping using stdlib net + protocol handshake (no new deps). Small-medium effort. Good Sonnet candidate.

### [ ] 6. ROAD-003: Watch mode (priority 65)
`smoke run --watch` using fsnotify on config dir. Developer ergonomics. Small effort.

## What Got Done (v0.3.0)

- 6 new assertions: process_running, response_time_ms, ssl_cert, redis_ping, memcached_version, grpc_health
- allow_failure flag
- Prometheus text-format reporter
- Makefile ldflags version injection (FB-001)
- Test suite: 144 → 176 tests
- Workflow pattern: 7 parallel Sonnet agents proven

## Carry-Over Tasks

None — v0.2 continuation prompt fully closed. All 8 goals shipped.

## Where We're Headed

**v0.4.0 theme:** Production DX (retry, docker, more DBs, watch).
**Strategic move:** Goss migration tool (ROAD-024) — first project with a migration tool for a stalled competitor.
**v0.5 targets:** Conditional execution (ROAD-008), multi-env (ROAD-017), monorepo recursive (ROAD-010).

## Priority Order

1. Push v0.3.0 to GitHub (one command)
2. ROAD-012 Retry + ROAD-016 Postgres/MySQL (small, parallel-dispatch candidates)
3. ROAD-024 Goss migration (large, brainstorm first)
4. ROAD-003 Watch mode (small, nice-to-have)
5. ROAD-015 Docker runner (medium, after retry lands)

## Constraints
- Use `ccs prompts complete` to close THIS prompt when next session finishes
- Budget: ~3-4 hours per session
- Goal checkboxes: `### [ ]` → `### [x]` when done
