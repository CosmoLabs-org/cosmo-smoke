---
id: "0008"
task: "Add tests to internal/baseline package. Create file internal/baseline/baseline_extra_test.go in package baseline. Test cases: concurrent file access (use t.Parallel, multiple goroutines calling Save and Load), corrupt JSON in Load (write garbage bytes to file then call Load returns error), negative duration values in Save/Load roundtrip, missing directory for Save returns error. Verify: go test ./internal/baseline/ -v passes."
status: merged
date: 2026-04-20
duration: 9m34s
commit: 9bc67f742d369d7229849bc079a10cc6d3b26c10
files_changed: 1
---

# Agent Report

## Task
Add tests to internal/baseline package: concurrent file access, corrupt JSON in Load, negative duration roundtrip, missing directory error.

## What I Did
1. Read `internal/baseline/baseline.go` to understand Load/Save API and existing types (File, Entry).
2. Read existing `internal/baseline/baseline_test.go` to follow established patterns (t.TempDir, filepath.Join, time.Date).
3. Created `internal/baseline/baseline_extra_test.go` with 4 test functions:
   - **TestConcurrentSaveAndLoad**: Uses `t.Parallel()`, spawns 5 writer + 5 reader goroutines via sync.WaitGroup, verifies final state is valid.
   - **TestLoad_CorruptJSON**: Writes garbage bytes to file, verifies Load returns an error.
   - **TestNegativeDuration_Roundtrip**: Saves File with DurationMs=-100, loads it back, verifies negative value preserved.
   - **TestSave_MissingDirectory**: Saves to deeply nested nonexistent path, verifies error returned.
4. Ran `go test ./internal/baseline/` — all 11 tests pass (7 existing + 4 new).
5. Committed with conventional commit message.

## Decisions Made
- Followed existing test patterns: `t.TempDir()` for isolation, `time.Date()` for deterministic timestamps.
- Used `sync.WaitGroup` for concurrency test rather than channels — simpler and matches Go testing conventions.
- Copied File map in each writer goroutine to avoid data races on the map.

## Verification
- Build: pass (`go build ./internal/baseline/`)
- Tests: pass (11 tests in 1 package)
- Vet/Lint: no errors (2 cosmetic rangeint suggestions from go1.26, not blocking)

## Files Changed
- `internal/baseline/baseline_extra_test.go` — new file with 4 test cases (107 lines)

## Issues or Concerns
- None. All tests are deterministic and use temp directories for isolation.
