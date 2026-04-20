# Task

Add tests to internal/runner/ for extended functionality. Create file internal/runner/runner_extended_test.go. Test cases: filterTests with multiple include tags, filterTests with exclude only, runTest with allow_failure flag, shouldSkip with FileMissing for absolute vs relative paths. Verify: go test ./internal/runner/ -run TestExtended -v passes.
