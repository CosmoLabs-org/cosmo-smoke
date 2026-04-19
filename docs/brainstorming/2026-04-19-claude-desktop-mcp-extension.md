# Claude Desktop MCP Extension for cosmo-smoke

**Date**: 2026-04-19
**Roadmap**: ROAD-032
**Status**: Brainstorm

## Goal

Make cosmo-smoke usable from Claude Desktop (and any MCP client) via the Model Context Protocol. Users ask Claude to generate, run, and debug smoke tests conversationally -- no CLI knowledge required.

---

## 1. MCP Tool Definitions

Seven tools covering the full smoke test lifecycle. Each maps to existing internal packages with minimal new logic.

### 1.1 `smoke_run` -- Execute smoke tests

**Purpose**: Run tests and return structured results for conversational debugging.

```json
{
  "name": "smoke_run",
  "description": "Run smoke tests from a .smoke.yaml config file. Returns pass/fail results with assertion details for each test. Use this to verify services are healthy, check endpoints, validate configs, or debug failures.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "config_path": {
        "type": "string",
        "description": "Path to .smoke.yaml (default: .smoke.yaml in working directory)"
      },
      "tags": {
        "type": "array",
        "items": { "type": "string" },
        "description": "Include only tests with these tags"
      },
      "exclude_tags": {
        "type": "array",
        "items": { "type": "string" },
        "description": "Exclude tests with these tags"
      },
      "fail_fast": {
        "type": "boolean",
        "description": "Stop on first failure (default: false)"
      },
      "timeout": {
        "type": "string",
        "description": "Per-test timeout override, e.g. '30s'"
      },
      "dry_run": {
        "type": "boolean",
        "description": "List tests without running them (default: false)"
      },
      "monorepo": {
        "type": "boolean",
        "description": "Discover and run .smoke.yaml in subdirectories (default: false)"
      },
      "env": {
        "type": "string",
        "description": "Load environment-specific config (e.g. 'staging' loads staging.smoke.yaml)"
      }
    }
  },
  "annotations": {
    "title": "Run Smoke Tests",
    "readOnlyHint": false,
    "destructiveHint": false,
    "idempotentHint": false,
    "openWorldHint": true
  }
}
```

**Returns**: JSON-structured result reusing `reporter/json.go` format (project, total, passed, failed, per-test assertion details with expected vs actual). On failure, includes `fix_suggestions` (see Section 5).

**Reuses**: `runner.Runner.Run()`, `runner.Runner.RunMonorepo()`, `schema.Load()`, `schema.MergeEnv()`, `reporter.JSON` pattern.

### 1.2 `smoke_init` -- Generate config for a project

**Purpose**: Auto-detect project type and generate `.smoke.yaml`. Conversational onboarding.

```json
{
  "name": "smoke_init",
  "description": "Generate a .smoke.yaml smoke test config for a project. Auto-detects Go, Node, Python, Docker, and Rust projects. Can also inspect a running Docker container. Returns the generated config without writing to disk unless confirm=true.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "directory": {
        "type": "string",
        "description": "Project directory to scan (default: working directory)"
      },
      "from_container": {
        "type": "string",
        "description": "Generate config by inspecting a running Docker container name"
      },
      "write": {
        "type": "boolean",
        "description": "Write .smoke.yaml to disk (default: false, returns YAML as text)"
      },
      "force": {
        "type": "boolean",
        "description": "Overwrite existing .smoke.yaml (default: false)"
      }
    }
  },
  "annotations": {
    "title": "Generate Smoke Test Config",
    "readOnlyHint": false,
    "destructiveHint": false,
    "idempotentHint": false,
    "openWorldHint": false
  }
}
```

**Returns**: Generated YAML as text. If `write=true`, confirms file path written.

**Reuses**: `detector.Detect()`, `detector.GenerateConfig()`, `detector.InspectContainer()`.

### 1.3 `smoke_validate` -- Validate a config file

**Purpose**: Check `.smoke.yaml` for errors without running tests. Returns all validation errors at once.

