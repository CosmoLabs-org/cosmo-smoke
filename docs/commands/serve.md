# smoke serve

Start an HTTP health endpoint that runs smoke tests on each request. Designed for container liveness/readiness probes.

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

## Examples

```bash
smoke serve                            # Default: :8080/healthz
smoke serve -p 9090 --path /ready      # Custom port and path
smoke serve -f /etc/smoke/config.yaml  # Custom config path
```
