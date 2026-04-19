# cosmo-smoke

Universal smoke test runner. Define lightweight "does it turn on?" checks in `.smoke.yaml` and run them with a single command — on any project, in any language.

## Install

**Go install:**
```bash
go install github.com/CosmoLabs-org/cosmo-smoke@latest
```

**Build from source:**
```bash
git clone https://github.com/CosmoLabs-org/cosmo-smoke
cd cosmo-smoke
go build -o smoke .
```

**Pre-commit hook:**
```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/CosmoLabs-org/cosmo-smoke
    rev: v0.6.0
    hooks:
      - id: smoke
```

## Quick Start

```bash
# 1. Generate a config for your project
smoke init

# 2. Run all tests
smoke run

# 3. Run only tagged tests
smoke run --tag build

# 4. CI-friendly JSON output
smoke run --format json

# 5. Watch mode — re-run on file changes
smoke run --watch
```

## Example .smoke.yaml

```yaml
version: 1
project: my-api
description: "Smoke tests for my-api"

settings:
  timeout: 30s
  fail_fast: true

prerequisites:
  - name: "Go installed"
    check: "go version"
    hint: "Install Go from https://go.dev"

tests:
  - name: "Compiles"
    run: "go build -o ./bin/api ./..."
    expect:
      exit_code: 0
    tags: [build]
    timeout: 60s
    cleanup: "rm -f ./bin/api"

  - name: "Health endpoint responds"
    run: "echo check"
    expect:
      http: { url: "http://localhost:8080/health", status_code: 200 }
    tags: [runtime]

  - name: "Redis is reachable"
    run: "echo check"
    expect:
      redis_ping: {}
    tags: [infra]

  - name: "API responds within 500ms"
    run: "echo check"
    expect:
      url_reachable: { url: "http://localhost:8080", timeout: 1s }
      response_time_ms: 500
    tags: [perf]

  - name: "Go version is 1.22+"
    run: "echo check"
    expect:
      version_check: { command: "go version", pattern: "go1\\.2[2-9]" }
    tags: [env]

  - name: "Config file exists"
    run: "echo check"
    expect:
      file_exists: "config.yaml"
    tags: [structure]

  - name: "Flaky external API"
    run: "curl -sf https://api.example.com/health"
    expect:
      exit_code: 0
    retry:
      count: 3
      backoff: 2s
    allow_failure: true
    skip_if:
      env_unset: "CI"
```

## Assertion Types

All assertions are optional and combinable within a single `expect` block.

### Process Assertions

| Type | Field | Description |
|------|-------|-------------|
| Exit code | `exit_code: <int>` | Exact process exit code match |
| Stdout substring | `stdout_contains: <string>` | Substring present in stdout |
| Stdout regex | `stdout_matches: <string>` | Go regex match against stdout |
| Stderr substring | `stderr_contains: <string>` | Substring present in stderr |
| Stderr regex | `stderr_matches: <string>` | Go regex match against stderr |
| Response time | `response_time_ms: <int>` | Test duration must not exceed threshold (ms) |

### File & Environment

| Type | Field | Description |
|------|-------|-------------|
| File exists | `file_exists: <path>` | Path exists relative to config file directory |
| Env variable | `env_exists: <string>` | Environment variable is set (non-empty) |

### Network Assertions

| Type | Field | Description |
|------|-------|-------------|
| Port listening | `port_listening: {port, protocol?, host?}` | TCP/UDP port is open |
| Process running | `process_running: <string>` | Named process is running (pgrep -x / tasklist) |
| HTTP check | `http: {url, method?, status_code?, body_contains?, body_matches?, header_contains?}` | Full HTTP endpoint validation |
| URL reachable | `url_reachable: {url, timeout?, status_code?}` | Lightweight HTTP/HTTPS connectivity check |
| Service reachable | `service_reachable: {url, timeout?}` | External service dependency check |
| SSL certificate | `ssl_cert: {host, port?, min_days_remaining?, allow_self_signed?}` | TLS cert validity + expiry threshold |
| gRPC health | `grpc_health: {address, service?, use_tls?, timeout?}` | grpc.health.v1 Health/Check returns SERVING |

### Database & Protocol

| Type | Field | Description |
|------|-------|-------------|
| Redis | `redis_ping: {host?, port?, password?}` | Redis PING returns +PONG |
| Memcached | `memcached_version: {host?, port?}` | Memcached `version` returns VERSION |
| PostgreSQL | `postgres_ping: {host?, port?}` | Postgres SSLRequest handshake valid |
| MySQL | `mysql_ping: {host?, port?}` | MySQL v10 handshake packet valid |

### Storage & Docker

| Type | Field | Description |
|------|-------|-------------|
| S3 bucket | `s3_bucket: {bucket, region?, endpoint?}` | S3-compatible bucket accessibility (anonymous HEAD) |
| Docker container | `docker_container_running: {name}` | Named Docker container is running |
| Docker image | `docker_image_exists: {image}` | Docker image exists locally |

### Tool Verification

| Type | Field | Description |
|------|-------|-------------|
| Version check | `version_check: {command, pattern}` | Shell command output matches regex |
| JSON field | `json_field: {path, equals?, contains?, matches?}` | JSONPath assertion on stdout |

### Test Modifiers

| Modifier | Description |
|----------|-------------|
| `allow_failure: true` | Test passes even if assertions fail (for flaky/optional checks) |
| `retry: {count, backoff}` | Retry flaky tests with exponential backoff |
| `skip_if: {env_unset, env_equals, file_missing}` | Conditionally skip a test |
| `tags: [...]` | Tag tests for selective runs |
| `timeout: <dur>` | Per-test timeout override |

## CLI Reference

```
smoke run [flags]
  -f, --file string          Config file (default ".smoke.yaml")
      --tag strings          Run only tests with these tags
      --exclude-tag strings  Skip tests with these tags
      --format string        Output format: terminal|json|junit|tap|prometheus (default "terminal")
      --fail-fast            Stop on first failure
      --timeout string       Per-test timeout override (e.g. "30s")
      --dry-run              List matching tests without running them
      --watch                Re-run tests on file changes

smoke init [flags]
  -f, --force                Overwrite existing .smoke.yaml

smoke version
```

## Auto-Detection

`smoke init` inspects the current directory and generates a starter config:

| Marker file | Detected type | Tests generated |
|-------------|---------------|----------------|
| `go.mod` | Go | build, vet, short tests |
| `package.json` | Node (bun/npm) | install, lint (if script exists) |
| `pyproject.toml` / `requirements.txt` | Python | import check |
| `Cargo.toml` | Rust | build, test |
| `Dockerfile` | Docker | docker build |

## Output Formats

`smoke run --format X` supports: `terminal` (default), `json`, `junit`, `tap`, `prometheus`.

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All tests passed |
| `1` | One or more tests failed |
| `2` | Config error or invalid arguments |

## License

MIT
