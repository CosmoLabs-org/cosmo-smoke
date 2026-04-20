# Task

Add tests to internal/migrate/goss/ for Goss migration. Create file internal/migrate/goss/goss_extra_test.go in package goss. Test cases: parse valid Gossfile with all assertion types, parse Gossfile with empty vars section, parse Gossfile with HTTP tests, parse Gossfile with process tests, convert Goss HTTP test to smoke HTTP assertion, convert Goss process test to process_running assertion. Verify: go test ./internal/migrate/goss/ -v passes.
