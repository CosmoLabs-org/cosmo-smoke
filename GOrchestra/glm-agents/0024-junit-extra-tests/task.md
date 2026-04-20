# Task

Add tests to internal/reporter/ for JUnit XML format. Create file internal/reporter/junit_extra_test.go in package reporter. Test cases: JUnit output is valid XML, testsuite has correct test count, testcase elements have name and classname, failed tests have failure element with message, skipped tests have skipped element, properties include hostname and timestamp. Verify: go test ./internal/reporter/ -run TestJUnit -v passes.
