---
branch: master
completed: "2026-04-17"
created: "2026-04-16"
goals_completed: 2
goals_total: 2
origin: /continuation-prompt
priority: high
related_prompts:
    - docs/prompts/2026-04-16-cosmo-smoke-v0.2-continuation.md
status: COMPLETED
tags:
    - continuation
    - v0.4
    - roadmap
title: cosmo-smoke v0.4 — Finish Docker + Goss (4 of 6 shipped)
---

# cosmo-smoke v0.4 — Finish Docker + Goss

## File Scope

```yaml
files_modified:
  - internal/schema/schema.go        # where new assertion structs land
  - internal/schema/validate.go      # where new assertion validation lands
  - internal/runner/assertion.go     # where new check functions land
  - internal/runner/runner.go        # where new assertion wiring lands
  - CLAUDE.md                        # assertion table + command table
  - cmd/root.go                      # if adding `migrate` subcommand
files_created:
  - cmd/migrate.go                   # for Goss migration
  - internal/migrate/goss/parser.go
  - internal/migrate/goss/translator.go
  - internal/migrate/goss/emitter.go
  - internal/runner/assertion_test.go  # extend, not create
```

## Context

v0.4 is 4-of-6 shipped. The assertion-pack cadence from v0.3 held up — retry, postgres/mysql, watch mode all landed this session via a mix of Sonnet-agent dispatches and one manual glm-in-worktree (watch mode). Test suite went 176 → 201. Master is clean, 6 commits unpushed, ready to go.

Two goals remain: **Goss migration tool** (strategic; design doc already drafted in `docs/brainstorming/2026-04-16-goss-migration-tool-design.md`, recommend deferring to v0.5 per that doc) and **Docker runner assertions** (small; direct Sonnet candidate). Project momentum is strong — each of these is under 2 hours of focused work.

## GLM/Sonnet Dispatch Rules

1. **Agent tool with `model: sonnet` + `isolation: worktree`** is the proven pattern for bounded implementation. Watch/retry/postgres all shipped this way.
2. Provide: exact file paths, exact patterns to follow (reference CheckRedisPing / CheckMemcachedVersion), test expectations, conventional commit format.
3. Review via `ccs verify-worktree <name> --approve` — Haiku scorer is fast (score 8/10 = ship).
4. Merge sequentially via `ccs merge <name> --force --yes` — expect `.review.json` conflicts and `.ccsession.json` dirty tree. Resolve by `git checkout --theirs .review.json` and `git checkout .ccsession.json` then retry.
5. **For Goss migration: do NOT one-shot.** Brainstorm first with user, then split into 2 Sonnet worktrees (parser+translator, emitter+CLI). Design doc already outlines phases — use it.
6. **Watch out for auto-spawned worktrees.** Marking a task `in_progress` with ROAD-xxx naming appears to trigger some automation that creates a parallel `road-<id>-<slug>` worktree. If you see one, check its session.json — it may already have completed work to cherry-pick (as happened for watch mode this session).

## What Got Done (this session)

- **v0.3.0 pushed to GitHub** (CosmoLabs-org/cosmo-smoke) — 50 commits + 3 tags landed
- **ROAD-012 retry with exponential backoff** — `retry: {count, backoff}` on test level, doubles per attempt, pairs with allow_failure. `runTest` wraps `runTestOnce`. +272/-7, 8 tests. (commit `032738e`)
- **ROAD-016 postgres_ping + mysql_ping** — stdlib net only, SSLRequest handshake for pg / server-initiated v10 handshake for mysql. +260/0, 8 tests including fake TCP listeners. (commit `c7d8ac4`)
- **ROAD-003 watch mode** — `smoke run --watch` via fsnotify, 500ms debounce, clean Ctrl+C, ignores chmod-only events. +125/-12, 7 table-test cases. Cherry-picked from auto-created `road-003-watch-mode` worktree after two false-start Sonnet dispatches. (commit `23260e0`)
- **Goss migration design doc** — `docs/brainstorming/2026-04-16-goss-migration-tool-design.md` — full Goss→cosmo-smoke mapping table, 6-phase implementation plan, 5 open questions, recommendation to defer to v0.5. Uncommitted.
- Tests: 176 → 201 (+25)
- Workflow lessons: parallel Sonnet dispatch works when briefs cite existing patterns literally. Stopping agents prematurely based on terminal tea-leaves wastes cycles. Automated ROAD-xxx worktree spawning exists somewhere in this environment — watch for it.

## Goals

### [ ] 1. ROAD-015: Docker runner assertions (priority 75, medium effort)

Ship `docker_container_running: {name}` and `docker_image_exists: {image}` as a bounded subset of ROAD-015. The full ROAD-015 scope (docker_build, docker_run, docker_health) is too large for a single worktree — defer those three to a follow-up.

