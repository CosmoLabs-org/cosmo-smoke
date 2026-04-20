# TODO: Increase Unit Test Coverage

## Goals

- [x] Goal 1: dashboard package — test handler.go (RegisterRoutes, handleResults, handleProjects, handleProjectHistory) — 37.0% → 83.2%
- [x] Goal 2: schema package — test MarshalYAML, LoadDefault, Validate error paths — 69.8% → 97.3%
- [x] Goal 3: detector package — container.go functions all require docker exec — skipped (external dep) — remains 64.6%
- [x] Goal 4: mcp package — test generateExpectBlock, boolArg, helpers — 62.0% → 89.0%
- [x] Goal 5: runner package — test RunMonorepo, port listening, assertions — 70.7% → 75.1%
- [x] Goal 6: Run final verification with -race and confirm coverage delta

## Final Coverage Summary

| Package | Before | After | Delta |
|---------|--------|-------|-------|
| dashboard | 37.0% | 83.2% | +46.2% |
| schema | 69.8% | 97.3% | +27.5% |
| mcp | 62.0% | 89.0% | +27.0% |
| runner | 70.7% | 75.1% | +4.4% |
| detector | 64.6% | 64.6% | +0% (external dep) |
| reporter | 87.9% | 87.9% | (unchanged) |
| monorepo | 95.5% | 95.5% | (unchanged) |

All new tests pass with -race detector. No anti-patterns found.

## Untested branches (follow-up)

- detector: container.go — all functions require Docker (exec.Command)
- runner: CheckDockerContainerRunning/CheckDockerImageExists — require Docker
- runner: CheckGRPCHealthWithTrace — requires gRPC build tag
- runner: queryJaeger — requires running Jaeger service
- mcp: ServeStdio — requires stdin/stdout
- mcp: adaptHandler — requires mcp-go internal plumbing
- cmd: most functions — heavy Cobra/exec dependencies
