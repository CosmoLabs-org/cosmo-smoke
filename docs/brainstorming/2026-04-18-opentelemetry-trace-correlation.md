---
title: "OpenTelemetry Trace Correlation for cosmo-smoke"
created: "2026-04-18"
status: DRAFT
idea_ref: IDEA-MO1FBPJB
tags: [opentelemetry, tracing, observability, jaeger]
---

# OpenTelemetry Trace Correlation

## Problem

Smoke tests verify infrastructure health but are invisible to observability pipelines. When a smoke test fails, there's no trace in Jaeger/Tempo to correlate with. Teams can't answer: "did our smoke test traffic actually generate traces?" or "is the trace pipeline working end-to-end?"

## Solution

Add W3C trace context propagation to cosmo-smoke's network assertions (HTTP, gRPC, WebSocket) and a dedicated `otel_trace` assertion that verifies traces arrive at a Jaeger-compatible collector.

- Smoke generates its own trace context (no OTel SDK dependency)
- Traceparent headers auto-injected into all network assertions when `otel` config is enabled
- `otel_trace` assertion queries Jaeger API to confirm trace reception
- Zero overhead when disabled

## Configuration Schema

### Global `otel` block

```yaml
otel:
  enabled: true
  jaeger_url: "http://jaeger:16686"   # Jaeger query API (not OTLP ingest)
  service_name: "cosmo-smoke"         # default: "smoke"
  trace_propagation: true             # auto-inject traceparent into network assertions
```

### Per-test `otel_trace` assertion

```yaml
tests:
  - name: trace arrived at collector
    run: curl http://localhost:8080/health
    expect:
      otel_trace:
        jaeger_url: "http://jaeger:16686"      # overrides global
        service_name: "my-service"             # overrides global
        min_spans: 1                           # at least N spans received (default: 1)
        timeout: 5s                            # wait for trace to appear
```

## Trace Context Model

### W3C Traceparent Format

All propagation uses the W3C `traceparent` header: `00-{traceID}-{spanID}-{flags}`

- `traceID`: 32 hex chars (16 bytes), generated once per suite run
- `spanID`: 16 hex chars (8 bytes), generated per test
- `flags`: `01` (sampled)

No OTel SDK dependency. Pure string construction from crypto/rand. The `tracestate` header is intentionally omitted — cosmo-smoke doesn't need vendor-specific trace context.

### Hierarchy

```
Suite (traceID)
├── Test 1 (spanID-1, traceparent injected into HTTP/gRPC/WS requests)
├── Test 2 (spanID-2, ...)
└── Test N (spanID-N, ...)
```

Each test is a child span under the suite's trace. The `otel_trace` assertion queries by the suite's `traceID`.

### Injection Points

| Assertion Type | Injection Target |
|---------------|-----------------|
| `http` | `traceparent` HTTP request header |
| `grpc_health` | `traceparent` gRPC metadata key |
| `websocket` | `traceparent` in WebSocket handshake headers |

When `otel` config is absent or `enabled: false`, no injection occurs. Zero overhead.

## Trace Verification

### Jaeger API Query

The `otel_trace` assertion queries Jaeger's trace API:

```
GET {jaeger_url}/api/traces/{traceID}?service={service_name}
```

### Assertion Logic

1. Build the Jaeger API URL from `jaeger_url` + `traceID` + `service_name`
2. Poll with 500ms interval until `timeout` expires
3. Parse Jaeger JSON response (`data[].spans[]`)
4. Assert `len(spans) >= min_spans`
5. Return `AssertionResult` with pass/fail

### Error Cases

| Condition | Result |
|-----------|--------|
| Collector unreachable | Fail with network error |
| Timeout with no spans | Fail: "no spans found for trace {id} within {timeout}" |
| Invalid collector URL | Fail: validation error at config parse time |

### Jaeger Response Parsing

Expected JSON structure:

```json
{
  "data": [
    {
      "traceID": "abc123...",
      "spans": [
        { "operationName": "...", "spanID": "..." }
      ]
    }
  ]
}
```

Assertion checks `data[0].spans` length against `min_spans`.

## Architecture

### New Files

