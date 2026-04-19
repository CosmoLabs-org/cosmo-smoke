---
branch: master
completed: "2026-04-18"
created: "2026-04-18"
goals_completed: 6
goals_total: 6
origin: /brainplan
priority: high
related_prompts:
    - docs/brainstorming/2026-04-18-v0-6-connect-and-verify.md
    - docs/planning-mode/2026-04-18-v0-6-connect-and-verify.md
status: COMPLETED
tags:
    - continuation
    - implementation
    - v0.6
title: cosmo-smoke v0.6 — Connect and Verify
---

# cosmo-smoke v0.6 — Connect and Verify

## Context

v0.5 shipped the Goss migration tool (ROAD-024), conditional execution (ROAD-008), and multi-env configs (ROAD-017). v0.6 expands assertion coverage — the "universal smoke test" needs to answer "does it connect to everything it needs?" Four new assertion types plus a pre-commit hook integration.

Design spec: `docs/brainstorming/2026-04-18-v0-6-connect-and-verify.md`
Implementation plan: `docs/planning-mode/2026-04-18-v0-6-connect-and-verify.md`

## Goals

### [x] 1. Add schema structs + validation for 4 new assertion types

Add `URLReachableCheck`, `ServiceReachableCheck`, `S3BucketCheck`, `VersionCheck` structs to schema.go. Add validation rules for URL format, required fields, and regex compilation. Wire into `Expect` struct.

### [x] 2. Implement httpReachable shared helper + url_reachable + service_reachable

Add `httpReachable` internal helper (stdlib net/http, timeout support, status code matching). Implement `CheckURLReachable` and `CheckServiceReachable`. 6 tests using httptest.

### [x] 3. Implement s3_bucket assertion

Add `CheckS3Bucket` using httpReachable helper with path-style URL construction. Anonymous HEAD only, explicit 403 hint about authentication. 3 tests.

### [x] 4. Implement version_check assertion

Add `CheckVersion` running shell command via exec, regex matching stdout. 3 tests covering match, no-match, command failure.

### [x] 5. Wire all assertions into runner + pre-commit hook

Add evaluation blocks in `runner.go runTestOnce()`. Create `.pre-commit-hooks.yaml`. Add validation test for hook file.

### [x] 6. Release v0.6.0

Update CLAUDE.md, run self-smoke, update changelog, bump version, tag and push.

## Execution Strategy

Parallel GLM agents for Tasks 2-4 (independent assertion implementations), sequential for Task 1 (schema dependency) and Task 5 (wiring dependency).

agents:
  - task: "Schema structs + validation"
    model: sonnet
    files: [internal/schema/schema.go]
    ready: true
  - task: "url_reachable + service_reachable assertions + tests"
    model: sonnet
    files: [internal/runner/assertion.go, internal/runner/assertion_reachable_test.go]
    ready: after task 1
  - task: "s3_bucket assertion + tests"
    model: sonnet
    files: [internal/runner/assertion.go, internal/runner/assertion_reachable_test.go]
    ready: after task 1
  - task: "version_check assertion + tests"
    model: sonnet
    files: [internal/runner/assertion.go, internal/runner/assertion_version_test.go]
    ready: after task 1
  - task: "Runner wiring + pre-commit hook"
    model: sonnet
    files: [internal/runner/runner.go, .pre-commit-hooks.yaml]
    ready: after tasks 2-4

## File Scope

```yaml
files_modified:
  - internal/schema/schema.go
  - internal/runner/runner.go
  - internal/runner/assertion.go
files_created:
  - internal/runner/assertion_reachable_test.go
  - internal/runner/assertion_version_test.go
  - .pre-commit-hooks.yaml
```
