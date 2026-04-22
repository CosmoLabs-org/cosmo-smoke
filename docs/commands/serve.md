# smoke serve

Start an HTTP health endpoint that runs smoke tests on each request. Designed for container liveness/readiness probes. Supports optional dashboard mode for portfolio-wide test aggregation.

## Usage

```bash
smoke serve [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-p, --port` | `8080` | Port to listen on |
| `--path` | `/healthz` | Health endpoint path |
| `-f, --file` | `.smoke.yaml` | Config file path |
| `--dashboard` | `false` | Enable dashboard aggregation mode |
| `--api-key` | (none) | API key for `POST /api/results` (`X-API-Key` header) |
| `--db-path` | `smoke-dashboard.db` | SQLite database path for dashboard storage |

## Response

Each request runs the full smoke suite and returns JSON:

```json
{
  "status": "healthy",
  "tests": {
    "total": 6,
    "passed": 6,
    "failed": 0
  },
  "duration_ms": 142
}
```

- All tests pass → `200 OK` with `"status": "healthy"`
- Any test fails → `503 Service Unavailable` with `"status": "unhealthy"`
- Config error → `500 Internal Server Error`

Graceful shutdown on SIGINT/SIGTERM with a 5s timeout.

## Dashboard Mode

With `--dashboard`, serve enables additional endpoints for portfolio-wide test aggregation:

- `POST /api/results` — Accept test results from remote smoke runs (requires `--api-key`)
- `GET /dashboard` — Embedded web dashboard showing historical results

Results are stored in a SQLite database (configurable with `--db-path`).

## Examples

## Examples

```bash
smoke serve                            # Default: :8080/healthz
smoke serve -p 9090 --path /ready      # Custom port and path
smoke serve -f /etc/smoke/config.yaml  # Custom config path
```
