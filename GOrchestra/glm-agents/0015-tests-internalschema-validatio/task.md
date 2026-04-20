# Task

Add tests to internal/schema/ for validation edge cases. Create file internal/schema/validation_extra_test.go. Test cases: Validate with websocket valid config (url, send, expect_contains), Validate with graphql valid config (url, query), Validate with credential_check all three sources (env, file, exec), Validate with s3_bucket with custom endpoint. Verify: go test ./internal/schema/ -run TestValidationExtra -v passes.
