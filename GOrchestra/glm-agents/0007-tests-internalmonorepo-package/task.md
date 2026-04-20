# Task

Add tests to internal/monorepo package. Create file internal/monorepo/monorepo_extra_test.go. Test cases: nested subdirectories 3+ levels deep with .smoke.yaml files, .smoke.yaml in hidden dirs should be excluded, empty directories skipped gracefully, monorepo with zero configs returns empty list, duplicate project names handled. Verify: go test ./internal/monorepo/ -v passes.
