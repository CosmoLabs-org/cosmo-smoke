# Task

Add tests to internal/baseline package. Create file internal/baseline/baseline_extra_test.go in package baseline. Test cases: concurrent file access (use t.Parallel, multiple goroutines calling Save and Load), corrupt JSON in Load (write garbage bytes to file then call Load returns error), negative duration values in Save/Load roundtrip, missing directory for Save returns error. Verify: go test ./internal/baseline/ -v passes.
