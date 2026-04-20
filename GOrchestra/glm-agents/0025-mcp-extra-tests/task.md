# Task

Add tests to internal/mcp/ for remaining helpers. Create file internal/mcp/mcp_extra_test.go in package mcp. Test cases: generateExpectBlock with all assertion types, generateExpectBlock with empty assertions, boolArg with various string values, parseAssertion from YAML map, template rendering with environment variables. Verify: go test ./internal/mcp/ -v passes.
