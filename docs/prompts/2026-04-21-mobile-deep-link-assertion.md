---
title: "Mobile Deep Link Assertion — Implementation"
created: "2026-04-21"
status: PENDING
priority: high
branch: master
origin: "/brainplan"
tags: [continuation, implementation, mobile, deep-link]
goals_total: 7
goals_completed: 0
brainstorm_ref: docs/brainstorming/2026-04-21-mobile-deep-link-assertion.md
plan_ref: docs/planning-mode/2026-04-21-mobile-deep-link-assertion.md
requires_reading:
    - docs/brainstorming/2026-04-21-mobile-deep-link-assertion.md
    - docs/planning-mode/2026-04-21-mobile-deep-link-assertion.md
schema_version: 1
---

# Mobile Deep Link Assertion — Implementation

## Context

Add a two-tier progressive `deep_link` assertion type for validating mobile deep link configuration and resolution. Tier 1 uses zero-dep HTTP/config checks; Tier 2 uses `adb`/`xcrun` when available. Covers Android, iOS, React Native, and Flutter projects.

Design spec: `docs/brainstorming/2026-04-21-mobile-deep-link-assertion.md`
Implementation plan: `docs/planning-mode/2026-04-21-mobile-deep-link-assertion.md`

## Goals

- [ ] G-01: Add DeepLink struct to schema (Task 1)
- [ ] G-02: Add mobile project types to detector (Task 2)
- [ ] G-03: Implement assetlinks.json and AASA HTTP validation (Task 3)
- [ ] G-04: Implement CheckDeepLink main entry with tier routing (Task 4)
- [ ] G-05: Wire deep_link assertion into runner.go (Task 5)
- [ ] G-06: Add mobile project smoke init templates (Task 6)
- [ ] G-07: Update CLAUDE.md, run full test suite (Task 7)

## Execution Strategy

Sequential — tasks have dependencies (schema → assertion logic → runner wiring → templates).

```
agents:
  - task: "Schema + detector types (Tasks 1-2)"
    model: sonnet
    files: [internal/schema/schema.go, internal/detector/detector.go]
    ready: true
  - task: "Tier 1 HTTP checks (Task 3)"
    model: sonnet
    files: [internal/runner/assertion_deeplink.go]
    ready: after tasks 1-2
  - task: "Main entry + tier routing (Task 4)"
    model: opus
    files: [internal/runner/assertion_deeplink.go]
    ready: after task 3
  - task: "Runner wiring + templates + docs (Tasks 5-7)"
    model: sonnet
    files: [internal/runner/runner.go, internal/detector/templates.go, CLAUDE.md]
    ready: after task 4
```

## File Scope

```
internal/runner/assertion_deeplink.go       # NEW
internal/runner/assertion_deeplink_test.go  # NEW
internal/schema/schema.go                   # MODIFY
internal/detector/detector.go               # MODIFY
internal/detector/templates.go              # MODIFY
internal/runner/runner.go                   # MODIFY
CLAUDE.md                                   # MODIFY
```
