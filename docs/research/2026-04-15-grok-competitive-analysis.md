---
title: Competitive Analysis — Grok Research
created: "2026-04-15"
source: Grok AI
tags: [research, competitive, goss, tools]
---

# Cosmo-Smoke Competitive Analysis (Grok Research)

## Executive Summary

**Goss** (5.9k stars) is our closest competitor — battle-tested since 2016 but development stalled (no major release in 18+ months). Cosmo-smoke can be the modern universal successor.

## Tool Comparison

| Tool | Stars | Strengths | Weaknesses | Steal From |
|------|-------|-----------|------------|------------|
| **Goss** | 5.9k | YAML config, 15+ resources, `serve` mode, `autoadd` | Linux-centric, stalled dev | `serve`, includes/vars |
| **Terratest** | 7.9k | Go library, cloud APIs, Docker, K8s | Requires Go code | Cloud assertion patterns |
| **Testinfra** | 2.5k | Python/pytest, file/package/service | No longer maintained | Assertion types |
| **Karate** | 8.8k | BDD API testing, JSONPath, mocks | Java-first, heavy | JSON assertions |
| **Bats** | 6k | Simple Bash/TAP | No YAML, shell-only | stdout/stderr patterns |
| **Robot Framework** | 11.6k | Keyword-driven, huge ecosystem | Steep learning curve | Plugin model |

## Assertions We're Missing

| Type | Priority | Implementation |
|------|----------|----------------|
| HTTP (status, headers, body, JSONPath) | P0 | `http` resource |
| Port listening | P1 | `port` resource |
| Process running | P1 | `process` resource |
| Service status | P2 | `service` resource |
| DNS resolution | P2 | `dns` resource |
| Package installed | P3 | `package` resource |
| File content matching | P3 | Extend `file_exists` |

## Critical Features to Ship

### 1. `smoke serve` (P0)
```bash
smoke serve --port 8080 --path /healthz
```
- HTTP endpoint that runs tests on-demand
- Returns 200 (pass) or 503 (fail) + JSON
- Perfect for Docker HEALTHCHECK and K8s probes

### 2. Config Inheritance (P1)
```yaml
includes:
  - base.smoke.yaml
vars:
  from_file: vars.${ENV}.yaml
```

### 3. Auto-generate (P2)
```bash
smoke init --from-running  # Magic like goss autoadd
```

## Container Health Check Integration

### Docker HEALTHCHECK
```dockerfile
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD ["smoke", "run", "--health-mode"] || exit 1
```

### Kubernetes Probes
```yaml
livenessProbe:
  exec:
    command: ["smoke", "run", "--health-mode"]
readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
```

## Strategic Positioning

**Where Goss wins today:**
- Mature, rock-solid
- `autoadd` magic
- `serve` mode
- 15+ resource types

**Where we win:**
- Universal (not Linux-centric)
- Multi-language scaffolding
- Active development
- Modern CI/CD focus
- Container-native design

**Our path to victory:**
1. Ship `serve` mode (Goss parity)
2. Add HTTP + JSONPath assertions
3. Implement includes/vars
4. Launch "Goss → cosmo-smoke" migration tool
5. Market to Goss users (stalled development)

## Roadmap Items Created

- ROAD-019: `smoke serve` — HTTP health endpoint (P95)
- ROAD-020: Config inheritance — includes/vars/templates (P85)
- ROAD-021: Port listening assertion (P70)
- ROAD-022: Process running assertion (P65)
- ROAD-023: Auto-generate from running container (P60)
- ROAD-024: Goss migration tool (P50)
