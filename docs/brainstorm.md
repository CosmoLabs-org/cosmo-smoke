---
completed: "2026-04-15"
created: "2026-04-15"
origin: /brainplan
status: COMPLETED
tags:
    - smoke-testing
    - cli
    - cross-project
    - design
title: cosmo-smoke — Universal Smoke Test System
---

# cosmo-smoke — Universal Smoke Test System

## Overview

A standalone Go binary (`smoke`) that reads a `.smoke.yaml` config file from any project root and executes lightweight "does it turn on?" verification tests. Designed for CosmoLabs' ~95-project portfolio but universally applicable.

**Repo**: `CosmoLabs-org/cosmo-smoke`
**Binary**: `smoke` (deployed to `~/bin/smoke`)
**Config**: `.smoke.yaml` at project root

## Design Decisions

| Decision | Choice | Why |
|----------|--------|-----|
| Home repo | Own repo (`cosmo-smoke`) | Clean separation, independent releases, can be open-sourced independently |
| Binary name | `smoke` | Short, universal, memorable |
| Config format | YAML (`.smoke.yaml`) | Familiar, widely supported, good for declarative configs |
| Config location | Project root | Convention over configuration, easy to find |
| v1 assertions | 5 types | Exit code, stdout contains/matches, stderr contains, file exists — covers 90% of cases |
| Performance budget | 10 seconds max | Fast enough for every-commit runs, forces smoke tests to stay lightweight |
| Output formats | Terminal + JSON | Colored terminal for humans, JSON for machine consumption |
| Auto-scaffold | Basic detection | Detect go.mod/package.json/etc, scaffold matching templates |
| Parallel execution | Global setting only | Per-test parallelism deferred to v2 |

## Architecture

```
cosmo-smoke/
├── main.go
├── cmd/
│   ├── root.go          # smoke --help, global flags (--version, --verbose)
│   ├── run.go           # smoke run [--tag X] [--exclude-tag Y] [--format json] [--fail-fast]
│   └── init.go          # smoke init (auto-detect + scaffold)
├── internal/
│   ├── schema/
│   │   ├── schema.go    # SmokeConfig, Test, Prerequisite structs, YAML parsing
│   │   └── validate.go  # Schema validation (required fields, valid assertion types)
│   ├── runner/
│   │   ├── runner.go     # Test execution engine (sequential or parallel)
│   │   ├── assertion.go  # Assertion implementations (exit_code, stdout_contains, etc.)
│   │   └── prereq.go     # Prerequisite/dependency checks
│   ├── reporter/
│   │   ├── reporter.go   # Reporter interface
│   │   ├── terminal.go   # Colored terminal output (green check, red X, summary)
│   │   └── json.go       # JSON output for machine consumption
│   └── detector/
│       └── detector.go   # Project type detection for smoke init
└── .smoke.yaml           # Eats its own dog food
```

**Dependencies**: Cobra (CLI), Lipgloss (terminal styling). No Viper needed — YAML parsing is simple enough with `gopkg.in/yaml.v3`.

## Schema: `.smoke.yaml`

### Format

```yaml
version: 1
project: goralph
description: "Go Ralph! CLI + web dashboard smoke tests"

settings:
  timeout: 5s           # Per-test default timeout
  fail_fast: true       # Stop on first failure
  parallel: false       # Run tests sequentially (default)

prerequisites:
  - name: "Go installed"
    check: "go version"
    hint: "Install Go 1.25+: https://go.dev/dl/"

tests:
  - name: "Go compiles"
    run: "go build ./..."
    expect:
      exit_code: 0
    tags: [build, go]

  - name: "CLI shows help"
    run: "go run . --help"
    expect:
      exit_code: 0
      stdout_contains: "Ralph Wiggum Loop"
    tags: [runtime, cli]

  - name: "CLI version"
    run: "go run . version"
    expect:
      exit_code: 0
      stdout_matches: "v\\d+\\.\\d+\\.\\d+"
    tags: [runtime, cli]

  - name: "Binary builds"
    run: "go build -o /tmp/goralph-smoke ."
    expect:
      exit_code: 0
      file_exists: "/tmp/goralph-smoke"
    cleanup: "rm -f /tmp/goralph-smoke"
    tags: [build, go]

  - name: "Frontend builds"
    run: "cd web && bun install && bun run build"
    expect:
      exit_code: 0
    timeout: 10s
    tags: [build, frontend]
```

### Execution Semantics

- **Shell**: All `run` and `cleanup` commands execute via `sh -c "..."` to support shell operators (`&&`, pipes, `cd`)
- **Working directory**: CWD for test commands is always the directory containing `.smoke.yaml`
- **Assertions**: All assertions in an `expect` block are evaluated independently. A test fails if ANY assertion fails. All failures are reported (no short-circuit).
- **Cleanup**: Runs even on test failure. If cleanup itself fails, a warning is printed but the test result is unchanged.
- **Prerequisites**: Inherit `settings.timeout`. A hung prereq is killed after timeout.
- **Fail-fast + parallel**: When both enabled, running tests are allowed to finish their current command, then remaining tests are skipped.
- **Precedence**: CLI flags > `settings` in YAML > built-in defaults
- **Regex dialect**: `stdout_matches` uses Go's `regexp` package (RE2 syntax — no lookahead/backreferences)
- **Path resolution**: `file_exists` paths resolve relative to the config file's directory. Absolute paths work as-is.
- **Schema version**: `version: 1` is required. The runner rejects any other value. Reserved for future schema migration.

