# Task

Add tests to cmd/ package for schema command. Create file cmd/schema_extra_test.go in package cmd. Test cases: JSON output is valid JSON and roundtrips, schema contains all expected assertion type names (exit_code, stdout_contains, http, json_field, ssl_cert, redis_ping, websocket, docker_container_running, url_reachable, s3_bucket, version_check, otel_trace, credential_check, graphql), each assertion type has at least one field, all field types are non-empty strings. Verify: go test ./cmd/ -run TestSchema -v passes.
