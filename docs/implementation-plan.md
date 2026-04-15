---
completed: "2026-04-15"
created: "2026-04-15"
origin: /brainplan
status: COMPLETED
tags:
    - smoke-testing
    - cli
    - implementation
title: cosmo-smoke — Implementation Plan
---

# cosmo-smoke — Implementation Plan

## Goal

Build `smoke` v1.0.0 — a standalone Go binary that reads `.smoke.yaml` and runs smoke tests. Ship with 5 assertion types, colored terminal + JSON output, `smoke init` auto-scaffolding, and GoRalph's `.smoke.yaml` as the first consumer.

## File Scope

```yaml
repo: ~/PROJECTS/cosmo-smoke (NEW — must create)

files_create:
  # Project setup
  - main.go
  - go.mod
  - go.sum
  - .gitignore
  - README.md
  - LICENSE
  - .smoke.yaml                    # Eats its own dog food

  # CLI commands
  - cmd/root.go                    # Root command, global flags
  - cmd/run.go                     # smoke run
  - cmd/init_cmd.go                # smoke init (init.go conflicts with Go)
  - cmd/version.go                 # smoke version

  # Core packages
  - internal/schema/schema.go      # SmokeConfig structs, YAML parsing
  - internal/schema/validate.go    # Schema validation
  - internal/runner/runner.go      # Test execution engine
  - internal/runner/assertion.go   # Assertion implementations
  - internal/runner/prereq.go      # Prerequisite checks
  - internal/reporter/reporter.go  # Reporter interface
  - internal/reporter/terminal.go  # Colored terminal output
  - internal/reporter/json.go      # JSON output

  # Auto-scaffolding
  - internal/detector/detector.go  # Project type detection
  - internal/detector/templates.go # .smoke.yaml templates per stack

  # Tests (TDD — write before implementation)
  - internal/schema/schema_test.go
  - internal/schema/validate_test.go
  - internal/runner/runner_test.go
  - internal/runner/assertion_test.go
  - internal/runner/prereq_test.go
  - internal/reporter/terminal_test.go
  - internal/reporter/json_test.go
  - internal/detector/detector_test.go
  - cmd/run_test.go

files_modify_in_goralph:
  - .smoke.yaml                    # NEW — GoRalph's smoke config (first consumer)
```

## Implementation Steps

### Step 1: Project Bootstrap
Create `~/PROJECTS/cosmo-smoke`, `go mod init github.com/CosmoLabs-org/cosmo-smoke`, add Cobra + Lipgloss deps, `.gitignore`, LICENSE (MIT), basic `main.go` + `cmd/root.go`.

**Verify**: `go build ./...` succeeds, `./smoke --help` shows banner.

### Step 2: Schema Package
Implement `internal/schema/schema.go`:
- `SmokeConfig` struct with `Version`, `Project`, `Description`, `Settings`, `Prerequisites`, `Tests`
- `Settings` struct: `Timeout` (duration), `FailFast` (bool), `Parallel` (bool)
- `Prerequisite` struct: `Name`, `Check`, `Hint`
- `Test` struct: `Name`, `Run`, `Expect`, `Tags`, `Timeout`, `Cleanup`
- `Expect` struct: `ExitCode` (*int), `StdoutContains`, `StdoutMatches`, `StderrContains`, `FileExists`
- `Load(path string) (*SmokeConfig, error)` — read and parse YAML
- `LoadDefault() (*SmokeConfig, error)` — find `.smoke.yaml` in cwd

Implement `internal/schema/validate.go`:
- `Validate(cfg *SmokeConfig) error` — check required fields, valid version, at least one test
- Return all validation errors at once (not just first)

**TDD**: Write tests first covering valid configs, missing required fields, invalid version, empty tests.

**Verify**: `go test ./internal/schema/...`

### Step 3: Assertion Engine
Implement `internal/runner/assertion.go`:
- `type AssertionResult struct { Type, Expected, Actual string; Passed bool }`
- `CheckExitCode(actual int, expected int) AssertionResult`
- `CheckStdoutContains(stdout, substr string) AssertionResult`
- `CheckStdoutMatches(stdout, pattern string) AssertionResult`
- `CheckStderrContains(stderr, substr string) AssertionResult`
- `CheckFileExists(path, configDir string) AssertionResult` — resolve relative to configDir

All assertions are pure functions ��� no side effects, easy to test.

**TDD**: Write tests first for each assertion type including edge cases (empty stdout, invalid regex, relative vs absolute paths).

**Verify**: `go test ./internal/runner/...`

### Step 4: Prerequisite Runner
Implement `internal/runner/prereq.go`:
- `CheckPrerequisites(prereqs []schema.Prerequisite, timeout time.Duration) ([]PrereqResult, error)`
- Runs each prerequisite command via `sh -c`
- Captures first line of stdout as output (for display: "go1.26.2")
- Respects timeout — kills hung commands
- Returns results, doesn't abort (caller decides)

**TDD**: Test with commands that pass, fail, hang (timeout), and produce output.

**Verify**: `go test ./internal/runner/...`

