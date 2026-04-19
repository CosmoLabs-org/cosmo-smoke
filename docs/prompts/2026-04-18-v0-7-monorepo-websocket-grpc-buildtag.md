---
branch: master
completed: "2026-04-18"
created: "2026-04-18"
goals_completed: 8
goals_total: 8
origin: /brainplan
priority: high
related_prompts:
    - docs/brainstorming/2026-04-18-v0-7-monorepo-websocket-grpc-buildtag.md
    - docs/planning-mode/2026-04-18-v0-7-monorepo-websocket-grpc-buildtag.md
status: COMPLETED
tags:
    - continuation
    - implementation
    - v0.7
title: cosmo-smoke v0.7 — Monorepo, WebSocket, gRPC Build Tag
---

# cosmo-smoke v0.7 — Monorepo, WebSocket, gRPC Build Tag

## Context

v0.6 shipped 4 new assertion types (url_reachable, service_reachable, s3_bucket, version_check) plus pre-commit hook integration. v0.7 adds monorepo sub-config discovery (ROAD-010), WebSocket assertion (ROAD-031), and optional gRPC build tag (ROAD-030). Three independent features, no new external deps.

Design spec: `docs/brainstorming/2026-04-18-v0-7-monorepo-websocket-grpc-buildtag.md`
Implementation plan: `docs/planning-mode/2026-04-18-v0-7-monorepo-websocket-grpc-buildtag.md`

## Goals

### [x] 1. Add WebSocket schema struct + validation

Add WebSocketCheck struct to schema.go, wire into Expect, add validation rules for url format and regex.

### [x] 2. Implement WebSocket client + Check function

Create assertion_ws.go with stdlib-only WebSocket client (upgrade, frame parsing, send). Add CheckWebSocket function. Wire into runner.go.

### [x] 3. Write WebSocket tests

Create assertion_ws_test.go with httptest-based test server. 5 tests: expect_contains pass, expect_matches pass, no match fail, connection refused, connect-only.

### [x] 4. Split gRPC code into build-tagged files

Move CheckGRPCHealth + gRPC imports to assertion_grpc.go (`//go:build grpc`). Create assertion_grpc_stub.go (`//go:build !grpc`). Remove gRPC from assertion.go.

### [x] 5. Move gRPC tests and add stub test

Move existing gRPC tests to assertion_grpc_test.go (`//go:build grpc`). Add stub test in assertion_grpc_stub_test.go (`//go:build !grpc`). Verify both build modes.

### [x] 6. Create monorepo discovery package

Create internal/monorepo/discover.go with Discover() function. 5 tests: finds sub-configs, skips ignored dirs, custom exclude, deep nesting, no smoke files.

### [x] 7. Add monorepo schema + CLI flag + runner integration

Add Settings.Monorepo + MonorepoExclude. Add --monorepo flag. Add RunMonorepo to runner. Wire in run.go.

### [x] 8. Update docs + release v0.7.0

Update CLAUDE.md, README.md, USAGE.md, FEATURES.md. Run self-smoke. Bump version, tag.

## Execution Strategy

Chunk 1 (Tasks 1-3) and Chunk 2 (Tasks 4-5) are independent — can run in parallel via GLM agents in separate worktrees. Chunk 3 (Tasks 6-7) is independent of both. Chunk 4 (Task 8) depends on all prior tasks.

agents:
  - task: "WebSocket schema + assertion + tests"
    model: sonnet
    files: [internal/schema/schema.go, internal/schema/validate.go, internal/runner/assertion_ws.go, internal/runner/assertion_ws_test.go, internal/runner/runner.go]
    ready: true
  - task: "gRPC build tag split + tests"
    model: sonnet
    files: [internal/runner/assertion.go, internal/runner/assertion_grpc.go, internal/runner/assertion_grpc_stub.go, internal/runner/assertion_grpc_test.go, internal/runner/assertion_grpc_stub_test.go]
    ready: true
  - task: "Monorepo discovery + runner integration"
    model: sonnet
    files: [internal/monorepo/discover.go, internal/monorepo/discover_test.go, internal/schema/schema.go, internal/runner/runner.go, cmd/run.go]
    ready: true

## File Scope

```yaml
files_modified:
  - internal/schema/schema.go
  - internal/schema/validate.go
  - internal/runner/runner.go
  - internal/runner/assertion.go
  - cmd/run.go
  - CLAUDE.md
files_created:
  - internal/runner/assertion_ws.go
  - internal/runner/assertion_ws_test.go
  - internal/runner/assertion_grpc.go
  - internal/runner/assertion_grpc_stub.go
  - internal/runner/assertion_grpc_test.go
  - internal/runner/assertion_grpc_stub_test.go
  - internal/monorepo/discover.go
  - internal/monorepo/discover_test.go
```