```json
{
  "name": "smoke_validate",
  "description": "Validate a .smoke.yaml config file without running tests. Checks for required fields, assertion consistency, regex validity, and structural correctness. Returns all errors at once.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "config_path": {
        "type": "string",
        "description": "Path to .smoke.yaml (default: .smoke.yaml)"
      }
    }
  },
  "annotations": {
    "title": "Validate Smoke Config",
    "readOnlyHint": true,
    "destructiveHint": false,
    "idempotentHint": true,
    "openWorldHint": false
  }
}
```

**Returns**: Valid (list of tests found) or list of validation errors with field paths.

**Reuses**: `schema.Load()`, `schema.Validate()`.

### 1.4 `smoke_list` -- List tests in a config

**Purpose**: Enumerate available tests, their assertions, and tags. Lets Claude understand what's configured before running.

```json
{
  "name": "smoke_list",
  "description": "List all smoke tests defined in a .smoke.yaml config. Shows test names, tags, command, and assertion types. Useful for understanding what's configured before running.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "config_path": {
        "type": "string",
        "description": "Path to .smoke.yaml (default: .smoke.yaml)"
      },
      "tags": {
        "type": "array",
        "items": { "type": "string" },
        "description": "Filter to tests with these tags"
      },
      "monorepo": {
        "type": "boolean",
        "description": "Discover configs in subdirectories (default: false)"
      }
    }
  },
  "annotations": {
    "title": "List Smoke Tests",
    "readOnlyHint": true,
    "destructiveHint": false,
    "idempotentHint": true,
    "openWorldHint": false
  }
}
```

**Returns**: Array of `{name, tags, run_command, assertion_types[], skip_if}`.

**Reuses**: `schema.Load()`, `monorepo.Discover()`.

### 1.5 `smoke_explain` -- Explain an assertion type

**Purpose**: Claude can look up what an assertion does and how to configure it. Reduces hallucination about assertion fields.

```json
{
  "name": "smoke_explain",
  "description": "Explain a smoke test assertion type and its configuration. Returns the assertion's fields, defaults, and an example YAML snippet. Use when you need to understand or construct assertion configurations.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "assertion_type": {
        "type": "string",
        "description": "Assertion type to explain",
        "enum": [
          "exit_code", "stdout_contains", "stdout_matches",
          "stderr_contains", "stderr_matches", "file_exists",
          "env_exists", "port_listening", "process_running",
          "http", "json_field", "response_time_ms",
          "ssl_cert", "redis_ping", "memcached_version",
          "postgres_ping", "mysql_ping", "grpc_health",
          "websocket", "docker_container_running",
          "docker_image_exists", "url_reachable",
          "service_reachable", "s3_bucket", "version_check",
          "otel_trace", "credential_check", "graphql"
        ]
      }
    },
    "required": ["assertion_type"]
  },
  "annotations": {
    "title": "Explain Assertion Type",
    "readOnlyHint": true,
    "destructiveHint": false,
    "idempotentHint": true,
    "openWorldHint": false
  }
}
```

**Returns**: Field descriptions, defaults, YAML example, notes on standalone vs command-based usage.

**Implementation**: Static lookup table in `internal/mcp/assertions.go`. No runtime dependency on existing packages -- pure data.

### 1.6 `smoke_discover` -- Find configs in a directory tree

**Purpose**: Locate `.smoke.yaml` files across a workspace or monorepo.

```json
{
  "name": "smoke_discover",
  "description": "Find all .smoke.yaml config files in a directory tree. Returns paths and project names. Useful for understanding the test landscape of a workspace.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "directory": {
        "type": "string",
        "description": "Root directory to search (default: working directory)"
      },
      "depth": {
        "type": "number",
        "description": "Maximum search depth (default: unlimited)"
      }
    }
  },
  "annotations": {
    "title": "Discover Smoke Configs",
    "readOnlyHint": true,
    "destructiveHint": false,
    "idempotentHint": true,
    "openWorldHint": false
  }
}
```

**Returns**: Array of `{path, directory, project_name}`.

**Reuses**: `monorepo.Discover()`.

### 1.7 `smoke_generate_test` -- Generate a single test

**Purpose**: Generate a YAML snippet for a specific test case. Claude describes what to test, this returns valid YAML.

