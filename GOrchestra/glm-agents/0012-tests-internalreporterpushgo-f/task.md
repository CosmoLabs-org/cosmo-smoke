# Task

Add tests to internal/reporter/push.go. Create file internal/reporter/push_test.go. Test cases: successful POST to httptest.Server endpoint returns nil, endpoint returns non-200 returns error, timeout handling with short context, empty API key handled gracefully, empty URL returns error, malformed URL returns error. Verify: go test ./internal/reporter/ -run TestPush -v passes.
