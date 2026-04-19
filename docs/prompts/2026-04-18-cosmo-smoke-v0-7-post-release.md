---
branch: master
completed: "2026-04-19"
created: "2026-04-18"
goals_completed: 5
goals_total: 5
origin: /continuation-prompt
priority: high
related_prompts:
    - docs/prompts/2026-04-18-v0-7-monorepo-websocket-grpc-buildtag.md
status: COMPLETED
tags:
    - continuation
    - post-release
    - v0.7
title: cosmo-smoke v0.7 Post-Release — Tag, Creds, GraphQL, Doc Gaps
---

# cosmo-smoke v0.7 Post-Release — Tag, Creds, GraphQL, Doc Gaps

## File Scope

```yaml
files_modified: []
files_created: []
```

Note: File scope will be populated once work begins. Expected areas:
- `internal/runner/` — credential smoke tests, GraphQL assertion
- `internal/schema/` — new schema types
- `cmd/` — any CLI flag additions
- `docs/` — READMEs, USAGE, CLAUDE.md updates

## Context

v0.7 just shipped with three solid features (WebSocket assertion, monorepo discovery, gRPC build tag) plus the optional run field and a clean assertion.go refactor into domain files. 253 tests all pass. The codebase is in good shape — clean architecture, pure assertion functions, minimal deps. The changelog is staged in `docs/changelog/unreleased.yaml` but v0.7.0 is not yet tagged. The project is in a "feature-complete plateau" for v0.7 — next moves are a mix of release hygiene, quick wins, and one medium design task (GraphQL).

## GLM Dispatch Rules

When goals involve dispatching subagents:

1. **ALWAYS** use `ccs glm-agent exec` for GLM agents (routes through queue with retry logic)
2. **NEVER** use Agent tool with `model:sonnet` or `model:haiku` for GLM work (bypasses queue, risks 429 rate limits)
3. Agent tool with `model:opus` is fine for Opus subagents
4. For parallel work: use `/glm-sprint` or `ccs glm-agent exec-batch`

## What Got Done

**v0.7 full implementation session** — completed all 8 goals from the v0.7 plan:

1. WebSocket schema + validation (schema.go)
2. WebSocket client + assertion (assertion_ws.go) — stdlib-only, no gorilla/websocket dep
3. WebSocket tests (assertion_ws_test.go) — 5 tests via httptest
4. gRPC build tag split — assertion_grpc.go (+build grpc) / assertion_grpc_stub.go (+build !grpc)
5. gRPC test migration + stub test
6. Monorepo discovery package (internal/monorepo/discover.go) — finds sub-config .smoke.yaml files
7. Monorepo CLI flag + runner integration (--monorepo flag, RunMonorepo)
8. Docs update — CLAUDE.md, README, USAGE, FEATURES updated
9. assertion.go refactored into per-domain files (assertion_net.go, assertion_docker.go, assertion_db.go, assertion_http.go)
10. Run field made optional for network-only tests

Key commits: `089eac6` (features), `7a53278` (docs/metadata), `d02d1c5` (session cleanup).

Test count: 253 passed across 8 packages. All green.

## Goals

### [x] 1. Tag and release v0.7.0

The changelog is staged in `docs/changelog/unreleased.yaml` with all v0.7 entries (WebSocket, monorepo, gRPC build tag, optional run field, pre-commit hook, assertion refactor). Need to:

- Run `ccs changelog finalize 0.7.0 "websocket-monorepo-grpc-buildtag"` to finalize the changelog
- Verify the binary builds: `go build -ldflags "-s -w -X github.com/CosmoLabs-org/cosmo-smoke/cmd.Version=0.7.0" -o smoke .`
- Tag: `git tag v0.7.0 -m "v0.7.0: WebSocket, monorepo discovery, gRPC build tag"`
- Push when ready: `git push origin master --tags`

### [x] 2. Credential smoke tests (ROAD-014)

Quick win. Add a `credential_check` assertion type that verifies credentials/secrets are accessible without leaking them. Design questions:

- What credential sources? Environment variables, files (like .env, kubeconfig), cloud IAM?
- Should this just check existence or also validate (e.g., test a DB connection with found creds)?
- Acceptance: schema struct, validation, assertion function, tests, CLAUDE.md update

Start with env + file credential checks. Keep it simple — this is a quick win, not a full secrets manager integration.

### [x] 3. GraphQL introspection assertion (FEAT-008)

Medium task. The issue exists at `docs/issues/FEAT-008.yaml` (currently has a stale description referencing S3 — update it). Design needed:

- Schema: `graphql_introspect` check with `{url, query?, expect_types?, expect_mutation?}` fields
- Implementation: HTTP POST to the GraphQL endpoint with introspection query, parse response
- Decide: full schema introspection or targeted type/query checks?
- Acceptance: schema struct, validation, assertion function, tests, CLAUDE.md update

This follows the same pattern as the HTTP and WebSocket assertions — should be straightforward once the API surface is decided.

### [x] 4. Fill command doc gaps

5 commands are missing dedicated README/USAGE docs:
- `cmd/init_cmd.go` — `smoke init`
- `cmd/migrate.go` — `smoke migrate`
- `cmd/run.go` — `smoke run`
- `cmd/serve.go` — `smoke serve`
- `cmd/version.go` — `smoke version`

Create `docs/commands/` directory with one doc per command covering: description, usage, flags, examples. This is a good GLM batch task since each doc is independent and bounded.

### [x] 5. Roadmap grooming — update ROAD-031 and ROAD-030 status

ROAD-031 (WebSocket assertion) and ROAD-030 (gRPC build tag) are still marked `captured` in the roadmap but are now implemented in v0.7. Update their status to `completed` with `completed_in: v0.7.0`. Also review other `captured` items for any that were implemented but not updated.

## Carry-Over Tasks

None — this session completed all planned v0.7 work cleanly.

## Carry-Overs

None — the previous v0.7 prompt (`docs/prompts/2026-04-18-v0-7-monorepo-websocket-grpc-buildtag.md`) was fully completed (8/8 goals).

## Where We're Headed

v0.7 is the last "incremental assertion types" release for a while. The project has 22 assertion types now — the diminishing returns curve is flattening. The next strategic pivot points are:

1. **ROAD-032: Claude Desktop MCP extension** — This is the big unlock. An MCP server that lets Claude Desktop generate and run smoke tests for any project. If we nail this, every CosmoLabs project gets instant smoke test coverage. Large scope, needs its own design session.

2. **Portfolio dashboard** — One of the seed ideas is a dashboard showing smoke test results across the ~95-project portfolio. This would need a `smoke serve` endpoint (already implemented) feeding into a central aggregator.

3. **OTel tracing** (seed idea) — Trace correlation with OpenTelemetry. More of a v0.9+ feature once the basics are rock-solid.

The credential smoke tests (ROAD-014) and GraphQL (FEAT-008) are the last two assertion-type features worth doing before pivoting to the MCP extension. Get these done, tag v0.7.1 or v0.8.0, then focus on ROAD-032.

## Priority Order

1. **Tag v0.7.0** — release hygiene, unblocks downstream consumers
2. **Roadmap grooming** (Goal 5) — 5 minutes, keeps roadmap accurate
3. **Credential smoke tests** (Goal 2) — quick win, small scope, good momentum builder
4. **Command doc gaps** (Goal 4) — good GLM batch task, can run in parallel with other work
5. **GraphQL introspection** (Goal 3) — needs design discussion before implementation
