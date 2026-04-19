# cosmo-smoke Features

Universal smoke test runner for any project, any language.

**Version**: 0.6.0 | **Status**: Stable | **License**: MIT

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
| **24 assertion types** | ✅ ⭐ | Process, file, env, network, database, Docker, storage, tool verification |
| **Multiple assertions per test** | ✅ | All assertions in an `expect` block must pass |
| **Prerequisites** | ✅ | Pre-flight checks that abort the run if they fail |
| **Per-test cleanup** | ✅ | `cleanup` command runs after each test regardless of pass/fail |
| **Per-test timeout** | ✅ | `timeout` field overrides the global default |
| **Global settings** | ✅ | `timeout`, `fail_fast`, `parallel` in `settings` block |
| **Tag filtering** | ✅ | `--tag` and `--exclude-tag` flags for selective runs |
| **Fail-fast mode** | ✅ | `--fail-fast` flag or `settings.fail_fast` stops on first failure |
| **Dry run** | ✅ | `--dry-run` lists matching tests without executing |
| **Parallel execution** | ✅ | `settings.parallel: true` runs tests concurrently |
| **Watch mode** | ✅ | `--watch` re-runs tests on file changes (fsnotify, 500ms debounce) |
| **Retry flaky tests** | ✅ | `retry: {count, backoff}` with exponential backoff |
| **Allow failure** | ✅ | `allow_failure: true` marks test as passing even on assertion failure |
| **Conditional execution** | ✅ | `skip_if: {env_unset, env_equals, file_missing}` to skip tests conditionally |

---

## Assertion Types

### Process & Output

| Type | Status | Description |
|------|--------|-------------|
| `exit_code` | ✅ | Exact process exit code match |
| `stdout_contains` | ✅ | Substring present in stdout |
| `stdout_matches` | ✅ | Go regex match against stdout |
| `stderr_contains` | ✅ | Substring present in stderr |
| `stderr_matches` | ✅ | Go regex match against stderr |
| `response_time_ms` | ✅ | Test duration must not exceed threshold (ms) |
| `json_field` | ✅ | JSONPath assertion on stdout (equals/contains/matches) |

### File & Environment

| Type | Status | Description |
|------|--------|-------------|
| `file_exists` | ✅ | Path exists relative to config file directory |
| `env_exists` | ✅ | Environment variable is set (non-empty) |

### Network

| Type | Status | Description |
|------|--------|-------------|
| `port_listening` | ✅ | TCP/UDP port is open |
| `process_running` | ✅ | Named process is running (pgrep -x / tasklist) |
| `http` | ✅ ⭐ | Full HTTP endpoint validation (status, body, headers) |
| `url_reachable` | ✅ | Lightweight HTTP/HTTPS connectivity check |
| `service_reachable` | ✅ | External service dependency check (semantic naming) |
| `ssl_cert` | ✅ | TLS cert validity + expiry threshold |

### Database & Protocol

| Type | Status | Description |
|------|--------|-------------|
| `redis_ping` | ✅ | Redis PING returns +PONG (RESP protocol) |
| `memcached_version` | ✅ | Memcached `version` returns VERSION |
| `postgres_ping` | ✅ | Postgres SSLRequest handshake valid |
| `mysql_ping` | ✅ | MySQL v10 handshake packet valid |
| `grpc_health` | ✅ | grpc.health.v1 Health/Check returns SERVING |

### Storage & Docker

| Type | Status | Description |
|------|--------|-------------|
| `s3_bucket` | ✅ | S3-compatible bucket accessibility (anonymous HEAD) |
| `docker_container_running` | ✅ | Named Docker container is running |
| `docker_image_exists` | ✅ | Docker image exists locally |

### Tool Verification

| Type | Status | Description |
|------|--------|-------------|
| `version_check` | ✅ | Shell command output matches regex |

---

## Configuration

| Feature | Status | Description |
|---------|--------|-------------|
| **Config-dir-relative paths** | ✅ ⭐ | Commands run from the config file's directory, not the caller's cwd |
| **Custom config path** | ✅ | `-f` flag to load config from any path |
| **Full validation on load** | ✅ | All errors reported at once before any test runs |
| **Go duration strings** | ✅ | Timeouts accept `30s`, `2m`, `1m30s`, etc. |
| **Shell command execution** | ✅ | All commands run via `sh -c` — pipes, redirects, and operators work |
| **Config inheritance** | ✅ | `includes:` directive to share tests across configs |
| **Go templates** | ✅ | `{{ .Env.FOO }}` env var references in config values |
| **Multi-environment** | ✅ | Load base config + env-specific overrides |

---

## Output & Reporting

| Feature | Status | Description |
|---------|--------|-------------|
| **Terminal reporter** | ✅ | Styled output with pass/fail indicators (Lipgloss) |
| **JSON reporter** | ✅ | Machine-readable output for CI pipelines (`--format json`) |
| **JUnit reporter** | ✅ | JUnit XML for GitHub Actions, Jenkins, GitLab CI (`--format junit`) |
| **TAP reporter** | ✅ | Test Anything Protocol (`--format tap`) |
| **Prometheus reporter** | ✅ | Prometheus metrics (`--format prometheus`) |
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
| **From running container** | ✅ | `--from-running <container>` generates config from a running Docker container |

---

## Integration

| Feature | Status | Description |
|---------|--------|-------------|
| **Pre-commit hook** | ✅ | `.pre-commit-hooks.yaml` for zero-config integration |
| **CI/CD ready** | ✅ | JSON/JUnit output, exit codes, `--fail-fast` |

---

## Planned

| Feature | Description |
|---------|-------------|
| **Monorepo sub-config** | Auto-discover and run `.smoke.yaml` in subdirectories |
| **WebSocket assertion** | Connect/send/expect pattern for real-time apps |
| **Optional gRPC build tag** | Compile without gRPC deps for smaller binaries |
| **Claude Desktop MCP** | Generate `.smoke.yaml` from Dockerfile/compose via MCP |

---

## Design Constraints

These are intentional limitations, not gaps:

- **No test discovery** — tests must be explicitly listed in config; no globbing
- **No secrets management** — pass secrets via Go template env var references in commands
- **Minimal dependencies** — Cobra + Lipgloss + yaml.v3 + gjson + gRPC; no Viper, no Bubbletea
- **S3 is anonymous-only** — authenticated access uses the `http` assertion with Go template env vars
- **version_check is Unix-only** — uses `sh -c` which doesn't exist on Windows

---

## Architecture

```
cosmo-smoke/
├── cmd/                # CLI commands (run, init, version)
├── internal/
│   ├── schema/         # SmokeConfig structs, YAML parsing, validation
│   ├── runner/         # Assertion engine (24 types), prereq runner, test execution
│   ├── reporter/       # Terminal (Lipgloss) + JSON + JUnit + TAP + Prometheus reporters
│   └── detector/       # Project type detection + template generation
├── .smoke.yaml         # Self-smoke tests for this repo
├── .pre-commit-hooks.yaml  # Pre-commit framework integration
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