### v1 Assertion Types

| Assertion | Type | Description |
|-----------|------|-------------|
| `exit_code` | int | Assert command exit code equals value |
| `stdout_contains` | string | Assert stdout contains substring |
| `stdout_matches` | string (regex) | Assert stdout matches regex pattern |
| `stderr_contains` | string | Assert stderr contains substring |
| `file_exists` | string (path) | Assert file exists after command runs |

### Schema Fields

**Top-level**: `version` (int, required), `project` (string), `description` (string), `settings` (object), `prerequisites` (array), `tests` (array)

**Settings**: `timeout` (duration string, default "5s"), `fail_fast` (bool, default true), `parallel` (bool, default false)

**Prerequisite**: `name` (string, required), `check` (string, required — command to run), `hint` (string — help text on failure)

**Test**: `name` (string, required), `run` (string, required — shell command), `expect` (object, required — assertions), `tags` (string array), `timeout` (duration string — overrides settings), `cleanup` (string — runs even on failure)

## Execution Model

### Run Order

1. Parse `.smoke.yaml` and validate schema
2. Run all `prerequisites` — if any fail, abort with hints
3. Execute `tests` sequentially (or parallel if `settings.parallel: true`)
4. For each test: run command -> capture stdout/stderr/exit -> evaluate `expect` -> run `cleanup`
5. Print summary

### Terminal Output

```
smoke v1.0.0 — goralph

Prerequisites
  ✓ Go installed (go1.26.2)
  ✓ Bun installed (1.2.4)

Tests
  ✓ Go compiles                          0.8s
  ✓ CLI shows help                       1.2s
  ✓ CLI version                          1.1s
  ✓ Binary builds                        0.9s
  ✗ Frontend builds                      3.4s
    └─ exit code 1 (expected 0)
       stderr: error: Could not resolve "react"

5 tests: 4 passed, 1 failed (7.4s)
```

### JSON Output (`--format json`)

```json
{
  "project": "goralph",
  "version": 1,
  "duration_ms": 7400,
  "summary": { "total": 5, "passed": 4, "failed": 1, "skipped": 0 },
  "prerequisites": [
    { "name": "Go installed", "passed": true, "output": "go1.26.2", "hint": "Install Go 1.25+" }
  ],
  "tests": [
    {
      "name": "Go compiles",
      "status": "passed",
      "duration_ms": 800,
      "assertions": { "exit_code": { "expected": 0, "actual": 0, "passed": true } }
    }
  ]
}
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All tests passed |
| 1 | One or more tests failed |
| 2 | Config/validation error |

### Fail-Fast Behavior

When `fail_fast: true` (default), stops after first failure. Remaining tests are skipped. Cleanup still runs for the failed test. Summary shows skipped count.

## `smoke init` Detection

Template-based scaffolding — detect project markers, generate matching `.smoke.yaml`:

| Marker | Tests Generated |
|--------|----------------|
| `go.mod` | `go build ./...`, `go vet ./...`, `go test -short ./...` |
| `package.json` + `bun.lock` | `bun install`, `bun run build`, `bun run lint` (if script exists) |
| `package.json` + `package-lock.json` | Same but with `npm` |
| `pyproject.toml` | `pip install -e .`, `pytest --co -q` |
| `Dockerfile` | `docker build .` |
| `Cargo.toml` | `cargo check`, `cargo test --no-run` |

Multi-stack projects (like GoRalph with Go + React) get tests from ALL detected markers, tagged by stack.

## v1 CLI Commands

```bash
smoke run                     # Run all tests from .smoke.yaml
smoke run --tag build         # Only tests tagged "build"
smoke run --exclude-tag frontend  # Skip frontend tests
smoke run --format json       # JSON output
smoke run --fail-fast=false   # Override fail_fast setting
smoke run --timeout 30s       # Override global timeout
smoke run -f path/to/smoke.yaml  # Explicit config path
smoke init                    # Auto-detect and scaffold .smoke.yaml
smoke run --dry-run            # Validate config + check prereqs without running tests
smoke version                 # Print version
```

## Deferred to v2

- HTTP endpoint checks (start server, hit URL, check response)
- JSON field assertions (parse stdout as JSON)
- File content checks
- Environment variable assertions
- Per-test parallel execution
- `allow_failure: true` for flaky tests
- Retry logic with backoff
- Conditional tests (OS-specific, Docker-available)
- JUnit XML / TAP output formats
- CCS integration (`ccs smoke` wrapper)
- GitHub Actions reusable workflow
- Cross-project dashboard
- Secret/credential handling
- Monorepo sub-config support

## GoRalph as First Consumer

GoRalph gets a `.smoke.yaml` at its project root covering:
- Go build + vet (build smoke)
- CLI help + version flags (runtime smoke)
- Binary compilation (build artifact)
- Frontend install + build (frontend smoke)

This file lives in the GoRalph repo, not in cosmo-smoke.