| File | Purpose |
|------|---------|
| `internal/runner/trace.go` | `traceContext` struct, W3C traceparent generation, child span creation |
| `internal/runner/assertion_otel.go` | `CheckOTelTrace` function + Jaeger HTTP client |
| `internal/runner/trace_test.go` | Trace context generation unit tests |
| `internal/runner/assertion_otel_test.go` | Verification tests with net/http/httptest mock Jaeger |

### Modified Files

| File | Change |
|------|--------|
| `internal/schema/schema.go` | Add `OTelConfig` struct at top level + `OTelTraceCheck` field on `Expect` |
| `internal/schema/validate.go` | Validate otel config: jaeger_url format, service_name required when enabled |
| `internal/runner/runner.go` | Initialize `traceContext` in `Run()`, pass to assertion execution |
| `internal/runner/assertion_network.go` | Inject `traceparent` header into HTTP requests when trace context available |
| `internal/runner/assertion_grpc.go` | Inject `traceparent` into gRPC metadata when trace context available |
| `internal/runner/assertion_ws.go` | Inject `traceparent` into WebSocket handshake headers when trace context available |
| `cmd/run.go` | Add `--otel-collector` CLI flag (override `otel.jaeger_url`) and `--no-otel` flag (disable at runtime) |

### Data Flow

```
smoke run
  └── Runner.Run(opts)
       ├── Parse OTelConfig from SmokeConfig
       ├── If otel.enabled: init traceContext (generate traceID)
       ├── For each test:
       │    ├── Create child spanID
       │    ├── For HTTP/gRPC/WS assertions:
       │    │    └── Inject traceparent into request BEFORE sending
       │    │        (at request-construction time inside each Check* function)
       │    ├── Run test command
       │    ├── For otel_trace assertion:
       │    │    └── Poll GET {jaeger}/api/traces/{traceID}?service={svc}
       │    └── Collect AssertionResult
       └── Return SuiteResult
```

### Impure Assertion Note

All 26 existing assertions are pure functions — no I/O beyond test command output. `otel_trace` departs from this by polling an external HTTP API. This is intentional: trace verification is inherently an external system check (like `port_listening` or `redis_ping`). The function signature accepts an `*http.Client` parameter for testability — tests inject an `httptest` mock, production uses the default client.

### No New Dependencies

All stdlib: `crypto/rand`, `encoding/hex`, `net/http`, `encoding/json`, `fmt`. W3C traceparent is a simple string format.

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| No OTel SDK | Keeps "minimal deps" principle. W3C traceparent is trivially constructed. cosmo-smoke is a test runner, not a telemetry pipeline. |
| Jaeger API only (v1) | Most widely deployed. Tempo is Jaeger-compatible. Can add backends later. |
| Suite-level traceID | Correlates all tests from one `smoke run` invocation. Matches how teams use traces. |
| Per-test spanID | Allows identifying which test generated which downstream span. |
| Polling for verification | Traces take time to propagate through collector pipeline. 500ms interval balances latency vs load. |
| `min_spans` threshold | Default 1 means "any trace arrived." Configurable for stricter checks. |
| Override per assertion | `otel_trace` allows per-test `jaeger_url` and `service_name` for multi-service scenarios. |

## Edge Cases

- **No `otel` config**: No trace context created, no injection, no overhead
- **`otel` enabled but no network assertions**: Trace context created but unused. `otel_trace` assertion still works standalone.
- **Multiple `otel_trace` assertions**: Each queries independently with its own timeout/collector config
- **Collector returns empty data**: Retries until timeout, then fails
- **W3C traceparent format compatibility**: Strict `00` version prefix, lowercase hex, 32+16+2 char parts
- **`allow_failure` interaction**: Trace context creation and injection are independent of assertion pass/fail. A test with `allow_failure: true` still gets a spanID and traceparent injection — the `otel_trace` assertion may fail without affecting suite outcome.

## Future Considerations

- Support additional backends (Tempo native API, Honeycomb, Datadog)
- Export smoke results as OTel spans (reverse direction)
- Trace-aware retries: only retry if trace was never received
- Integration with `--watch` mode: continuous trace health monitoring
