# smoke run

Execute smoke tests defined in `.smoke.yaml`.

## Usage

```bash
smoke run [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --file` | `.smoke.yaml` | Config file path |
| `--tag` | (none) | Include only tests with these tags (repeatable) |
| `--exclude-tag` | (none) | Exclude tests with these tags (repeatable) |
| `--format` | `terminal` | Output format: `terminal`, `json`, `junit`, `tap`, `prometheus` |
| `--fail-fast` | `false` | Stop on first failure |
| `--timeout` | (none) | Per-test timeout override (e.g. `30s`) |
| `--dry-run` | `false` | List tests without running |
| `--watch` | `false` | Re-run tests when files change (Ctrl+C to exit) |
| `--env` | (none) | Load environment-specific config (e.g. `staging` loads `staging.smoke.yaml`) |
| `--monorepo` | `false` | Auto-discover `.smoke.yaml` in subdirectories |
| `--otel-collector` | (none) | Override `otel.jaeger_url` and enable tracing |
| `--no-otel` | `false` | Disable OTel trace propagation for this run |

## Examples

```bash
smoke run                              # Run all tests in .smoke.yaml
smoke run -f staging.smoke.yaml        # Use a specific config file
smoke run --tag critical               # Only tests tagged "critical"
smoke run --exclude-tag slow           # Skip tests tagged "slow"
smoke run --format json                # JSON output for CI
smoke run --fail-fast --timeout 10s    # Fail fast with 10s per test
smoke run --watch                      # Auto re-run on file changes
smoke run --monorepo                   # Run sub-project configs
smoke run --env staging                # Merge staging.smoke.yaml overrides
smoke run --otel-collector http://jaeger:16686  # Enable OTel tracing
```

## Watch Mode

With `--watch`, smoke stays resident and re-runs tests on file changes. Uses fsnotify with a 500ms debounce. Press Ctrl+C to exit.

## Monorepo Mode

With `--monorepo`, smoke discovers `.smoke.yaml` files in subdirectories and runs each as a separate test suite. Respects `settings.monorepo_exclude` for ignoring directories.
