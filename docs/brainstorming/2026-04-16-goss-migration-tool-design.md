---
title: Goss migration tool — design doc
created: 2026-04-16
status: decisions-resolved
roadmap: ROAD-024
related:
  - docs/roadmap/items/ROAD-024.yaml
---

# Goss Migration Tool — Design

## Why this exists

Goss (https://github.com/goss-org/goss) is a server-testing tool with a YAML-based schema that overlaps heavily with cosmo-smoke's model. Goss development has been effectively stalled for 18+ months. Users looking for a maintained alternative are an organic audience.

**Strategic move:** `smoke migrate goss path/to/goss.yaml` one-shots a `.smoke.yaml` that captures as much of the Goss intent as stdlib cosmo-smoke assertions allow, with explicit `# TODO:` stubs for the things we don't cover natively.

## Goal

Ship `smoke migrate goss <input.yaml> [-o output.smoke.yaml]` that:

1. Parses a Goss YAML file (standard schema, any version)
2. Emits a `.smoke.yaml` with one `tests:` entry per Goss resource
3. Preserves intent even where semantics differ (use `command:` as fallback)
4. Documents unmapped assertions as inline TODO comments so users know what to hand-fix
5. Never silently drops a resource — every Goss entry becomes either a direct assertion or a flagged TODO

## Goss schema overview (what we're migrating FROM)

Goss's top-level keys, each holding a map of resource-name → attributes:

| Goss key | Example value | What it verifies |
|----------|---------------|------------------|
| `package` | `nginx: {installed: true, version: "1.22"}` | OS package presence/version |
| `service` | `nginx: {enabled: true, running: true}` | systemd/init service state |
| `process` | `nginx: {running: true}` | Named process is running |
| `port` | `tcp:80: {listening: true, ip: [0.0.0.0]}` | TCP/UDP port open |
| `command` | `'ls /foo': {exit-status: 0, stdout: ["bar"]}` | Run a command, assert exit/stdout/stderr |
| `file` | `/etc/hosts: {exists: true, mode: "0644", owner: root, contains: [...]}` | File presence, attributes, content |
| `user` | `nginx: {exists: true, uid: 101, gid: 101}` | OS user exists |
| `group` | `nginx: {exists: true, gid: 101}` | OS group exists |
| `http` | `http://localhost/: {status: 200, body: [...]}` | HTTP endpoint |
| `dns` | `example.com: {resolvable: true, addrs: [1.2.3.4]}` | DNS resolution |
| `addr` | `tcp://example.com:443: {reachable: true}` | TCP reachability |
| `interface` | `eth0: {exists: true, addrs: [...]}` | Network interface |
| `mount` | `/data: {exists: true, source: /dev/sda1}` | Mount point |
| `kernel-param` | `net.ipv4.ip_forward: {value: "1"}` | Sysctl value |
| `gossfile` | `./common.yaml: {}` | Include another gossfile |

## Mapping table (Goss → cosmo-smoke)

Status column: **direct** = native assertion exists; **command** = emit as shell-out fallback; **skip** = emit as TODO comment, drop the assertion but keep the test placeholder.

| Goss key | cosmo-smoke target | Status | Notes |
|----------|-------------------|--------|-------|
| `package` | `run: "dpkg -l NAME \| grep ^ii"` + `exit_code: 0` | command | Auto-detect distro? v1: just dpkg. v2: parameterize. |
| `service` | `run: "systemctl is-active NAME"` + `stdout_contains: "active"` | command | Could also check `enabled` via is-enabled. Emit both if Goss asks. |
| `process` | `process_running: NAME` | **direct** | |
| `port` (tcp) | `port_listening: {port, protocol: tcp, host}` | **direct** | `udp` same with protocol: udp. Goss uses `tcp:80` format — parse. |
| `command` | `run: "..."` with expect mapping | **direct** | Full fidelity: exit-status → exit_code, stdout → stdout_contains, stderr → stderr_contains. |
| `file` (exists) | `file_exists: /path` | **direct** | Only the `exists` attribute. Mode/owner/contains → TODO (future: file_mode, file_owner, file_contains assertions). |
| `user` | `run: "id USER"` + `exit_code: 0` | command | |
| `group` | `run: "getent group NAME"` + `exit_code: 0` | command | |
| `http` | `http: {url, status_code, body_contains}` | **direct** | Field renames: Goss `status` → `status_code`, `body` (array) → `body_contains` (single). Multi-body entries emit one test per entry. |
| `dns` | `run: "getent hosts NAME"` + `exit_code: 0` | command | `addrs` array → `stdout_contains` per addr (one assertion each, first only if simpler). |
| `addr` | `port_listening: {host, port}` | **direct** | Parse `tcp://host:port` or `udp://host:port`. |
| `interface` | `run: "ip link show NAME"` + `exit_code: 0` | command | |
| `mount` | `run: "mountpoint -q PATH"` + `exit_code: 0` | command | |
| `kernel-param` | `run: "sysctl -n KEY"` + `stdout_contains: VALUE` | command | |
| `gossfile` | `includes: [path]` | **direct** | Direct rewrite — rename key, reuse path. |

### Metadata preservation

- Goss `title:` and `meta:` fields → emit as YAML `# `-prefixed comments on the test
- Goss `vars:` → emit as `env:` at top of .smoke.yaml (needs template-time resolution, which cosmo-smoke already supports via Go templates)
- Goss `skip: true` → cosmo-smoke has no per-test skip yet. Emit as `# TODO: migrate skip flag` comment + omit the test.

## CLI design

```
smoke migrate goss <input.yaml> [flags]

Flags:
  -o, --output string    Output .smoke.yaml path (default: print to stdout)
  --overwrite            Overwrite output file if exists (default: error)
  --strict               Fail on any unmappable assertion (default: emit TODO and continue)
  --stats                Print mapping stats to stderr (directly mapped / command-fallback / skipped)
```

Defaults prefer user-friendly lossy conversion. `--strict` is for CI-driven migration audits.

## Output structure

Generated `.smoke.yaml` follows this layout:

```yaml
# Generated by: smoke migrate goss
# Source: /path/to/input.yaml
# Migrated: 2026-04-16T12:34:56Z
# Stats: 8 direct, 5 command-fallback, 2 TODO

tests:
  # ---- from Goss file: /etc/hosts ----
  - name: "file:/etc/hosts exists"
    run: "true"
    expect:
      exit_code: 0
      file_exists: "/etc/hosts"
    # TODO: Goss specified mode=0644, owner=root — add file_mode/file_owner assertions when available

  # ---- from Goss port: tcp:80 ----
  - name: "port:tcp:80 listening"
    run: "true"
    expect:
      exit_code: 0
      port_listening: {port: 80, protocol: tcp, host: "0.0.0.0"}

  # ---- from Goss command ----
  - name: "command:ls /tmp"
    run: "ls /tmp"
    expect:
      exit_code: 0
      stdout_contains: "lost+found"

  # ---- from Goss package (command fallback) ----
  - name: "package:nginx installed"
    run: "dpkg -l nginx | grep ^ii"
    expect:
      exit_code: 0
    # Migrated via command fallback — assumes Debian/Ubuntu
```

### TODO comment format

Inline comments on assertions that were partially mapped OR test-level comments above tests that were skipped entirely:

```yaml
  - name: "user:nginx"
    run: "id nginx"
    expect:
      exit_code: 0
    # TODO: Goss checked uid=101 gid=101 — not yet supported, add stdout_contains if needed
```

## Implementation plan

### Phase 1: Scaffolding (1 hour)

Files to create:
- `cmd/migrate.go` — parent `migrate` cobra command with `goss` subcommand
- `internal/migrate/goss/parser.go` — parses Goss YAML into typed structs
- `internal/migrate/goss/translator.go` — maps Goss structs → cosmo-smoke `schema.SmokeConfig`
- `internal/migrate/goss/emitter.go` — writes SmokeConfig with comment preservation (custom YAML marshaller since yaml.v3 doesn't preserve arbitrary comments easily — use string-builder approach instead of structural marshal)

### Phase 2: Typed parser (1-2 hours)

Goss YAML is a top-level map with fixed keys. Define a `GossFile` struct covering all known top-level keys as `map[string]interface{}` (since per-resource schemas vary). Start with strict struct parsing of the core subset (package, service, process, port, command, file, http, dns, addr), treat the long tail as generic maps with TODO emission.

### Phase 3: Translator (2-3 hours)

One function per Goss key: `translatePackage`, `translateService`, ... Each returns `([]schema.Test, []TranslationWarning)`. Collect all warnings for the stats report.

Test each translator in isolation with golden-file fixtures: `testdata/goss/<feature>.yaml` → `testdata/smoke/<feature>.expected.yaml`.

### Phase 4: Emitter (1 hour)

Can't use `yaml.Marshal(config)` alone because we need interleaved comments. Two options:

**Option A — Write raw YAML via templates.** Simple Go text/template that iterates tests and emits comments + YAML. Lose round-trip validity if a user manually edits it.

**Option B — Marshal + post-process.** Marshal the SmokeConfig, then walk the output and inject comments at known insertion points. Fragile.

**Recommendation:** Option A. Output is write-only from our perspective; users own the file after generation.

### Phase 5: CLI wiring + stats (1 hour)

- Hook `cmd/migrate.go` into rootCmd
- Add `--strict`, `--stats`, `--overwrite` flags
- Emit stats to stderr in `--stats` mode
- Non-zero exit on `--strict` when any skip/fallback happened

### Phase 6: Tests + fixtures (2 hours)

- Unit tests for each translator (golden files)
- Integration test: real-world goss.yaml (grab one from goss's own testdata) → verify .smoke.yaml parses back via `schema.Load`
- Error tests: malformed input, missing required fields, unknown Goss keys

**Estimated total: 8-10 hours.** Plausibly two parallel Sonnet worktrees (parser+translator in one, emitter+CLI in the other) could do it in 4-5 wall-clock hours.

## Resolved design decisions (2026-04-18)

1. **Multi-distro packages → `--distro` flag.** Ship `--distro=deb|rpm|apk` flag, default `deb`. v2 can add auto-detection later.
2. **Lossy vs strict → command: fallback + TODO is acceptable for v1.** `--strict` flag exists for CI audits; default is lossy with documented stubs.
3. **Output naming → flatten into one `.smoke.yaml`.** Gossfile includes are resolved at migration time and merged into a single output file.
4. **Reverse migration → deferred to v0.6+.** Not in scope for v0.5.
5. **Bulk migration → per-file only for v0.5.** `smoke migrate goss <file>`, no directory walking. Directory support can ship in v0.5.x or v0.6.

## v0.4 scope proposal

Ship Phase 1 + Phase 2 + Phase 3 core subset (package, service, process, port, command, file, http) + Phase 4 emitter + Phase 5 CLI. That's ~6 hours of focused work. Defer the long-tail Goss features (interface, mount, kernel-param, dns, addr) to a fast-follow release after landing what the majority of Goss files use (empirically: the 7 listed above).

## Dispatch recommendation

Not a Sonnet one-shot — too many decisions. Options:

- **Sonnet-friendly scope:** Ship ONLY the `goss → smoke` mapping table as a design doc now (this doc), then brainstorm with the user on the 5 open questions above before coding.
- **Opus does the parser + translator in this session after brainstorming**, Sonnet handles emitter + CLI + tests in a worktree.
- **Full deferral:** Sit on this doc. Ship v0.4 without Goss migration. Revisit as a v0.4.x point release.

Proposal: **Full deferral.** v0.4's theme is DX (watch + docker + retry + db). Goss migration is a strategic move that deserves its own focused session with marketing/launch coordination. Ship it as v0.5 cornerstone.

## Deferred decision

Recommended next action: commit this design doc + the 5 open questions to `docs/roadmap/items/ROAD-024.yaml` as the "ready-to-brainstorm" state, then ship v0.4 without Goss migration and tackle it fresh in v0.5.
