package runner

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// AssertionResult holds the outcome of a single assertion check.
type AssertionResult struct {
	Type     string // "exit_code", "stdout_contains", "stdout_matches", "stderr_contains", "stderr_matches", "file_exists", "env_exists"
	Expected string
	Actual   string
	Passed   bool
}

// CheckExitCode verifies that the process exit code matches the expected value.
func CheckExitCode(actual int, expected int) AssertionResult {
	return AssertionResult{
		Type:     "exit_code",
		Expected: fmt.Sprintf("%d", expected),
		Actual:   fmt.Sprintf("%d", actual),
		Passed:   actual == expected,
	}
}

// CheckStdoutContains verifies that stdout contains the given substring.
func CheckStdoutContains(stdout, substr string) AssertionResult {
	return AssertionResult{
		Type:     "stdout_contains",
		Expected: substr,
		Actual:   stdout,
		Passed:   strings.Contains(stdout, substr),
	}
}

// CheckStdoutMatches verifies that stdout matches the given regex pattern.
// If the pattern is invalid, the assertion fails with an explanatory Actual value.
func CheckStdoutMatches(stdout, pattern string) AssertionResult {
	matched, err := regexp.MatchString(pattern, stdout)
	if err != nil {
		return AssertionResult{
			Type:     "stdout_matches",
			Expected: pattern,
			Actual:   fmt.Sprintf("invalid regex: %v", err),
			Passed:   false,
		}
	}
	return AssertionResult{
		Type:     "stdout_matches",
		Expected: pattern,
		Actual:   stdout,
		Passed:   matched,
	}
}

// CheckStderrContains verifies that stderr contains the given substring.
func CheckStderrContains(stderr, substr string) AssertionResult {
	return AssertionResult{
		Type:     "stderr_contains",
		Expected: substr,
		Actual:   stderr,
		Passed:   strings.Contains(stderr, substr),
	}
}

// CheckStderrMatches verifies that stderr matches the given regex pattern.
// If the pattern is invalid, the assertion fails with an explanatory Actual value.
func CheckStderrMatches(stderr, pattern string) AssertionResult {
	matched, err := regexp.MatchString(pattern, stderr)
	if err != nil {
		return AssertionResult{
			Type:     "stderr_matches",
			Expected: pattern,
			Actual:   fmt.Sprintf("invalid regex: %v", err),
			Passed:   false,
		}
	}
	return AssertionResult{
		Type:     "stderr_matches",
		Expected: pattern,
		Actual:   stderr,
		Passed:   matched,
	}
}

// CheckEnvExists verifies that an environment variable is set (non-empty).
func CheckEnvExists(name string) AssertionResult {
	value := os.Getenv(name)
	return AssertionResult{
		Type:     "env_exists",
		Expected: name,
		Actual:   value,
		Passed:   value != "",
	}
}

// CheckPortListening verifies that a port is open and accepting connections.
func CheckPortListening(port int, protocol, host string) AssertionResult {
	if protocol == "" {
		protocol = "tcp"
	}
	if host == "" {
		host = "localhost"
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout(protocol, addr, 5*time.Second)
	if err != nil {
		return AssertionResult{Type: "port_listening", Expected: addr, Actual: err.Error(), Passed: false}
	}
	conn.Close()
	return AssertionResult{Type: "port_listening", Expected: addr, Actual: "open", Passed: true}
}

// CheckFileExists verifies that a file exists at the given path.
// Relative paths are resolved against configDir using filepath.Join.
func CheckFileExists(path, configDir string) AssertionResult {
	resolved := path
	if !filepath.IsAbs(path) {
		resolved = filepath.Join(configDir, path)
	}

	_, err := os.Stat(resolved)
	passed := err == nil

	return AssertionResult{
		Type:     "file_exists",
		Expected: resolved,
		Actual:   resolved,
		Passed:   passed,
	}
}
