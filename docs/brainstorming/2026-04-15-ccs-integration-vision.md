---
created: "2026-04-15"
status: approved
tags:
  - integration
  - ccs
  - strategy
title: CCS Integration Vision — cosmo-smoke as Universal Smoke Testing Tool
---

# CCS Integration Vision

## Core Principle

**cosmo-smoke is 100% standalone.** It has zero dependencies on CCS. CCS provides integration hooks that call smoke as an external tool — same relationship as CCS has with `go`, `git`, or `gh`.

```
┌─────────────────────────────────────────────┐
│  cosmo-smoke (standalone)                   │
│  - Repo: CosmoLabs-org/cosmo-smoke          │
│  - Binary: ~/bin/smoke                      │
│  - Config: .smoke.yaml                      │
│  - Zero CCS dependencies                    │
│  - Can be open-sourced independently        │
└─────────────────────────────────────────────┘
                    ▲
                    │ calls via Bash
                    │
┌─────────────────────────────────────────────┐
│  CCS Integration Layer                      │
│  - ccs smoke        → wrapper command       │
│  - /project-init    → offers smoke init     │
│  - /health-check    → verifies smoke passes │
│  - /audit           → flags missing .smoke  │
│  - SessionStart     → optional auto-smoke   │
│  - /session-end     → smoke before commit   │
└─────────────────────────────────────────────┘
```

## Why Smoke Tests Matter

| Aspect | Benefit |
|--------|---------|
| **Token efficiency** | Go binary runs via Bash — zero tokens for execution |
| **Deterministic** | Same config = same result, no AI variance |
| **Fast** | 2-5 seconds typical, 10s budget max |
| **First gate** | Catches "completely broken" before wasting time |

**The philosophy:** Smoke tests answer "does it turn on?" — not "does it work correctly?" They're the prerequisite for deeper testing.

## SOP Connections

### Dependencies SOP

Smoke prerequisites verify tools are installed:

```yaml
prerequisites:
  - name: "Go installed"
    check: "go version"
    hint: "Install Go 1.25+: https://go.dev/dl/"
  - name: "Bun installed"
    check: "bun --version"
```

Future: `version_check` assertion type (ROAD-013) that validates versions match requirements.

### Credentials SOP

`env_exists` assertion verifies required credentials are set:

```yaml
tests:
  - name: "API keys configured"
    run: "true"  # No-op command
    expect:
      env_exists: "STRIPE_SECRET_KEY"
```

Future: Generate credential checks from `~/.credentials/PROJECT/project.env` automatically (ROAD-014).

### Tech Stack

`smoke init` detects project type and generates appropriate tests:

| Detected | Tests Generated |
|----------|-----------------|
| `go.mod` | `go build`, `go vet`, `go test -short` |
| `package.json` + `bun.lock` | `bun install`, `bun build`, `bun lint` |
| `Dockerfile` | `docker build .` |
| `Cargo.toml` | `cargo check`, `cargo test --no-run` |

## CCS Integration Points

### 1. `ccs smoke` wrapper

```bash
ccs smoke           # → smoke run
ccs smoke init      # → smoke init
ccs smoke --tag X   # → smoke run --tag X
```

Passes through all flags. Adds project context if needed.

### 2. `/project-init`

After scaffolding, offer:
```
Smoke tests: Run 'smoke init' to generate .smoke.yaml? [Y/n]
```

### 3. `/health-check`

Check items:
- [ ] `.smoke.yaml` exists
- [ ] `smoke run --dry-run` validates config
- [ ] `smoke run` passes (optional, can be slow)

### 4. `/audit`

Audit finding: "Missing smoke tests" if no `.smoke.yaml` in project root.

### 5. SessionStart hook

Optional config in `settings.json`:
```json
{
  "hooks": {
    "SessionStart": ["smoke run || echo 'Smoke tests failing'"]
  }
}
```

### 6. `superpowers:systematic-debugging`

Add as Step 0:
```
0. Run smoke tests first
   - If smoke fails, that's likely the bug
   - Fix smoke before investigating deeper
```

## Roadmap Summary

### v1.1 (Implemented)
- [x] FEAT-001: stderr_matches
- [x] FEAT-002: env_exists

### v1.2 (CI/DX)
- [ ] FEAT-003: JUnit XML output
- [ ] ROAD-003: Watch mode (--watch)
- [ ] ROAD-005: GitHub Actions reusable workflow

### v2.0 (Advanced)
- [ ] ROAD-006: HTTP endpoint assertions
- [ ] ROAD-007: JSON field assertions
- [ ] ROAD-008: Conditional test execution
- [ ] ROAD-009: allow_failure for flaky tests
- [ ] ROAD-010: Monorepo sub-config support
- [ ] ROAD-011: TAP output format
- [ ] ROAD-012: Retry with backoff
- [ ] ROAD-013: Dependency version assertions
- [ ] ROAD-014: Credential smoke tests
- [ ] ROAD-015: Docker container smoke tests
- [ ] ROAD-016: Database connectivity assertions
- [ ] ROAD-017: Multi-environment configs
- [ ] ROAD-018: Service dependency checks

## Feedback Sent to CCS

| ID | Title |
|----|-------|
| FB-453 | Add cosmo-smoke as external tool dependency |
| FB-454 | Study global integration for all projects |
| FB-455 | Full CCS integration spec |
| FB-456 | Study cosmo-smoke for universal integration |

## Next Steps

1. Push cosmo-smoke to GitHub (`/move-to-github`)
2. Implement FEAT-003 (JUnit XML) for CI integration
3. Create GitHub Actions reusable workflow (ROAD-005)
4. CCS team reviews FB-453/454/455/456 and plans integration
