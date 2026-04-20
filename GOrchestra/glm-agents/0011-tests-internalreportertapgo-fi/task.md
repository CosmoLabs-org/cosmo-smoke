# Task

Add tests to internal/reporter/tap.go. Create file internal/reporter/tap_test.go. Test cases: TAP plan line format (1..N), passing tests output 'ok N - name', failing tests output 'not ok N - name', skipped tests with SKIP directive, multiple tests ordering preserved. Verify: go test ./internal/reporter/ -run TestTAP -v passes.
