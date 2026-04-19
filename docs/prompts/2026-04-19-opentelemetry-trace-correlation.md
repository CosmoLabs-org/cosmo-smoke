---
branch: master
completed: "2026-04-19"
created: "2026-04-19"
goals_completed: 10
goals_total: 10
origin: /brainplan
priority: high
related_prompts:
    - docs/brainstorming/2026-04-18-opentelemetry-trace-correlation.md
    - docs/planning-mode/2026-04-19-opentelemetry-trace-correlation.md
status: COMPLETED
tags:
    - continuation
    - implementation
    - opentelemetry
    - tracing
title: OpenTelemetry Trace Correlation — Full Implementation
---

# OpenTelemetry Trace Correlation — Full Implementation

## Context

Add W3C trace context propagation to cosmo-smoke's network assertions (HTTP, gRPC, WebSocket) and a Jaeger-based trace verification assertion. This makes smoke tests visible in observability pipelines and enables end-to-end trace pipeline validation.

Design spec: `docs/brainstorming/2026-04-18-opentelemetry-trace-correlation.md`
Implementation plan: `docs/planning-mode/2026-04-19-opentelemetry-trace-correlation.md`

## Goals

- [x] Task 1: Add OTelConfig + OTelTraceCheck to schema
- [x] Task 2: Add validation for otel config fields
- [x] Task 3: Implement TraceContext + SpanContext
- [x] Task 4: Implement CheckOTelTrace with Jaeger polling
- [x] Task 5: Inject traceparent into HTTP requests
- [x] Task 6: Inject traceparent into WebSocket handshake
- [x] Task 7: Inject traceparent into gRPC metadata
- [x] Task 8: Wire trace context into runner pipeline
- [x] Task 9: Add --otel-collector and --no-otel CLI flags
- [x] Task 10: Update CLAUDE.md and self-smoke config

## Execution Strategy

Chunk 1 (Tasks 1-2) and Chunk 2 (Tasks 3-4) are independent — can run in parallel.
Chunk 3 (Tasks 5-7) depends on Chunks 1 and 2.
Chunk 4 (Task 8) depends on all prior chunks.
Chunk 5 (Tasks 9-10) depends on Chunk 4.

Recommended: GLM dispatch for independent chunks, Opus for runner integration.

    agents:
      - task: "Schema + validation (Tasks 1-2)"
        model: sonnet
        files: [internal/schema/schema.go, internal/schema/validate.go, internal/schema/schema_test.go, internal/schema/validate_test.go]
        ready: true
      - task: "Trace context + otel assertion (Tasks 3-4)"
        model: sonnet
        files: [internal/runner/trace.go, internal/runner/trace_test.go, internal/runner/assertion_otel.go, internal/runner/assertion_otel_test.go]
        ready: true
      - task: "Network assertion injection (Tasks 5-7)"
        model: sonnet
        files: [internal/runner/assertion_network.go, internal/runner/assertion_ws.go, internal/runner/assertion_grpc.go, internal/runner/assertion_grpc_stub.go]
        ready: false  # depends on schema + trace context
      - task: "Runner integration (Task 8)"
        model: opus
        files: [internal/runner/runner.go]
        ready: false  # depends on all above
      - task: "CLI + docs (Tasks 9-10)"
        model: sonnet
        files: [cmd/run.go, CLAUDE.md]
        ready: false  # depends on runner integration

## File Scope

### New files
- `internal/runner/trace.go`
- `internal/runner/trace_test.go`
- `internal/runner/assertion_otel.go`
- `internal/runner/assertion_otel_test.go`

### Modified files
- `internal/schema/schema.go`
- `internal/schema/validate.go`
- `internal/runner/runner.go`
- `internal/runner/assertion_network.go`
- `internal/runner/assertion_ws.go`
- `internal/runner/assertion_grpc.go`
- `internal/runner/assertion_grpc_stub.go`
- `cmd/run.go`
- `CLAUDE.md`
