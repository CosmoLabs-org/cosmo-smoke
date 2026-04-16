package runner

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
	"github.com/tidwall/gjson"
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

// CheckProcessRunning verifies that a named process is currently running on the host.
// Uses exact process-name matching (pgrep -x on Unix, CSV-parsed tasklist on Windows).
// Bounded by a 2s timeout to prevent hangs.
func CheckProcessRunning(name string) AssertionResult {
	if name == "" {
		return AssertionResult{Type: "process_running", Expected: name, Actual: "empty name", Passed: false}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if runtime.GOOS == "windows" {
		filter := fmt.Sprintf("IMAGENAME eq %s", name)
		out, err := exec.CommandContext(ctx, "tasklist", "/FI", filter, "/FO", "CSV", "/NH").Output()
		if err != nil {
			return AssertionResult{Type: "process_running", Expected: name, Actual: "lookup error", Passed: false}
		}
		if !strings.Contains(string(out), "\""+name) {
			return AssertionResult{Type: "process_running", Expected: name, Actual: "not found", Passed: false}
		}
		return AssertionResult{Type: "process_running", Expected: name, Actual: "running", Passed: true}
	}
	out, err := exec.CommandContext(ctx, "pgrep", "-x", name).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return AssertionResult{Type: "process_running", Expected: name, Actual: "not found", Passed: false}
		}
		return AssertionResult{Type: "process_running", Expected: name, Actual: "lookup error: " + err.Error(), Passed: false}
	}
	if len(out) == 0 {
		return AssertionResult{Type: "process_running", Expected: name, Actual: "not found", Passed: false}
	}
	return AssertionResult{Type: "process_running", Expected: name, Actual: strings.TrimSpace(string(out)), Passed: true}
}

// CheckResponseTime fails if actual duration exceeds the threshold.
func CheckResponseTime(actualMs, thresholdMs int) AssertionResult {
	return AssertionResult{
		Type:     "response_time_ms",
		Expected: fmt.Sprintf("<= %dms", thresholdMs),
		Actual:   fmt.Sprintf("%dms", actualMs),
		Passed:   actualMs <= thresholdMs,
	}
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

// CheckHTTP performs an HTTP request and returns assertion results for status, body, and headers.
func CheckHTTP(check *schema.HTTPCheck) []AssertionResult {
	var results []AssertionResult

	// Default method to GET
	method := check.Method
	if method == "" {
		method = "GET"
	}

	// Set timeout (default 10s)
	timeout := 10 * time.Second
	if check.Timeout.Duration > 0 {
		timeout = check.Timeout.Duration
	}

	client := &http.Client{Timeout: timeout}

	// Build request
	var bodyReader io.Reader
	if check.Body != "" {
		bodyReader = strings.NewReader(check.Body)
	}

	req, err := http.NewRequest(method, check.URL, bodyReader)
	if err != nil {
		return []AssertionResult{{
			Type:     "http_request",
			Expected: check.URL,
			Actual:   fmt.Sprintf("invalid request: %v", err),
			Passed:   false,
		}}
	}

	// Add headers
	for k, v := range check.Headers {
		req.Header.Set(k, v)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return []AssertionResult{{
			Type:     "http_request",
			Expected: check.URL,
			Actual:   fmt.Sprintf("request failed: %v", err),
			Passed:   false,
		}}
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []AssertionResult{{
			Type:     "http_body",
			Expected: "readable body",
			Actual:   fmt.Sprintf("failed to read body: %v", err),
			Passed:   false,
		}}
	}
	bodyStr := string(body)

	// Check status code
	if check.StatusCode != nil {
		results = append(results, AssertionResult{
			Type:     "http_status",
			Expected: fmt.Sprintf("%d", *check.StatusCode),
			Actual:   fmt.Sprintf("%d", resp.StatusCode),
			Passed:   resp.StatusCode == *check.StatusCode,
		})
	}

	// Check body contains
	if check.BodyContains != "" {
		results = append(results, AssertionResult{
			Type:     "http_body_contains",
			Expected: check.BodyContains,
			Actual:   bodyStr,
			Passed:   strings.Contains(bodyStr, check.BodyContains),
		})
	}

	// Check body matches regex
	if check.BodyMatches != "" {
		matched, err := regexp.MatchString(check.BodyMatches, bodyStr)
		if err != nil {
			results = append(results, AssertionResult{
				Type:     "http_body_matches",
				Expected: check.BodyMatches,
				Actual:   fmt.Sprintf("invalid regex: %v", err),
				Passed:   false,
			})
		} else {
			results = append(results, AssertionResult{
				Type:     "http_body_matches",
				Expected: check.BodyMatches,
				Actual:   bodyStr,
				Passed:   matched,
			})
		}
	}

	// Check header contains
	for k, v := range check.HeaderContains {
		actual := resp.Header.Get(k)
		results = append(results, AssertionResult{
			Type:     "http_header_contains",
			Expected: fmt.Sprintf("%s: %s", k, v),
			Actual:   fmt.Sprintf("%s: %s", k, actual),
			Passed:   strings.Contains(actual, v),
		})
	}

	return results
}

// CheckJSONField extracts a field from JSON and validates it against equals/contains/matches.
func CheckJSONField(jsonStr string, check *schema.JSONFieldCheck) []AssertionResult {
	var results []AssertionResult

	// Check if JSON is valid
	if !gjson.Valid(jsonStr) {
		return []AssertionResult{{
			Type:     "json_field",
			Expected: check.Path,
			Actual:   "invalid JSON",
			Passed:   false,
		}}
	}

	// Extract the field value
	result := gjson.Get(jsonStr, check.Path)
	if !result.Exists() {
		return []AssertionResult{{
			Type:     "json_field",
			Expected: check.Path,
			Actual:   "field not found",
			Passed:   false,
		}}
	}

	actual := result.String()

	// Check equals
	if check.Equals != "" {
		results = append(results, AssertionResult{
			Type:     "json_field_equals",
			Expected: check.Equals,
			Actual:   actual,
			Passed:   actual == check.Equals,
		})
	}

	// Check contains
	if check.Contains != "" {
		results = append(results, AssertionResult{
			Type:     "json_field_contains",
			Expected: check.Contains,
			Actual:   actual,
			Passed:   strings.Contains(actual, check.Contains),
		})
	}

	// Check matches
	if check.Matches != "" {
		matched, err := regexp.MatchString(check.Matches, actual)
		if err != nil {
			results = append(results, AssertionResult{
				Type:     "json_field_matches",
				Expected: check.Matches,
				Actual:   fmt.Sprintf("invalid regex: %v", err),
				Passed:   false,
			})
		} else {
			results = append(results, AssertionResult{
				Type:     "json_field_matches",
				Expected: check.Matches,
				Actual:   actual,
				Passed:   matched,
			})
		}
	}

	return results
}
