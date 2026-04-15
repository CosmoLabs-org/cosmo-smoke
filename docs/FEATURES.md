# cosmo-smoke Features

Universal smoke test runner for any project, any language.

**Version**: 1.0.0 | **Status**: Stable | **License**: MIT

---

## Feature Status Legend

| Icon | Meaning |
|------|---------|
| ✅ | Implemented and stable |
| 📋 | Planned |
| ⭐ | Key differentiator |

---

## Core Runner

| Feature | Status | Description |
|---------|--------|-------------|
| **YAML config** | ✅ | Single `.smoke.yaml` file defines all tests |
| **5 assertion types** | ✅ ⭐ | exit_code, stdout_contains, stdout_matches, stderr_contains, file_exists |
| **Multiple assertions per test** | ✅ | All assertions in an `expect` block must pass |
| **Prerequisites** | ✅ | Pre-flight checks that abort the run if they fail |
| **Per-test cleanup** | ✅ | `cleanup` command runs after each test regardless of pass/fail |
| **Per-test timeout** | ✅ | `timeout` field overrides the global default |
| **Global settings** | ✅ | `timeout`, `fail_fast`, `parallel` in `settings` block |
| **Tag filtering** | ✅ | `--tag` and `--exclude-tag` flags for selective runs |
| **Fail-fast mode** | ✅ | `--fail-fast` flag or `settings.fail_fast` stops on first failure |
| **Dry run** | ✅ | `--dry-run` lists matching tests without executing |
| **Parallel execution** | ✅ | `settings.parallel: true` runs tests concurrently |

---

## Configuration

| Feature | Status | Description |
|---------|--------|-------------|
| **Config-dir-relative paths** | ✅ ⭐ | Commands run from the config file's directory, not the caller's cwd |
| **Custom config path** | ✅ | `-f` flag to load config from any path |
| **Full validation on load** | ✅ | All errors reported at once before any test runs |
| **Go duration strings** | ✅ | Timeouts accept `30s`, `2m`, `1m30s`, etc. |
| **Shell command execution** | ✅ | All commands run via `sh -c` — pipes, redirects, and operators work |

---

## Output & Reporting

| Feature | Status | Description |
|---------|--------|-------------|
| **Terminal reporter** | ✅ | Styled output with pass/fail indicators (Lipgloss) |
| **JSON reporter** | ✅ | Machine-readable output for CI pipelines (`--format json`) |
| **Pluggable reporter interface** | ✅ | Clean interface for adding custom reporters |
| **Exit codes** | ✅ | `0` = all pass, `1` = failures, `2` = config/arg error |

---

## Project Detection (smoke init)

| Feature | Status | Description |
|---------|--------|-------------|
| **Go detection** | ✅ | Detects `go.mod`, generates build + test checks |
| **Node detection** | ✅ | Detects `package.json`, generates install + lint checks |
| **Python detection** | ✅ | Detects `pyproject.toml` / `requirements.txt`, generates import check |
| **Rust detection** | ✅ | Detects `Cargo.toml`, generates build + test checks |
| **Docker detection** | ✅ | Detects `Dockerfile`, generates docker build check |
| **Force overwrite** | ✅ | `--force` flag regenerates config even if one already exists |

---

## Planned

| Feature | Description |
|---------|-------------|
| **Watch mode** | Re-run tests on file change |
| **HTML reporter** | Browser-viewable test results |
| **Test retries** | `retry: N` field to retry flaky tests before marking failed |
| **Skip field** | `skip: true` or `skip_if: <condition>` to conditionally exclude tests |
| **Remote config** | Load `.smoke.yaml` from a URL |
| **Matrix runs** | Run the same test suite against multiple targets/environments |

---

## Design Constraints

These are intentional limitations, not gaps:

- **No test discovery** — tests must be explicitly listed in config; no globbing
- **No built-in assertions for HTTP** — use `curl` with `stdout_contains` instead
- **No secrets management** — pass secrets via environment variables in `run` commands
- **Minimal dependencies** — Cobra + Lipgloss + yaml.v3 only; no Viper, no Bubbletea

---

## Architecture

```
cosmo-smoke/
├── cmd/                # CLI commands (run, init, version)
├── internal/
│   ├── schema/         # SmokeConfig structs, YAML parsing, validation
│   ├── runner/         # Assertion engine, prereq runner, test execution
│   ├── reporter/       # Terminal (Lipgloss) + JSON reporters
│   └── detector/       # Project type detection + template generation
├── .smoke.yaml         # Self-smoke tests for this repo
└── SPEC.md             # Full schema reference
```

---

## Quick Start

```bash
go install github.com/CosmoLabs-org/cosmo-smoke@latest
cd my-project
smoke init
smoke run
```