```json
{
  "name": "smoke_generate_test",
  "description": "Generate a single smoke test YAML snippet. Provide what you want to test and get back valid YAML to add to .smoke.yaml. Supports all 29 assertion types.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "description": "Test name"
      },
      "description": {
        "type": "string",
        "description": "What this test should verify (natural language)"
      },
      "assertion_type": {
        "type": "string",
        "description": "Primary assertion type (e.g. 'http', 'port_listening', 'redis_ping')"
      },
      "params": {
        "type": "object",
        "description": "Assertion parameters as key-value pairs",
        "additionalProperties": true
      },
      "tags": {
        "type": "array",
        "items": { "type": "string" },
        "description": "Tags for this test"
      }
    },
    "required": ["name", "assertion_type"]
  },
  "annotations": {
    "title": "Generate Smoke Test",
    "readOnlyHint": true,
    "destructiveHint": false,
    "idempotentHint": true,
    "openWorldHint": false
  }
}
```

**Returns**: Valid YAML snippet for the test.

**Implementation**: Template-based generation using `internal/detector/templates.go` patterns.

---

## 2. Architecture Decision: `smoke mcp` Subcommand vs Standalone Binary

### Option A: `smoke mcp` subcommand

Add a `cmd/mcp.go` that starts the MCP server via stdio.

**Pros**:
- Single binary distribution -- no new build target
- Shares version, root command banner, ldflags injection
- Users already have `smoke` installed; just run `smoke mcp`
- Claude Desktop config: `"command": "smoke", "args": ["mcp"]`

**Cons**:
- Binary size grows with `mcp-go` dependency (~200KB)
- Users who don't use MCP still pull in the dependency
- Longer process startup (Cobra init, though negligible)

### Option B: Standalone `smoke-mcp` binary

Separate `cmd/smoke-mcp/main.go` with its own build target.

**Pros**:
- Clean separation -- MCP users get a focused binary
- No dependency bloat for core `smoke` users
- Can version independently

**Cons**:
- Two binaries to build, distribute, and version
- Duplicate schema/runner/reporter imports
- More complex Makefile, release process, and Homebrew formula
- Users must install two tools

### Recommendation: Option A -- `smoke mcp` subcommand

Justification:
1. cosmo-smoke's philosophy is "minimal deps" but mcp-go (~200KB compiled) is modest.
2. Single-binary UX is critical for Claude Desktop adoption. Users install `smoke` once, add `"command": "smoke", "args": ["mcp"]` to their MCP config, done.
3. The existing `smoke serve` command already established the pattern of long-running subcommands.
4. Can use a build tag (`//go:build mcp`) to make it opt-in if dependency size becomes a concern later.
5. The Go plugin landscape for MCP is nascent -- having one binary avoids version skew between `smoke` and `smoke-mcp` as the protocol evolves.

**Claude Desktop configuration** (what the user adds):
```json
{
  "mcpServers": {
    "cosmo-smoke": {
      "command": "smoke",
      "args": ["mcp"]
    }
  }
}
```

---

## 3. Long-Running Tests

### Problem

Smoke tests can take 30s+ (HTTP timeouts, otel_trace waits, retry with backoff). MCP tool calls should return within a reasonable time. Claude Desktop shows no progress UI during tool execution.

### Design: Task-Augmented Tools via mcp-go

mcp-go v0.34+ supports **task-augmented tools** with three modes:
- `TaskSupportForbidden` -- synchronous only (default)
- `TaskSupportOptional` -- can run sync or async
- `TaskSupportRequired` -- always async, returns task ID for polling

**Recommendation**: Use `TaskSupportOptional` for `smoke_run`.

**Flow**:

1. **Fast tests (<30s)**: Claude calls `smoke_run` without task parameter. Synchronous response. Most smoke tests complete in 1-10s.

2. **Slow tests (>30s)**: Claude calls `smoke_run` with `{ "_task": {} }` parameter. Server returns a task ID immediately. Claude polls `tasks/result` to get the outcome.

3. **Progress notifications**: While the task runs, the server sends `notifications/tasks/{taskId}` with status updates (running, percentage complete based on tests completed / total tests).

**Implementation**:

