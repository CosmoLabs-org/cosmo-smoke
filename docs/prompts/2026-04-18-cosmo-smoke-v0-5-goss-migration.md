---
title: "cosmo-smoke v0.5 — Goss migration tool + complementary features"
created: 2026-04-18
status: PENDING
priority: high
branch: master
origin: "/continuation-prompt"
tags: [continuation]
goals_total: 5
goals_completed: 0
carried_over_from: null
carried_over_goals: 0
related_prompts:
  - docs/prompts/2026-04-16-cosmo-smoke-v0.4-continuation.md
---

# cosmo-smoke v0.5 — Goss Migration Tool - Continuation Prompt

## File Scope

```yaml
files_modified:
  - cmd/root.go
files_created:
  - cmd/migrate.go
  - internal/migrate/goss/parser.go
  - internal/migrate/goss/translator.go
  - internal/migrate/goss/emitter.go
  - internal/migrate/goss/testdata/
```

Files for complementary features (scope-dependent):

```yaml
files_created_maybe:
  - internal/runner/conditional.go
  - internal/schema/conditional.go
  - internal/schema/multienv.go
  - cmd/run.go (--env flag additions)
```

## Context

v0.4.0 shipped cleanly today — retry, postgres/mysql ping, watch mode, and docker assertions all landed. Master is tagged, green, and the feedback inbox is empty. The project has 198 passing tests and a clean worktree. v0.5 is fully greenfield.

The strategic bet for v0.5 is the Goss migration tool (ROAD-024, priority 85). Goss is effectively abandoned upstream; cosmo-smoke is the natural alternative, and `smoke migrate goss` is the migration path that lowers friction to near zero. It's a competitive headline — no other Goss-alternative ships a one-command migration path. The design doc is complete at `docs/brainstorming/2026-04-16-goss-migration-tool-design.md`. Before building, 5 open questions need answers (see Goal 1). After that, the recommended execution model is two parallel Sonnet worktrees.

## GLM Dispatch Rules

When goals involve dispatching subagents:

1. **ALWAYS** use `ccs glm-agent exec` for GLM agents (routes through queue with retry logic)
2. **NEVER** use Agent tool with `model:sonnet` or `model:haiku` for GLM work (bypasses queue, risks 429 rate limits)
3. Agent tool with `model:opus` is fine for Opus subagents
4. For parallel work: use `/glm-sprint` or `ccs glm-agent exec-batch`

## What Got Done

v0.4.0 shipped with four headline features across this session and the previous:

- **Retry** — opt-in per-test `retry: {count, backoff}` with exponential backoff
- **Postgres + MySQL ping** — new protocol-level assertions (`postgres_ping`, `mysql_ping`)
- **Watch mode** — `--watch` flag, fsnotify-backed with 500ms debounce
- **Docker assertions** — `docker_container_running` and `docker_image_exists`

The session also handled feedback triage: FB-511 filed against CosmoHooks for a health false-positive duplicate warning, and FB-506–510 forwarded to ClaudeCodeSetup. Stale worktrees cleaned, session metadata updated, v0.4 changelog finalized. The v0.4 continuation prompt is now COMPLETED.

## Goals

### [ ] 1. Resolve the 5 open design questions for ROAD-024 (Goss migration)

Before any code is written, the 5 open questions from the design doc need answers:

1. **Multi-distro packages:** v1 hardcodes `dpkg`. Add `--distro=deb|rpm|apk` flag, or detect from env/hint?
2. **Lossy vs strict:** Is `command:` fallback with a TODO comment acceptable for v1? (Design doc recommends yes — confirm.)
3. **Output naming:** Flatten includes into one `.smoke.yaml`, or preserve `includes:` and emit sibling files?
4. **Reverse migration (`smoke-to-goss`):** Defer to v0.6? (Design doc says defer — confirm.)
5. **Bulk migration (`smoke migrate goss <dir>`):** Per-file only for v0.5, directory support later?

Run `/brainstorm` or discuss directly. Once answered, update the design doc at `docs/brainstorming/2026-04-16-goss-migration-tool-design.md` with the resolved decisions and save an implementation plan to `docs/planning-mode/2026-04-18-goss-migration-tool.md`.

Design doc is at: `docs/brainstorming/2026-04-16-goss-migration-tool-design.md`

### [ ] 2. Implement ROAD-024 — Goss migration tool (parallel worktrees)

After design questions are resolved, execute per the design doc's dispatch recommendation:

**Worktree A — parser + translator** (Sonnet):
- `internal/migrate/goss/parser.go` — `GossFile` struct, YAML parsing
- `internal/migrate/goss/translator.go` — one `translate*` function per Goss key, returns `([]schema.Test, []TranslationWarning)`
- Unit tests with golden-file fixtures in `internal/migrate/goss/testdata/`
- Scope: core 7 keys (package, service, process, port, command, file, http) per the design doc

