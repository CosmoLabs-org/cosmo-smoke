# Task

Add tests to internal/reporter/prometheus package. Create file internal/reporter/prometheus_test.go if not exists, or extend existing. Test cases: all tests passing produces valid metric format with smoke_test_passed total, all tests failing produces smoke_test_failed metric, mixed pass/fail/skip, zero duration tests produce valid metrics, very long test names with special characters. Verify: go test ./internal/reporter/ -run TestPrometheus -v passes.