```go
// cmd/mcp.go
smokeRunTool := mcp.NewTool("smoke_run",
    mcp.WithDescription("Run smoke tests..."),
    mcp.WithTaskSupport(mcp.TaskSupportOptional),
    // ... other params
)

s.AddTaskTool(smokeRunTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CreateTaskResult, error) {
    // Execute tests in background
    // Send progress notifications via ctx
    // Return task result
})
```

**Timeout handling**:
- Default per-test timeout: 30s (matches existing behavior in `runner.go:296`)
- Total suite timeout: sum of all test timeouts + 60s buffer
- Context cancellation from MCP client propagates to `exec.CommandContext` in runner

**What about other tools?**: `smoke_validate`, `smoke_list`, `smoke_explain`, `smoke_discover`, `smoke_generate_test` are all fast (<1s) and always synchronous. `smoke_init` is fast for auto-detect but `from_container` inspection could take 5-10s -- still fine for synchronous.

---

## 4. Config Discovery

### Working Directory

MCP servers launched by Claude Desktop inherit the working directory from the Claude Desktop process, which is typically `$HOME`. This is wrong for project-scoped smoke tests.

**Solution**: All tools that take a `config_path` or `directory` parameter resolve relative to the **project root**, not the process cwd. Detection strategy:

1. **Explicit path**: If user says "run smoke tests in /Users/gab/myproject", pass `config_path=/Users/gab/myproject/.smoke.yaml`.
2. **Claude Desktop workspace**: Claude Desktop passes project context to tools. The MCP handler checks if the `config_path` is relative and resolves it against the workspace root if available.
3. **Walk up**: If no path specified, walk up from cwd looking for `.smoke.yaml`, `.git/`, or `go.mod` to find the project root. This matches how `smoke run` already works (defaults to `.smoke.yaml` in cwd).

**Implementation in handler**:

```go
func resolveConfigPath(configPath string) (string, error) {
    if configPath == "" {
        configPath = ".smoke.yaml"
    }
    if filepath.IsAbs(configPath) {
        return configPath, nil
    }
    // Walk up from cwd looking for project root markers
    cwd, _ := os.Getwd()
    dir := cwd
    for {
        if _, err := os.Stat(filepath.Join(dir, configPath)); err == nil {
            return filepath.Join(dir, configPath), nil
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            break
        }
        dir = parent
    }
    return "", fmt.Errorf("no %s found in %s or any parent", configPath, cwd)
}
```

### Monorepo Support

The `smoke_run` and `smoke_list` tools accept a `monorepo: true` parameter, which triggers `monorepo.Discover()` exactly as the CLI does today. The MCP handler reuses `runner.Runner.RunMonorepo()`.

For `smoke_discover`, the tool walks a directory tree finding all `.smoke.yaml` files. This is a lightweight wrapper around `monorepo.Discover()` with an optional depth parameter.

### Multi-Project Sessions

Claude Desktop may have multiple projects open. Each `smoke_run` call specifies which project via `config_path`. No session-level state is needed -- each tool call is self-contained. This is stateless by design and avoids the complexity of per-session working directory tracking.

---

## 5. Error Surfacing for Conversational Debugging

### Design Principle

When a test fails, Claude needs enough context to explain *why* and suggest *what to do*. Raw assertion output is not enough -- we need conversational remediation hints.

### Error Response Format

Every `smoke_run` result includes a structured JSON body:

```json
{
  "tool_result": {
    "status": "failed",
    "summary": {
      "total": 6,
      "passed": 4,
      "failed": 1,
      "skipped": 1,
      "duration_ms": 3200
    },
    "tests": [
      {
        "name": "API health check",
        "status": "passed",
        "duration_ms": 150
      },
      {
        "name": "Redis connectivity",
        "status": "failed",
        "duration_ms": 50,
        "assertions": [
          {
            "type": "redis_ping",
            "expected": "+PONG response from Redis",
            "actual": "connection refused on localhost:6379",
            "passed": false
          }
        ],
        "fix_suggestions": [
          "Redis is not running or not listening on localhost:6379",
          "Start Redis: docker run -d -p 6379:6379 redis:alpine",
          "If Redis is on a different host/port, update redis_ping.host and redis_ping.port in .smoke.yaml"
        ]
      }
    ]
  }
}
```

