---
completed: "2026-04-18"
created: "2026-04-18"
status: COMPLETED
title: cosmo-smoke v0.7 — Monorepo, WebSocket, gRPC Build Tag
---

# cosmo-smoke v0.7 — Monorepo, WebSocket, gRPC Build Tag

## Why

v0.6 expanded assertion coverage with 4 new connectivity checks. v0.7 addresses three gaps:

1. **Monorepo support (ROAD-010)** — CosmoLabs has ~95 projects. Many live in monorepos. Currently users must `cd` into each subdirectory or write a wrapper script. Auto-discovery makes `smoke run --monorepo` work from the repo root.

2. **WebSocket assertion (ROAD-031)** — Real-time apps use WebSocket extensively. A connect-send-expect assertion fills the gap between HTTP checks and custom shell scripts.

3. **Optional gRPC module (ROAD-030)** — The gRPC health check adds 121MB of dependencies and inflates the binary from ~8MB to 13.9MB. Most users don't need gRPC. An opt-in build tag lets them skip it.

## Design Decisions

1. **Monorepo is explicit opt-in.** `--monorepo` flag or `monorepo: true` in root config. No auto-detection — avoids accidental scans of large non-monorepo directories.
2. **Unlimited discovery depth.** Real monorepos have varied layouts (`services/team-a/api/`). Depth limits would be arbitrary.
3. **WebSocket via stdlib only.** The connect-send-expect pattern only needs HTTP upgrade + frame parsing. No external WS library. ~80 lines.
4. **gRPC is opt-in via `-tags grpc`.** Default build excludes gRPC. Users who need it add the tag. No behavior change for existing users.
5. **No new external dependencies.** All three features use stdlib only. The "minimal deps" principle is maintained.

## Feature 1: Monorepo Sub-Config Discovery

### Behavior

- `smoke run --monorepo` triggers discovery from cwd (or config file dir)
- `monorepo: true` in root `.smoke.yaml` settings also enables it
- CLI flag takes precedence over config
- Walks all subdirectories (unlimited depth), finds `.smoke.yaml` files
- Skips: `.git/`, `node_modules/`, `vendor/`, `__pycache__/`, `dist/`, `build/`, `target/`
- Additional skip dirs configurable via `monorepo_exclude` in settings
- Each discovered config runs as a sub-suite with its own project name
- Reporter shows per-project results + aggregate summary
- Exit code 1 if any sub-suite has failures
- `--tag`/`--exclude-tag`/`--fail-fast` apply across all sub-suites
- `--dry-run` lists all discovered configs without running
- `--watch` + `--monorepo`: watches all subdirectories for changes and re-runs the full suite. Re-discovers on each run (new configs are picked up).
- Root `.smoke.yaml` is optional — can use `--monorepo` flag alone
- If no `.smoke.yaml` files are found, returns error: "no smoke configs found in <dir>"

### Schema Changes

```yaml
# Root .smoke.yaml (optional)
version: 1
project: my-monorepo
settings:
  monorepo: true
  monorepo_exclude:
    - "internal/"
    - "examples/"
```

New fields on `Settings`:
```go
Monorepo         bool     `yaml:"monorepo,omitempty"`
MonorepoExclude  []string `yaml:"monorepo_exclude,omitempty"`
```

### Implementation

New `internal/monorepo/` package:
- `Discover(root string, exclude []string) ([]SubConfig, error)` — walks directory tree
- `SubConfig` struct: `{Path, Project string}`
- Skips common non-project dirs by default

Runner changes:
- New `RunMonorepo(opts RunOptions) (*SuiteResult, error)` method
- Runs each sub-config as its own `Runner` instance
- Aggregates results into a single `SuiteResult`
- Reporter shows project-level grouping

CLI changes:
- `--monorepo` flag on `smoke run`
- `run.go`: check flag + config, call `RunMonorepo` when active

### Output Example