**Pattern**: Follow `CheckRedisPing` / `CheckMemcachedVersion` in `internal/runner/assertion.go` but shell out to `docker` CLI via `exec.CommandContext`. Tests gracefully skip when docker daemon unavailable via `isDockerAvailable()` helper that probes `docker info`.

**Files to modify**: `internal/schema/schema.go` (2 new structs + 2 Expect fields), `internal/schema/validate.go` (required-name validation), `internal/runner/assertion.go` (CheckDockerContainerRunning + CheckDockerImageExists), `internal/runner/runner.go` (wiring after Memcached block), `CLAUDE.md` (assertion table).

**Tests**: 4 runner tests (2 always-run failure-path, 2 docker-gated pass-path) + 2 validate tests. Pattern mirrors the postgres/mysql dispatch from this session.

**Dispatch**: `Agent` tool, `model:sonnet`, `isolation:worktree`, `run_in_background:true`. The full brief was drafted this session — see conversation context or rewrite from the Redis/Memcached pattern.

**Acceptance**: Single conventional commit `feat(assertions): add docker_container_running and docker_image_exists`, 203-205 tests passing, `go vet` clean, `go.mod` unchanged (stdlib `os/exec` only).

### [ ] 2. ROAD-024: Goss migration tool (priority 85, large effort)

**Before dispatching any code agent: brainstorm with user** on the 5 open questions at the bottom of `docs/brainstorming/2026-04-16-goss-migration-tool-design.md`:

1. Multi-distro packages (dpkg hardcode vs `--distro=` flag vs autodetect)
2. Lossy vs strict mode default
3. Output file layout for `gossfile:` includes (flatten vs preserve)
4. Reverse migration scope (defer)
5. Bulk migration (per-file vs directory walk)

**Then**: split Phase 1-6 from the design doc across 2 parallel Sonnet worktrees. Worktree A does parser + translator (core subset: package, service, process, port, command, file, http). Worktree B does emitter + CLI wiring + golden-file tests. Merge order: A then B.

**Alternative path (recommended in design doc)**: defer entirely to v0.5. v0.4 ships cleanly at 5-of-6 (retry, pg/mysql, watch, docker + v0.3 baseline). Goss migration then gets its own focused v0.5 launch with marketing coordination. If user agrees, commit the design doc and move to v0.4 release.

**Files**: design doc lists the 11 files to create/modify.

**Acceptance (if building)**: `smoke migrate goss <input.yaml>` produces a valid `.smoke.yaml` that `schema.Load` accepts. Golden-file tests for each translator. `--stats` reports mapping counts. Single feature commit per worktree.

## Carry-Over Tasks

- [ ] ROAD-015 Docker runner (was: pending) — see Goal 1 above
- [ ] ROAD-024 Goss migration (was: pending) — see Goal 2 above

## Where We're Headed

**v0.4.0 release**: Ship once Docker lands. Tag v0.4.0, update release notes (retry + pg/mysql + watch + docker + design doc for v0.5 Goss preview). Push tags.

**v0.5 theme**: **Competitive move** — Goss migration tool as the headline feature. First Goss-alternative project to offer a migration path. Pair with conditional execution (ROAD-008), multi-env configs (ROAD-017), monorepo recursive (ROAD-010) for a broader v0.5.

**v0.6 forward-look**: File-attribute assertions (mode, owner, content) to close the Goss parity gap that the migration tool currently maps to `command:` fallbacks.

**Feedback signal**: 3 CCS improvements filed during v0.3 session (FB-*); 10 open ideas in roadmap. No blocking patterns — momentum unblocked.

## Priority Order

1. **Docker runner** — small, bounded, Sonnet-dispatchable, unblocks v0.4.0 release
2. **v0.4.0 tag + release notes + push** — fast-follow once Docker merges
3. **Goss brainstorm with user** — decide build-now vs defer-to-v0.5
4. **(If build-now) Goss parser+translator worktree** — dispatch Sonnet A
5. **(If build-now) Goss emitter+CLI worktree** — dispatch Sonnet B, after A merges

## Session Cleanup Context

Uncommitted state entering next session:
- `docs/brainstorming/2026-04-16-goss-migration-tool-design.md` (new, unstaged) — commit separately as `docs(brainstorming): add Goss migration design`
- `docs/prompts/2026-04-16-cosmo-smoke-v0.4-continuation.md` (this file, updated)
- Tracking/intel files (auto-managed)

Stale worktrees cleaned up: `agent-a1e42bae`, `agent-a281a237`, `agent-a700f078`, `agent-aa6c4029`, `agent-ace1a370`, `road-003-watch-mode` all killed. Older `_glm-agent-0001-*` remains from a prior session; untouched.

## Constraints
- Use `ccs prompts complete` to close THIS prompt when Docker+Goss decisions land
- Budget: ~2-3 hours for Docker alone; +6-10h if Goss is built in same session (recommend separate)
- Goal checkboxes: `### [ ]` → `### [x]` when done
- `git push origin master --tags` after v0.4.0 tag
