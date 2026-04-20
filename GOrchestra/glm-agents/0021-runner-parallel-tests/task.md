# Task

Add tests to internal/runner/ for parallel execution and retry. Create file internal/runner/runner_parallel_test.go in package runner. Test cases: parallel execution runs all tests concurrently (verify via timing), fail-fast stops after first failure and skips remaining, retry with count=2 retries failed tests, retry_on_trace_only skips retry when trace assertion passes. Use noopReporter and newConfig helpers from runner_test.go. Verify: go test ./internal/runner/ -run TestParallel -v passes.
