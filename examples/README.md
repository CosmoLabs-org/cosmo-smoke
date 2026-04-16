# cosmo-smoke Examples

Production-quality `.smoke.yaml` configs for common project types.
Each example is self-contained and designed to be dropped into your project with minimal edits.

---

## Examples

| Directory | Stack | Covers |
|-----------|-------|--------|
| [`go-api/`](go-api/.smoke.yaml) | Go REST API | build, vet, short tests, binary, /health endpoint |
| [`node-fullstack/`](node-fullstack/.smoke.yaml) | Node.js (React + Express) | install, lint, typecheck, frontend build, server start |
| [`python-fastapi/`](python-fastapi/.smoke.yaml) | Python + FastAPI | install, ruff lint/format, mypy, pytest collect, uvicorn start |
| [`docker-compose/`](docker-compose/.smoke.yaml) | Multi-container Compose | build, up, per-service health checks, teardown |
| [`rust-cli/`](rust-cli/.smoke.yaml) | Rust CLI tool | cargo check, clippy, fmt, test compile, --help, --version |
| [`monorepo/`](monorepo/.smoke.yaml) | Monorepo (api + web + shared) | per-package tags, cross-package build order |
| [`kubernetes/`](kubernetes/.smoke.yaml) | Kubernetes deployment | namespace, rollout, endpoints, ingress, secrets, /health |

---

## Running an Example

```bash
# Dry-run (validates config + prints what would run — no commands executed)
smoke run --dry-run -f examples/go-api/.smoke.yaml

# Run all tests in an example
smoke run -f examples/go-api/.smoke.yaml

# Run only build-tagged tests
smoke run -f examples/monorepo/.smoke.yaml --tag build

# Run only api-tagged tests in the monorepo
smoke run -f examples/monorepo/.smoke.yaml --tag api

# Run with JSON output (for CI pipelines)
smoke run -f examples/docker-compose/.smoke.yaml --format json

# Stop on first failure
smoke run -f examples/kubernetes/.smoke.yaml --fail-fast
```

---

## Tag Conventions

These examples follow a consistent tagging convention:

| Tag | Used for |
|-----|----------|
| `build` | Compilation, asset generation, artifact checks |
| `lint` | Static analysis, formatting, type checking |
| `test` | Test execution (unit, short, collect) |
| `runtime` | Process startup, endpoint checks, live behavior |
| `infra` | Infrastructure state (Docker, Kubernetes, cloud) |
| `[package]` | Package-scoped tags in monorepos (e.g. `api`, `web`, `shared`) |

---

## Adapting an Example

1. Copy the relevant `.smoke.yaml` to your project root
2. Edit `project:` and `description:` to match your project
3. Adjust `run:` commands to match your actual scripts
4. Update prerequisites to match your required tools
5. Run `smoke run --dry-run` to validate before your first real run

---

## Further Reading

- [Schema reference](../SPEC.md) — full field documentation with types and defaults
- [CLI reference](../USAGE.md) — all flags for `smoke run`, `smoke init`, `smoke version`
- [Main README](../README.md) — project overview and quick start
