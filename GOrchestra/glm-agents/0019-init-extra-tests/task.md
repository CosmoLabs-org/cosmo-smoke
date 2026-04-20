# Task

Add tests to cmd/ package for the init command. Create file cmd/init_extra_test.go in package cmd. Test cases: init in empty directory creates .smoke.yaml, init with --force overwrites existing config, init detects project type from go.mod, init detects project type from package.json. Verify: go test ./cmd/ -run TestInit -v passes.