### Step 5: Test Runner
Implement `internal/runner/runner.go`:
- `type Runner struct { Config *schema.SmokeConfig; Reporter reporter.Reporter; ConfigDir string }`
- `Run(opts RunOptions) (*SuiteResult, error)` — main entry point
- `RunOptions`: `Tags`, `ExcludeTags`, `FailFast` (override), `DryRun`, `Timeout` (override)
- Executes each test: `sh -c` the command, capture stdout/stderr/exit, evaluate all assertions, run cleanup
- Cleanup runs via defer-like pattern (even on failure/panic)
- Cleanup failures produce warnings, not errors
- Sequential by default; parallel when `settings.parallel: true` (use `errgroup`)
- Fail-fast: stop scheduling new tests, let running tests finish
- Tag filtering: include/exclude based on test tags

**TDD**: Test sequential execution, fail-fast behavior, tag filtering, cleanup execution, timeout enforcement.

**Verify**: `go test ./internal/runner/...`

### Step 6: Reporters
Implement `internal/reporter/reporter.go`:
- `type Reporter interface { PrereqStart(name string); PrereqResult(r PrereqResult); TestStart(name string); TestResult(r TestResult); Summary(s SuiteResult) }`

Implement `internal/reporter/terminal.go`:
- Colored output using Lipgloss
- Green ✓ / Red ✗ / Yellow ⊘ (skip)
- Right-aligned duration
- Failure details indented below test name
- Summary line: "N tests: X passed, Y failed, Z skipped (Ns)"

Implement `internal/reporter/json.go`:
- Collects all results, emits JSON on `Summary()` call
- Includes prerequisites with `hint` field
- Includes per-assertion detail

**TDD**: Test terminal reporter produces expected output patterns; JSON reporter produces valid JSON with correct structure.

**Verify**: `go test ./internal/reporter/...`

### Step 7: CLI Commands
Implement `cmd/run.go`:
- `smoke run` with flags: `--tag`, `--exclude-tag`, `--format` (terminal|json), `--fail-fast`, `--timeout`, `-f` (config path), `--dry-run`
- Loads config, creates runner + reporter, calls `runner.Run()`, exits with appropriate code

Implement `cmd/init_cmd.go`:
- `smoke init` — calls detector, writes `.smoke.yaml` to cwd
- If `.smoke.yaml` exists, prompt to overwrite

Implement `cmd/version.go`:
- `smoke version` — print version string

**Verify**: `go build . && ./smoke --help && ./smoke version`

### Step 8: Project Detector
Implement `internal/detector/detector.go`:
- `Detect(dir string) []ProjectType` — scan for markers (go.mod, package.json, etc.)
- `type ProjectType string` — Go, Node, Python, Docker, Rust

Implement `internal/detector/templates.go`:
- `GenerateConfig(types []ProjectType) *schema.SmokeConfig`
- Template tests per detected type, merged into one config
- Smart: check for `bun.lock` vs `package-lock.json`, check if `lint` script exists

**TDD**: Test detection with mock directory structures.

**Verify**: `go test ./internal/detector/...`

### Step 9: Self-Smoke & GoRalph Config
Create `cosmo-smoke/.smoke.yaml`:
```yaml
version: 1
project: cosmo-smoke
description: "Smoke tests for the smoke test runner"
settings:
  timeout: 5s
  fail_fast: true
tests:
  - name: "Compiles"
    run: "go build ./..."
    expect: { exit_code: 0 }
    tags: [build]
  - name: "Tests pass"
    run: "go test -short ./..."
    expect: { exit_code: 0 }
    tags: [test]
  - name: "Help flag"
    run: "go run . --help"
    expect:
      exit_code: 0
      stdout_contains: "smoke"
    tags: [runtime]
```

Create GoRalph `.smoke.yaml` (in the GoRalph repo) per the schema design doc.

**Verify**: `./smoke run` passes on itself. `cd ~/PROJECTS/GoRalph && smoke run` passes.

### Step 10: Polish & Release
- README.md: installation, quickstart, full schema reference
- `go install github.com/CosmoLabs-org/cosmo-smoke@latest` works
- Build + deploy to `~/bin/smoke` with codesign
- Tag v1.0.0

**Verify**: `smoke run` from GoRalph root succeeds. `smoke init` in a fresh Go project produces valid config.

## Execution Strategy

Steps 1-6 are independent enough for parallel GLM agents:
- **Agent A**: Steps 1-2 (bootstrap + schema) — must complete first
- **Agent B**: Step 3 (assertions) — after Step 2
- **Agent C**: Step 4 (prerequisites) — after Step 2
- **Agent D**: Steps 5-6 (runner + reporters) — after Steps 3-4
- **Agent E**: Steps 7-8 (CLI + detector) — after Step 5

Steps 9-10 are integration work — run in main session after agents merge.

Recommended: `/glm-sprint` with Opus review gates after each merge.

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/charmbracelet/lipgloss` — terminal styling
- `gopkg.in/yaml.v3` — YAML parsing
- No Viper, no Bubbletea — keep it minimal