```
  ● api: Compiles               ✓
  ● api: Tests pass              ✓
  ● api: Health endpoint         ✓
    api: 3 passed  (2.1s)

  ● worker: Builds              ✓
  ● worker: Connects to Redis    ✓
    worker: 2 passed  (1.8s)

  ● auth: Builds                ✓
  ● auth: JWT validation         ✓
    auth: 2 passed  (1.2s)

  Total: 3 projects, 7 passed  (5.1s)
```

### Testing

| Test | Description |
|------|-------------|
| TestDiscover_FindsSubConfigs | Create temp dir with 2 subdirs each having `.smoke.yaml`, verify both found |
| TestDiscover_SkipsIgnoredDirs | `node_modules/` with `.smoke.yaml` is skipped |
| TestDiscover_CustomExclude | `monorepo_exclude` dirs are skipped |
| TestDiscover_DeepNesting | Finds config 3+ levels deep |
| TestDiscover_NoSmokeFiles | Directory tree with zero `.smoke.yaml` returns empty slice, runner returns error |
| TestRunMonorepo_AllPass | Run monorepo suite, all pass |
| TestRunMonorepo_PartialFail | One sub-suite fails, aggregate exit code is 1 |
| TestRunMonorepo_TagFilter | `--tag build` runs only tagged tests across all sub-suites |
| TestRunMonorepo_DryRun | Lists discovered configs without running |

## Feature 2: WebSocket Assertion

### Schema

```yaml
expect:
  websocket:
    url: "ws://localhost:8080/ws"      # ws:// or wss:// (required)
    send: '{"type": "ping"}'           # optional message to send after connect
    expect_contains: "pong"            # substring match on received message
    expect_matches: "connected.*true"  # regex match (alternative to contains)
    timeout: 5s                        # connection + response timeout (default 10s)
```

Two flows are supported:

1. **Connect-send-expect** (both `send` and `expect_*` set): Connect → send message → wait for response → match.
2. **Connect-and-wait** (`send` empty, `expect_*` set): Connect → wait for server to send an unsolicited message → match.
3. **Connect-only** (neither `send` nor `expect_*` set): Just verify the WebSocket handshake succeeds. Passes if upgrade response is 101.

New struct:
```go
type WebSocketCheck struct {
    URL            string   `yaml:"url"`
    Send           string   `yaml:"send,omitempty"`
    ExpectContains string   `yaml:"expect_contains,omitempty"`
    ExpectMatches  string   `yaml:"expect_matches,omitempty"`
    Timeout        Duration `yaml:"timeout,omitempty"` // default 10s
}
```

New field on `Expect`:
```go
WebSocket *WebSocketCheck `yaml:"websocket,omitempty"`
```

### Validation

- `url` is required, must start with `ws://` or `wss://`
- If `expect_matches` is set, must compile as valid Go regex
- It is valid to omit both `expect_contains` and `expect_matches` (connect-only mode)

### Implementation

Pure stdlib WebSocket client (~100 lines):
- HTTP GET with `Upgrade: websocket`, `Connection: Upgrade` headers
- `Sec-WebSocket-Key` via `crypto/rand` + `base64`
- `Sec-WebSocket-Accept` verification via `crypto/sha1`
- Frame parsing: opcode 1 (text), opcode 2 (binary — read payload but match as text), opcode 8 (close), opcode 9 (ping/pong)
- Masking for client-to-server frames (required by RFC 6455)
- Default timeout: 10s (consistent with HTTPCheck default)

New `CheckWebSocket` function in `internal/runner/assertion_ws.go`.

Error handling:
- Connection refused → "connection failed: ..."
- Timeout → "timed out after Xs waiting for message"
- Close frame → "server closed connection: <reason>"
- No match → "received <message> did not contain/match <pattern>"

### Testing

Test server: `httptest.NewServer` with a custom handler that performs the WebSocket upgrade handshake using the same stdlib frame logic (upgrade handler in test helper file, ~40 lines).

