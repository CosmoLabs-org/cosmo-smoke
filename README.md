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

  - name: "Help flag works"
    run: "./bin/api --help"
    expect:
      exit_code: 0
      stdout_contains: "usage"
    tags: [runtime]

  - name: "Rejects bad flags"
    run: "./bin/api --invalid-flag"
    expect:
      exit_code: 1
      stderr_contains: "unknown flag"
    tags: [runtime]

  - name: "Config file exists"
    run: "echo check"
    expect:
      file_exists: "config.yaml"
    tags: [structure]
```

## Assertion Types

All assertions are optional and combinable within a single `expect` block.

| Type | Field | Description |
|------|-------|-------------|
| Exit code | `exit_code: <int>` | Exact process exit code match |
| Stdout substring | `stdout_contains: <string>` | Substring present in stdout |
| Stdout regex | `stdout_matches: <string>` | Go regex match against stdout |
| Stderr substring | `stderr_contains: <string>` | Substring present in stderr |
| File existence | `file_exists: <path>` | Path exists relative to config file directory |

## CLI Reference

```
smoke run [flags]
  -f, --file string          Config file (default ".smoke.yaml")
      --tag strings          Run only tests with these tags
      --exclude-tag strings  Skip tests with these tags
      --format string        Output format: terminal|json (default "terminal")
      --fail-fast            Stop on first failure
      --timeout string       Per-test timeout override (e.g. "30s")
      --dry-run              List matching tests without running them

smoke init [flags]
  -f, --force                Overwrite existing .smoke.yaml

smoke version
```

## Auto-Detection

`smoke init` inspects the current directory and generates a starter config:

| Marker file | Detected type | Tests generated |
|-------------|---------------|-----------------|
| `go.mod` | Go | build, vet, short tests |
| `package.json` | Node | install, lint (if script exists) |
| `pyproject.toml` / `requirements.txt` | Python | import check |
| `Cargo.toml` | Rust | build, test |
| `Dockerfile` | Docker | docker build |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All tests passed |
| `1` | One or more tests failed |
| `2` | Config error or invalid arguments |

## License

MIT
