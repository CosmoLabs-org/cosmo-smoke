# cosmo-smoke — Project Instructions

## Overview

Universal smoke test runner. Standalone Go binary that reads `.smoke.yaml` and runs lightweight smoke tests. Designed for CosmoLabs' ~95-project portfolio.

**Repository**: `github.com/CosmoLabs-org/cosmo-smoke`
**Company**: CosmoLabs
**Version**: 0.6.0

## Architecture

```
cmd/
├── root.go          # Cobra root command with banner
├── run.go           # smoke run — main entry point
├── init_cmd.go      # smoke init — auto-detect + generate config
└── version.go       # smoke version (ldflags-injected)
internal/
├── schema/          # SmokeConfig structs, YAML parsing, validation
├── runner/          # Assertion engine (15 types), prereq runner, test execution
├── reporter/        # Terminal (Lipgloss) + JSON reporters
└── detector/        # Project type detection + template generation
```

## Key Design Decisions

- **Minimal deps**: Cobra + Lipgloss + yaml.v3 + gjson. No Viper, no Bubbletea.
- **Pure assertions**: All 15 assertion types are pure functions — no side effects.
- **Config inheritance**: `includes:` directive + Go templates (`{{ .Env.FOO }}`).
- **Config-dir-relative**: Commands execute from the config file's directory, not cwd.
- **All errors at once**: Validation returns all errors, not just the first.
- **Reporter interface**: Terminal and JSON reporters are pluggable via interface.
- **Watch mode**: `--watch` keeps smoke resident and re-runs on file changes. fsnotify-backed. 500ms debounce.
- **Retry**: Opt-in `retry: {count, backoff}` on test level. Exponential backoff. No side effects on pass-first-try.

## Build & Test

```bash
go build ./...                    # Build
go test ./...                     # Run all tests (246 total)
smoke run                         # Self-smoke (6 tests)
go build -ldflags "-s -w -X github.com/CosmoLabs-org/cosmo-smoke/cmd.Version=X.Y.Z" -o smoke .
```

## Commands

```bash
smoke run [--tag X] [--exclude-tag X] [--format terminal|json|junit|tap|prometheus] [--fail-fast] [--timeout 30s] [-f path] [--dry-run] [--watch]
smoke init [--force] [--from-running CONTAINER]
smoke version
```

## Assertion Types

| Type | Field | Description |
|------|-------|-------------|
| exit_code | `*int` | Process exit code match |
| stdout_contains | `string` | Substring match on stdout |
| stdout_matches | `string` | Regex match on stdout |
| stderr_contains | `string` | Substring match on stderr |
| stderr_matches | `string` | Regex match on stderr |
| file_exists | `string` | File exists relative to config dir |
| env_exists | `string` | Environment variable exists |
| port_listening | `{port, protocol?, host?}` | TCP/UDP port is open |
| process_running | `string` | Named process currently running (pgrep -x / tasklist) |
| http | `{url, method?, status_code?, body_contains?, body_matches?, header_contains?}` | HTTP endpoint check |
| json_field | `{path, equals?, contains?, matches?}` | JSONPath assertion on stdout |
| response_time_ms | `*int` | Test duration must not exceed this threshold |
| ssl_cert | `{host, port?, min_days_remaining?, allow_self_signed?}` | TLS cert valid + expiry threshold |
| redis_ping | `{host?, port?, password?}` | Redis PING returns +PONG (RESP protocol) |
| memcached_version | `{host?, port?}` | Memcached `version` command returns VERSION |
| postgres_ping | `{host?, port?}` | Postgres server SSLRequest handshake returns valid protocol byte |
| mysql_ping | `{host?, port?}` | MySQL server sends valid v10 handshake packet on connection |
| grpc_health | `{address, service?, use_tls?, timeout?}` | grpc.health.v1 Health/Check returns SERVING |
| docker_container_running | `{name}` | Named Docker container is currently running |
| docker_image_exists | `{image}` | Docker image exists locally |
| url_reachable | `{url, timeout?, status_code?}` | HTTP/HTTPS connectivity check |
| service_reachable | `{url, timeout?}` | External service dependency check |
| s3_bucket | `{bucket, region?, endpoint?}` | S3-compatible bucket accessibility (anonymous HEAD) |
| version_check | `{command, pattern}` | Tool version verification via shell command + regex |

Plus `allow_failure: true` on Test for flaky/allowed-failure tests.

## Output Formats

`smoke run --format X` supports: `terminal` (default), `json`, `junit`, `tap`, `prometheus`.

## Detected Project Types

Go, Node (bun/npm), Python, Docker, Rust — each with tailored smoke test templates.
