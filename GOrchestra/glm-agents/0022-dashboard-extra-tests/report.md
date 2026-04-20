---
id: "0022"
task: "Add tests to internal/dashboard/ for concurrent writes and edge cases. Create file internal/dashboard/dashboard_extra_test.go in package dashboard. Test cases: concurrent writes from multiple goroutines complete without data loss, result with empty project name handled, result with special characters in test names (unicode, quotes) stored correctly, pagination returns correct subset. Use testStore and makePayload helpers from store_test.go. Verify: go test ./internal/dashboard/ -v passes."
status: merged
date: 2026-04-20
duration: 26m34s
commit: 177a2d67d9a0741f4dbce7e78639ab26fb5f8242
files_changed: 1
---

# Agent Report

## Task
Add tests to internal/dashboard/ for concurrent writes and edge cases. Create file internal/dashboard/dashboard_extra_test.go with tests for: concurrent writes from multiple goroutines, empty project name, special characters in test names, pagination subset correctness.

## What I Did
1. Read store.go to understand the Store API (InsertRun, GetProjects, GetProjectHistory) and schema
2. Read store_test.go to learn the testStore() and makePayload() helper conventions
3. Created dashboard_extra_test.go with 4 test functions
4. Hit SQLITE_BUSY errors on concurrent writes with in-memory SQLite — switched to file-based DB with WAL mode and busy_timeout pragma via DSN (`file:path?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)`)
5. Ran all 35 tests (27 existing + 8 new) — all pass
6. Ran go vet — no issues

## Decisions Made
- **File-based SQLite for concurrent test**: In-memory SQLite (`:memory:`) creates a separate DB per connection in the pool, causing "no such table" errors under concurrency. WAL mode with busy_timeout on a file-based DB resolved both issues.
- **DSN pragma format**: Used `file:path?_pragma=...` format supported by modernc.org/sqlite driver to set per-connection pragmas that persist across the connection pool.
- **Subtests for special characters**: Used t.Run for each character case (unicode, quotes, backslash, newlines) for clear failure isolation.
- **Pagination test design**: Since GetProjectHistory has no offset param, tested that limit returns the correct prefix of the full descending-order result set.

## Verification
- Build: pass (pre-existing errors from other agents' GOrchestra directories, not from dashboard package)
- Vet: pass
- Tests: 35/35 pass (8 new test cases across 4 functions)

## Files Changed
- `internal/dashboard/dashboard_extra_test.go` — 177 lines, 4 test functions with 8 total test cases

## Issues or Concerns
- The concurrent writes test relies on WAL mode + busy_timeout pragmas via DSN. The production Store (store.go) does not set these pragmas itself. In production use with real concurrent traffic, the caller would need to ensure the DSN includes these pragmas or modify NewStore to set them. This is outside the scope of this test task but worth noting.
- The "database is locked" issue under concurrency is a known SQLite limitation — WAL mode mitigates it but does not eliminate it under extreme contention.
