# Task

Add tests to cmd/ package for root command and version command. Create file cmd/root_extra_test.go in package cmd. Test cases: root command has expected subcommands (run, validate, schema, init, version, serve), version command outputs version string, --help flag produces output, unknown subcommand returns error. Verify: go test ./cmd/ -run TestRoot -v passes.