### Fix Suggestion Engine

Each assertion type has a suggestion mapper. This is a static lookup keyed by assertion type + failure pattern.

```go
// internal/mcp/suggestions.go

type Suggestion struct {
    Condition string // "connection_refused", "timeout", "wrong_status", "not_found", etc.
    Message   string
    Action    string // concrete fix command or config change
}

var suggestionMap = map[string][]Suggestion{
    "redis_ping": {
        {Condition: "connection_refused", Message: "Redis is not running", Action: "docker run -d -p 6379:6379 redis:alpine"},
        {Condition: "auth_failed", Message: "Redis requires authentication", Action: "Add redis_ping.password to .smoke.yaml"},
    },
    "http": {
        {Condition: "connection_refused", Message: "Server not listening", Action: "Check if the service is running and the port is correct"},
        {Condition: "wrong_status", Message: "Unexpected HTTP status code", Action: "Verify the endpoint returns the expected status code"},
        {Condition: "timeout", Message: "Request timed out", Action: "Increase timeout or check if the service is overloaded"},
    },
    "port_listening": {
        {Condition: "not_listening", Message: "Port is not open", Action: "Start the service that should listen on this port"},
    },
    "postgres_ping": {
        {Condition: "connection_refused", Message: "Postgres not running", Action: "docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:alpine"},
    },
    // ... for all 29 assertion types
}
```

The suggestion engine extracts the failure condition from `AssertionResult.Actual` and matches it against known patterns. Fallback: generic "check the configuration for this assertion type" message.

### Truncation for Large Output

If stdout/stderr in assertion actual values exceeds 2KB, truncate with `[... truncated, full output: X bytes]`. This prevents context window flooding. Claude can ask the user to run a targeted test if they need the full output.

---

## 6. Dependencies

### Primary: `github.com/mark3labs/mcp-go`

- **Current version**: v0.34.0 (latest as of 2026-04)
- **MCP spec version**: 2025-11-25 (backward compat to 2024-11-05)
- **Transport**: stdio (for Claude Desktop)
- **Features needed**: Tools, Task-augmented tools, Progress notifications
- **License**: MIT
- **Stars**: ~4k, actively maintained

**Usage pattern**:
```go
import (
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)
```

**Key API surface we need**:
- `server.NewMCPServer()` -- server creation
- `mcp.NewTool()` -- tool definition with typed params
- `s.AddTool()` -- sync tool handler
- `s.AddTaskTool()` -- async tool handler (for `smoke_run`)
- `server.ServeStdio()` -- stdio transport for Claude Desktop
- `server.WithToolCapabilities()` -- enable tool listing
- `server.WithRecovery()` -- panic recovery in handlers
- `server.WithHooks()` -- request lifecycle hooks for logging

### No new runtime dependencies

mcp-go is the only new dependency. It has its own transitive deps but they don't conflict with cosmo-smoke's existing deps (Cobra, Lipgloss, yaml.v3, gjson, fsnotify, grpc).

### Dependency audit

| Package | Added | Purpose |
|---------|-------|---------|
| `github.com/mark3labs/mcp-go` | Yes | MCP protocol implementation |
| `github.com/mark3labs/mcp-go/mcp` | Yes (transitive) | Tool/resource types |
| `github.com/mark3labs/mcp-go/server` | Yes (transitive) | Server + transport |

No new external service dependencies. MCP server runs locally via stdio.

---

## 7. Implementation Plan

### File Structure

```
cmd/
  mcp.go                    # NEW: `smoke mcp` subcommand, server bootstrap
  mcp_test.go               # NEW: integration test for MCP server startup
internal/
  mcp/
    server.go               # NEW: MCP server creation, tool registration
    handlers.go             # NEW: Tool handler implementations (7 tools)
    handlers_test.go        # NEW: Handler unit tests
    suggestions.go          # NEW: Fix suggestion engine
    suggestions_test.go     # NEW: Suggestion engine tests
    assertions.go           # NEW: Assertion type documentation lookup
    assertions_test.go      # NEW: Assertion lookup tests
    types.go                # NEW: Shared types (MCP result wrappers)
    discovery.go            # NEW: Config path resolution (walk-up logic)
    discovery_test.go       # NEW: Discovery tests
```

