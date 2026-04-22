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
| `--report-url` | (none) | POST results JSON to this URL after run |
| `--report-api-key` | (none) | API key for report-url (sent as `X-API-Key` header) |
| `--baseline` | `false` | Save and compare test timings against baseline |
| `--baseline-threshold` | `50` | Regression threshold % (flag if current > baseline * (1+threshold/100)) |

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
smoke run --baseline                          # Track performance regressions
smoke run --baseline --baseline-threshold 25  # Tighter regression threshold
smoke run --report-url https://hooks.example.com/smoke  # Push results
```

## Watch Mode

With `--watch`, smoke stays resident and re-runs tests on file changes. Uses fsnotify with a 500ms debounce. Press Ctrl+C to exit.

## Monorepo Mode

With `--monorepo`, smoke discovers `.smoke.yaml` files in subdirectories and runs each as a separate test suite. Respects `settings.monorepo_exclude` for ignoring directories.

## Baseline Tracking

With `--baseline`, smoke saves test durations to `.smoke-baseline.json` and compares each run against the saved baseline. Tests that exceed `baseline-threshold`% slower are flagged as regressions. New tests are reported. The baseline file is updated after each run.

## Push Reporting

With `--report-url`, smoke POSTs a JSON summary of results to the given URL after each run. Use `--report-api-key` to authenticate via the `X-API-Key` header.
