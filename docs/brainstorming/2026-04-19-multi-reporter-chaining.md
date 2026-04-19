# Multi-Reporter Chaining

**Date**: 2026-04-19
**Status**: Approved

## Goal

Allow `smoke run --format` to accept comma-separated format values so tests run once and output to multiple reporters simultaneously. Primary use cases: CI/CD pipelines (terminal + JUnit + Prometheus) and dashboard ingestion (terminal + JSON push).

## Current State

- `--format` accepts a single value: `terminal`, `json`, `junit`, `tap`, or `prometheus`
- `MultiReporter` exists in `internal/reporter/multi.go` — fans out events to multiple reporters
- All 5 reporter constructors accept `io.Writer` (not hardcoded stdout)
- OTel and Push reporters already chain via `MultiReporter` internally
- The `switch format` block is duplicated in `cmd/run.go` (monorepo path at L138-149, single-config path at L205-216)

## Design

### Reporter factory

Add `reporter.Chain()` in a new file `internal/reporter/chain.go`:

```go
func Chain(format string, stdout io.Writer) (Reporter, []io.Closer, error)
```

Input normalization:
- Trim whitespace around each format name after splitting on commas
- Ignore empty segments (from leading/trailing/repeated commas)
- Format names are case-insensitive (`JSON` == `json`)

Responsibilities:
1. Normalize and split format string on commas
2. Deduplicate format names
3. Create first reporter with `stdout`
4. Create subsequent reporters with auto-named files (see table below)
5. Wrap all in `MultiReporter`
6. Return reporter + opened files in creation order (caller closes in reverse order)
7. Return error if format string is empty after normalization

### File naming

| Format | Default file |
|--------|-------------|
| json | `smoke-results.json` |
| junit | `smoke-junit.xml` |
| prometheus | `smoke-metrics.prom` |
| tap | `smoke-tap.txt` |
| terminal | `smoke-output.txt` |

Only non-primary formats create files. Single-format usage creates no files. Files are written to CWD and overwrite without warning — this matches CI expectations for idempotent runs. A future `--output-dir` flag may be added.

### CLI usage

```bash
# Terminal on stdout + JSON to file + Prometheus to file
smoke run --format terminal,json,prometheus

# JSON on stdout + JUnit to file (CI)
smoke run --format json,junit

# Single format — unchanged behavior
smoke run --format terminal
```

### Changes to cmd/run.go

Replace both duplicated `switch format` blocks with:

```go
rep, closers, err := reporter.Chain(format, os.Stdout)
if err != nil {
    return err
}
defer func() {
    for i := len(closers) - 1; i >= 0; i-- {
        if err := closers[i].Close(); err != nil {
            fmt.Fprintf(os.Stderr, "warning: closing reporter: %v\n", err)
        }
    }
}()
rep = withOTelExport(rep, cfg)
rep = withPushReport(rep)
```

### Edge cases

- **Single format**: No files created, existing behavior preserved
- **Duplicate formats**: `--format json,json` deduplicates to single JSON reporter
- **Unknown format**: Returns error before any tests run
- **File creation failure**: Returns error immediately with path in message
- **Empty/whitespace format**: `--format ""` or `--format ",,"` returns error
- **Case insensitivity**: `--format JSON` normalized to `json`
- **Close order**: Closers returned in creation order, closed in reverse (LIFO). Close errors logged to stderr.
- **Watch mode**: File reporters are created once and overwrite on each re-run cycle. This matches terminal behavior (each run produces fresh output).
- **Monorepo mode**: File reporters receive the same aggregated events as stdout — `RunMonorepo` calls `Summary` once with combined results. One file per format, containing all sub-projects.

### Files changed

| File | Change |
|------|--------|
| `internal/reporter/chain.go` | New — `Chain()` factory + format parsing |
| `internal/reporter/chain_test.go` | New — tests for Chain() |
| `cmd/run.go` | Replace both switch blocks with Chain() call |

### What stays the same

- `--format` remains a string flag (not StringSlice) — comma-separated parsing is internal
- OTel export (`withOTelExport`) and Push reporter (`withPushReport`) continue wrapping after Chain()
- All existing reporter implementations unchanged
- `MultiReporter` unchanged
- Backward compatible — single-format `--format terminal` works identically

## Testing

Unit tests for `Chain()`:
- Single format (no files created)
- Multiple formats (files created with correct names)
- Format deduplication (`json,json` → one JSON reporter)
- Unknown format returns error
- Empty and whitespace-only format returns error
- Case insensitivity (`JSON`, `Json`, `json` all work)
- Trailing/leading comma handling
- File creation failure (mock permission error)
- Close order verification (LIFO)
- Close error logged to stderr

Integration tests:
- `smoke run --format terminal,json` against cosmo-smoke's own `.smoke.yaml`
- Monorepo mode with multi-format produces valid combined output
- Watch mode with multi-format overwrites files per cycle
- OTel/Push wrappers work correctly after Chain()