### Implementation Order

**Phase 1: Core MCP server (MVP)**

1. `cmd/mcp.go` -- Cobra subcommand that creates and starts stdio MCP server
2. `internal/mcp/server.go` -- Server factory with tool registration
3. `internal/mcp/types.go` -- Result wrapper types
4. `internal/mcp/handlers.go` -- `smoke_run` handler (wraps existing `runner.Runner`)
5. Test: Start server, send `tools/list`, verify `smoke_run` appears
6. Test: Call `smoke_run` against cosmo-smoke's own `.smoke.yaml`, verify results

**Phase 2: Read-only tools**

7. `smoke_validate` handler -- wraps `schema.Load()` + `schema.Validate()`
8. `smoke_list` handler -- wraps `schema.Load()` + test enumeration
9. `smoke_discover` handler -- wraps `monorepo.Discover()`
10. `smoke_explain` handler -- static assertion docs lookup
11. Tests for all four

**Phase 3: Config generation**

12. `smoke_init` handler -- wraps `detector.Detect()` + `detector.GenerateConfig()`
13. `smoke_generate_test` handler -- template-based YAML generation
14. Tests for both

**Phase 4: Error UX**

15. `internal/mcp/suggestions.go` -- Fix suggestion engine for all 29 assertion types
16. Integrate suggestions into `smoke_run` response
17. Output truncation for large stdout/stderr
18. Tests

**Phase 5: Long-running support**

19. Convert `smoke_run` to task-augmented tool with `TaskSupportOptional`
20. Progress notifications (test N of M completed)
21. Test with slow-running tests (otel_trace, retry)

**Phase 6: Polish**

22. `internal/mcp/discovery.go` -- Config path resolution with walk-up
23. Claude Desktop `claude_desktop_config.json` example in README
24. Documentation: MCP tools reference
25. `go.mod` update with `mcp-go` dependency

### Estimated Effort

| Phase | Files | LOC (new) | Time |
|-------|-------|-----------|------|
| 1: Core server | 5 | ~300 | 2h |
| 2: Read-only tools | 1 | ~250 | 1.5h |
| 3: Config generation | 1 | ~150 | 1h |
| 4: Error UX | 2 | ~400 | 2h |
| 5: Long-running | 1 | ~100 | 1h |
| 6: Polish | 3 | ~150 | 1h |
| **Total** | **13** | **~1350** | **~8.5h** |

### Testing Strategy

- **Unit tests**: Each handler tested in isolation with mock `schema.Load` / `runner.Runner`
- **Integration test**: Start stdio MCP server, send JSON-RPC messages, verify responses
- **Self-smoke**: Run `smoke mcp` and call `smoke_run` against cosmo-smoke's own `.smoke.yaml`
- **Claude Desktop test**: Configure MCP server in Claude Desktop, run conversational smoke tests

### MCP Prompts (Future Enhancement)

Not in v1, but the MCP `prompts` capability could expose:
- `"troubleshoot_failures"` -- Claude reads recent failure output and suggests fixes
- `"add_test_for"` -- conversational test generation with guided follow-up questions
- `"migrate_from_goss"` -- walks through goss-to-smoke migration

These would use `s.AddPrompt()` from mcp-go and reuse the `internal/migrate/goss/` package.

---

## Key Decisions Summary

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Binary strategy | `smoke mcp` subcommand | Single binary, simpler distribution |
| Transport | stdio | Claude Desktop's native MCP transport |
| Library | `mark3labs/mcp-go` v0.34+ | Most complete Go MCP implementation |
| Long-running tests | Task-augmented (`TaskSupportOptional`) | Fast tests sync, slow tests async |
| Error format | JSON with `fix_suggestions` array | Actionable conversational debugging |
| Working directory | Explicit `config_path` + walk-up | No reliance on inherited cwd |
| State management | Stateless tool calls | No session state, simpler, more robust |
| New dependency | Only `mcp-go` | Minimal dep philosophy preserved |
