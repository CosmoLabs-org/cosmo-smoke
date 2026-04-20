# Task

Add tests to internal/dashboard/ for concurrent writes and edge cases. Create file internal/dashboard/dashboard_extra_test.go in package dashboard. Test cases: concurrent writes from multiple goroutines complete without data loss, result with empty project name handled, result with special characters in test names (unicode, quotes) stored correctly, pagination returns correct subset. Use testStore and makePayload helpers from store_test.go. Verify: go test ./internal/dashboard/ -v passes.
