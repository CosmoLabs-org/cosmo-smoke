---
created: "2026-04-16"
status: approved
tags:
  - research
  - competitive-analysis
  - expansion
  - grok
title: Feature Expansion Research — Grok Competitive Analysis
source: Grok AI research + Claude synthesis
---

# Feature Expansion Research

## Research Source

Grok AI conducted competitive analysis on 2026-04-15, examining:
- Goss (5.9k stars) — closest competitor, stalled development
- Terratest (7.9k stars) — Go library for cloud/infra
- Karate (8.8k stars) — API testing
- Testinfra (2.5k stars) — Python infra testing
- Robot Framework (11.6k stars) — keyword-driven

Full research saved: `docs/research/2026-04-15-grok-competitive-analysis.md`

## Strategic Opportunity

**Goss development has stalled** (no major release in 18+ months). We can be the modern successor:
- Universal (not Linux-centric)
- Multi-language scaffolding
- Container-native design
- Active development

## Feature Categories

### Tier 1: Critical (v1.2)
| Feature | Why | Status |
|---------|-----|--------|
| `smoke serve` | Health endpoint for Docker/K8s | ROAD-019 |
| HTTP + JSONPath assertions | 90% of apps need this | ROAD-006/007 |
| Config inheritance | DRY configs, env overrides | ROAD-020 |

### Tier 2: Important (v2.0)
| Feature | Why | Status |
|---------|-----|--------|
| Port listening | Network health | FEAT-004 |
| Process running | Daemon checks | FEAT-005 |
| TAP output | CI compatibility | FEAT-006 |
| allow_failure | Flaky test handling | FEAT-007 |
| Auto-generate | Magic like `goss autoadd` | ROAD-023 |

### Tier 3: Differentiation
| Feature | Why | Status |
|---------|-----|--------|
| Goss migration tool | Capture users | ROAD-024 |
| Prometheus output | Monitoring integration | IDEA |
| WebSocket assertions | Real-time apps | IDEA |
| GraphQL assertions | Modern APIs | IDEA |
| gRPC health checks | Microservices | IDEA |
| Cloud storage (S3/GCS) | Cloud-native | IDEA |
| SSL cert validation | Prevent outages | IDEA |
| Response time thresholds | Performance smoke | IDEA |

## Ideas Captured

13 ideas filed in `docs/ideas/`:
1. Prometheus metrics output
2. Claude Desktop MCP extension
3. OpenTelemetry trace correlation
4. Pre-commit hook integration
5. Portfolio smoke dashboard
6. WebSocket assertion
7. GraphQL introspection
8. gRPC health check
9. Redis/Memcached connectivity
10. S3/Cloud storage
11. SSL certificate validation
12. Response time threshold
13. Mobile deep link assertion

## GLM-Parallelizable Tasks

4 bounded tasks ready for GLM dispatch:
- FEAT-004: Port listening assertion
- FEAT-005: Process running assertion
- FEAT-006: TAP output format
- FEAT-007: allow_failure for flaky tests

Manifest: `docs/prompts/2026-04-16-glm-dispatch-assertions.yaml`

## Container Health Check Integration

### Docker
```dockerfile
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD ["smoke", "run", "--health-mode"]
```

### Kubernetes
```yaml
livenessProbe:
  exec:
    command: ["smoke", "run", "--health-mode"]
readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
```

## Next Actions

1. **Dispatch GLM agents** for FEAT-004/005/006/007 (parallel)
2. **Implement `smoke serve`** (ROAD-019) — critical for adoption
3. **Add HTTP assertions** (ROAD-006) — core functionality
4. **Ship v1.2** with serve + HTTP + config inheritance
