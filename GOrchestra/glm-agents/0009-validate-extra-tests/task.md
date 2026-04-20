# Task

Add tests to cmd/ package for validate command. Create file cmd/validate_extra_test.go in package cmd. Test cases: config with all 28 assertion types valid, config with OTel enabled and valid jaeger_url, config with retry policy (count and backoff), config with skip_if conditions, config with env-specific overrides. Each test should create a temp YAML file, run the validate command, and check for success. Verify: go test ./cmd/ -run TestValidate -v passes.