| Test | Description |
|------|-------------|
| TestCheckWebSocket_ExpectContains_Pass | Server echoes, send "ping", expect "ping" |
| TestCheckWebSocket_ExpectMatches_Pass | Server sends JSON, regex matches |
| TestCheckWebSocket_NoMatch_Fail | Server sends "hello", expect "pong" |
| TestCheckWebSocket_ConnectionRefused | Connect to unused port, verify graceful failure |
| TestCheckWebSocket_ConnectOnly | No send/expect, verify handshake success |
| TestCheckWebSocket_BinaryFrame | Server sends binary frame, assertion reads it as text |

### Testing

| Test | Description |
|------|-------------|
| TestCheckWebSocket_ExpectContains_Pass | Server echoes, send "ping", expect "ping" |
| TestCheckWebSocket_ExpectMatches_Pass | Server sends JSON, regex matches |
| TestCheckWebSocket_NoMatch_Fail | Server sends "hello", expect "pong" |
| TestCheckWebSocket_ConnectionRefused | Connect to unused port, verify graceful failure |

## Feature 3: Optional gRPC Module via Build Tag

### Approach

Opt-in: `-tags grpc`. Default build excludes gRPC deps entirely.

### File Changes

```
internal/runner/
  assertion_grpc.go          # //go:build grpc — current gRPC code, moved from assertion.go
  assertion_grpc_stub.go     # //go:build !grpc — stub returning "not available"
```

`assertion_grpc.go` contains:
- `CheckGRPCHealth` function (moved from `assertion.go`)
- gRPC-related imports (`google.golang.org/grpc`, healthpb, etc.)

`assertion_grpc_stub.go` contains:
```go
//go:build !grpc

package runner

import (
    "fmt"

    "github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

func CheckGRPCHealth(check *schema.GRPCHealthCheck) AssertionResult {
    return AssertionResult{
        Type:     "grpc_health",
        Expected: check.Address,
        Actual:   "grpc_health not available — rebuild with -tags grpc",
        Passed:   false,
    }
}
```

### Impact

| Build | Command | Binary Size | gRPC Available |
|-------|---------|-------------|----------------|
| Default | `go build -o smoke .` | ~8MB | No (stub returns error) |
| With gRPC | `go build -tags grpc -o smoke .` | ~13.9MB | Yes |

- `go.mod` unchanged — build tags only control compilation
- Schema struct `GRPCHealthCheck` remains in schema.go (always parsed, validation always works)
- Runner wiring in `runner.go` unchanged — calls `CheckGRPCHealth` regardless
- No behavior change for users who install via `go install` (they get default build)

### Testing

Existing gRPC tests move to `assertion_grpc_test.go` with `//go:build grpc` tag. New stub test:
```go
//go:build !grpc

func TestCheckGRPCHealth_StubReturns(t *testing.T) {
    result := CheckGRPCHealth(&schema.GRPCHealthCheck{Address: "localhost:9090"})
    if result.Passed {
        t.Error("stub should not pass")
    }
    if !strings.Contains(result.Actual, "grpc") {
        t.Error("should mention grpc in output")
    }
}
```

## File Scope

```yaml
files_modified:
  - cmd/run.go                      # --monorepo flag
  - internal/schema/schema.go       # new Settings fields, WebSocketCheck struct
  - internal/schema/validate.go     # WebSocket + monorepo validation
  - internal/runner/runner.go       # RunMonorepo method, WebSocket wiring
  - internal/runner/assertion.go    # Remove gRPC code, add CheckWebSocket
  - CLAUDE.md                       # Update assertion table
files_created:
  - internal/monorepo/discover.go   # SubConfig discovery
  - internal/monorepo/discover_test.go
  - internal/runner/assertion_grpc.go       # +build grpc
  - internal/runner/assertion_grpc_stub.go  # +build !grpc
  - internal/runner/assertion_ws_test.go    # WebSocket tests
  - internal/runner/assertion_grpc_test.go  # gRPC tests (moved, +build grpc)
```

## v0.7 Release Scope

Headline: "cosmo-smoke v0.7 — Monorepo, WebSocket, Lean Binary"

- Monorepo sub-config auto-discovery (ROAD-010)
- WebSocket connect-send-expect assertion (ROAD-031)
- Optional gRPC module via build tag (ROAD-030)
- ~258+ tests passing (246 current + ~14 new)
