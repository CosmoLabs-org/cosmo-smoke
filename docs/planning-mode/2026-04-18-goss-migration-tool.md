---
title: Goss Migration Tool — Implementation Plan
created: 2026-04-18
status: APPROVED
roadmap: ROAD-024
branch: master
---

# Goss Migration Tool — Implementation Plan

## Objective

Ship `smoke migrate goss <input.yaml>` that converts Goss YAML to cosmo-smoke `.smoke.yaml` format. Core 7 keys with full fidelity, long-tail keys as TODO stubs.

## Resolved Decisions

1. `--distro=deb|rpm|apk` flag (default deb)
2. `command:` fallback + TODO comments acceptable for v1
3. Flatten includes into single output file
4. No reverse migration in v0.5
5. Per-file only (no directory walking)

## File Scope

### New files
- `cmd/migrate.go` — Cobra `migrate` parent + `goss` subcommand
- `internal/migrate/goss/parser.go` — GossFile struct, YAML parsing
- `internal/migrate/goss/translator.go` — per-key translate functions
- `internal/migrate/goss/emitter.go` — string-builder YAML emitter
- `internal/migrate/goss/parser_test.go`
- `internal/migrate/goss/translator_test.go`
- `internal/migrate/goss/emitter_test.go`
- `internal/migrate/goss/testdata/` — golden file fixtures

### Modified files
- `cmd/root.go` — wire `migrateCmd`

## Execution Model

Two parallel Sonnet worktrees via `ccs spawn`:

### Worktree A — Parser + Translator (no external deps)

**Scope:**
- `internal/migrate/goss/parser.go`
  - `GossFile` struct with all top-level keys as `map[string]GossResource`
  - `GossResource` = `map[string]interface{}` (flexible per-key attrs)
  - `Parse(data []byte) (*GossFile, error)`
  - Include gossfile resolution (flatten into one GossFile)
- `internal/migrate/goss/translator.go`
  - `Translate(gf *GossFile, opts TranslateOptions) ([]schema.Test, []TranslationWarning)`
  - Per-key functions: `translatePackage`, `translateService`, `translateProcess`, `translatePort`, `translateCommand`, `translateFile`, `translateHTTP`
  - Long-tail stubs: `translateDNS`, `translateAddr`, `translateInterface`, `translateMount`, `translateKernelParam`
  - `TranslateOptions` includes `Distro string`
- `internal/migrate/goss/testdata/`
  - `goss/basic.yaml` — core 7 keys
  - `goss/longtail.yaml` — dns, addr, interface, mount, kernel-param
  - `goss/includes.yaml` — gossfile references
  - `smoke/basic.expected.yaml`
  - `smoke/longtail.expected.yaml`
  - `smoke/includes.expected.yaml`
- Unit tests: golden-file comparison for each translator

**Acceptance:**
- All 7 core keys translate with full fidelity
- Long-tail keys produce TODO stubs (no panics)
- gossfile includes get flattened
- `go test ./internal/migrate/goss/...` passes

### Worktree B — Emitter + CLI (depends on translator types)

**Scope:**
- `internal/migrate/goss/emitter.go`
  - `Emit(tests []schema.Test, warnings []TranslationWarning, meta EmitMeta) (string, error)`
  - String-builder approach (Option A from design doc)
  - Header comment with source, timestamp, stats
  - Per-test comment blocks preserving TODO stubs
- `cmd/migrate.go`
  - `migrateCmd` parent command
  - `gossCmd` subcommand with positional arg
  - Flags: `-o/--output`, `--overwrite`, `--strict`, `--stats`, `--distro`
  - Reads input file, parses, translates, emits, writes output
  - Non-zero exit on `--strict` when warnings exist
- `cmd/root.go` — add `migrateCmd` to root
- Emitter tests + CLI integration test

**Acceptance:**
- `smoke migrate goss testdata/goss/basic.yaml` outputs valid YAML
- `schema.Load` parses emitted output without errors
- `--stats` prints summary to stderr
- `--strict` exits non-zero for files with TODO stubs
- `--distro rpm` changes package commands

**Merge order:** A first (B imports translator types), then B.

## Post-Merge: Complementary Features

After ROAD-024 merges, implement in master:

1. **ROAD-008 (skip_if):** Add `skip_if` to schema, evaluate in runner, SKIPPED status in reporter
2. **ROAD-017 (multi-env):** Add `--env` flag, deep-merge env overrides onto base config

Both are small, single-session work. No worktrees needed.