**Worktree B — emitter + CLI** (Sonnet):
- `internal/migrate/goss/emitter.go` — Option A string-builder emitter (preserves TODO comments)
- `cmd/migrate.go` — `smoke migrate goss <input.yaml>` cobra subcommand, wired into rootCmd
- Flags: `--output`, `--overwrite`, `--strict`, `--stats`
- Non-zero exit on `--strict` when any skip/fallback happened

**Acceptance criteria:**
- `smoke migrate goss testdata/goss/example.yaml` produces valid `.smoke.yaml`
- `schema.Load` parses the emitted file without errors
- `--stats` prints mapping summary to stderr
- All 7 core Goss keys translate correctly (golden-file tests pass)
- Long-tail keys (interface, mount, kernel-param, dns, addr) emit `# TODO:` stubs, not panics

**Merge order:** parser+translator first (B depends on translator types), then emitter+CLI.

Use `ccs spawn` for worktrees. Never `git worktree add` directly. After each merges: `ccs verify-worktree NAME --fix --approve && ccs merge NAME && ccs kill NAME`.

### [ ] 3. Implement ROAD-008 — Conditional test execution (small)

Complementary to Goss migration: Goss has `skip: true`; cosmo-smoke currently has no skip. This resolves the TODO emitted for Goss's `skip` flag.

**Scope:** Add `skip_if:` field to `schema.Test`. Conditions: `env_unset: VAR`, `env_equals: {var: X, value: Y}`, `file_missing: path`. Runner skips test if condition evaluates true, marks it SKIPPED in reporter (not FAIL).

Files to touch: `internal/schema/` (add `SkipIf` field), `internal/runner/` (evaluate skip condition before execution), `internal/reporter/` (handle SKIPPED status in terminal + JSON output).

Issue: ROAD-008. Check `ccs roadmap show ROAD-008` for any additional spec before implementing.

### [ ] 4. Implement ROAD-017 — Multi-environment smoke configs (small)

Low-overhead v0.5 inclusion. Allows `smoke run --env staging` to load `staging.smoke.yaml` overrides merged onto base `.smoke.yaml`.

**Scope:** Add `--env` flag to `smoke run`. When set, load `<env>.smoke.yaml` from same dir as `.smoke.yaml`, deep-merge over base config (env-specific tests extend, don't replace base tests). Document merge semantics clearly.

Files to touch: `cmd/run.go` (`--env` flag), `internal/schema/` (merge logic), `internal/runner/` (pass env context).

Issue: ROAD-017.

### [ ] 5. Release v0.5.0

Once ROAD-024 is merged (and optionally ROAD-008/017):

```bash
# Update version
ccs version-track bump minor --notes "Goss migration tool, conditional execution, multi-env configs"

# Changelog
ccs changelog add feat "smoke migrate goss: one-command Goss → cosmo-smoke migration (ROAD-024)"
ccs changelog add feat "Conditional test execution via skip_if: (ROAD-008)"
ccs changelog add feat "Multi-environment configs via --env flag (ROAD-017)"

# Tag and push
git tag v0.5.0
git push origin master --tags
```

Run `smoke run` self-smoke before tagging. All 198+ tests must pass.

## Where We're Headed

v0.5 is the "Goss migration tool" release — cosmo-smoke becomes the first Goss alternative with a first-class migration path. This is the competitive moat moment. Once `smoke migrate goss` ships, the pitch to Goss users writes itself: "Goss is unmaintained. cosmo-smoke is the drop-in replacement, and we'll migrate your existing tests in one command."

After v0.5, the natural v0.6 candidates are ROAD-010 (monorepo sub-config) and ROAD-018 (service dependency checks) — both of which expand the addressable project types. ROAD-013 (dependency version assertions) and ROAD-014 (credential smoke tests) are small, shippable in a single session alongside a larger feature.

The long-term arc: each release expands a different dimension of coverage — assertions (v0.4), migration/ecosystem (v0.5), project scale (v0.6). Keep that pattern; it gives each release a coherent story.

## Priority Order

1. **Goal 1** (design questions) — blocks all implementation work
2. **Goal 2** (ROAD-024, Goss migration) — the headline feature, blocks v0.5 release
3. **Goal 3** (ROAD-008, conditional) — resolves the Goss `skip:` TODO, completes the Goss story
4. **Goal 4** (ROAD-017, multi-env) — quick win, high user value
5. **Goal 5** (release) — execute once Goals 2-4 are verified
