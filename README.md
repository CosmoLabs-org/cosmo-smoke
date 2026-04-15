# cosmo-smoke

Universal smoke test runner. Reads `.smoke.yaml` from any project root and runs lightweight smoke tests.

## Install

```bash
go install github.com/CosmoLabs-org/cosmo-smoke@latest
```

## Quick Start

```bash
smoke init        # Auto-detect project type, generate .smoke.yaml
smoke run         # Run all smoke tests
smoke run --tag build   # Run only build-tagged tests
smoke run --format json # JSON output for CI
```

## Config Schema

```yaml
version: 1
project: my-app
description: "Smoke tests for my app"

settings:
  timeout: 30s       # Default per-test timeout
  fail_fast: true    # Stop on first failure
  parallel: false    # Run tests concurrently

prerequisites:
  - name: "Go installed"
    check: "go version"
    hint: "Install Go from https://go.dev"

tests:
  - name: "Compiles"
    run: "go build ./..."
    expect:
      exit_code: 0
    tags: [build]
    timeout: 60s
    cleanup: "rm -f ./binary"

  - name: "Help output"
    run: "./binary --help"
    expect:
      exit_code: 0
      stdout_contains: "usage"
      stdout_matches: "^Usage:"
    tags: [runtime]

  - name: "Error handling"
    run: "./binary --invalid 2>&1"
    expect:
      stderr_contains: "unknown flag"

  - name: "Config exists"
    run: "echo check"
    expect:
      file_exists: "config.yaml"
    tags: [structure]
```

## Assertion Types

| Type | Description |
|------|-------------|
| `exit_code` | Process exit code (integer) |
| `stdout_contains` | Substring match on stdout |
| `stdout_matches` | Regex match on stdout |
| `stderr_contains` | Substring match on stderr |
| `file_exists` | File exists (relative to config dir) |

## CLI Reference

```
smoke run [flags]
  -f, --file string       Config file path (default ".smoke.yaml")
      --tag strings       Include only tests with these tags
      --exclude-tag strings  Exclude tests with these tags
      --format string     Output format: terminal|json (default "terminal")
      --fail-fast         Stop on first failure
      --timeout string    Per-test timeout override (e.g. "30s")
      --dry-run           List tests without running

smoke init [flags]
  -f, --force             Overwrite existing .smoke.yaml

smoke version
```

## Auto-Detection

`smoke init` detects project types and generates appropriate tests:

| Marker | Type | Tests Generated |
|--------|------|-----------------|
| `go.mod` | Go | build, test |
| `package.json` | Node | install, lint (if available) |
| `pyproject.toml` | Python | import check |
| `Cargo.toml` | Rust | build, test |
| `Dockerfile` | Docker | docker build |

## License

MIT - CosmoLabs
