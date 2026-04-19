package runner

import (
	"bytes"
	"context"
	"crypto/tls"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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

// CheckSSLCert dials host:port over TLS and validates the certificate chain,
// expiry, and days-remaining threshold.
func CheckSSLCert(check *schema.SSLCertCheck) AssertionResult {
	port := check.Port
	if port == 0 {
		port = 443
	}
	addr := fmt.Sprintf("%s:%d", check.Host, port)
	conf := &tls.Config{
		ServerName:         check.Host,
		InsecureSkipVerify: check.AllowSelfSigned, //nolint:gosec -- opt-in per AllowSelfSigned flag
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, conf)
	if err != nil {
		return AssertionResult{Type: "ssl_cert", Expected: addr, Actual: err.Error(), Passed: false}
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return AssertionResult{Type: "ssl_cert", Expected: addr, Actual: "no peer certificate", Passed: false}
	}
	leaf := certs[0]
	now := time.Now()
	if now.After(leaf.NotAfter) {
		return AssertionResult{
			Type:     "ssl_cert",
			Expected: "valid cert",
			Actual:   "expired on " + leaf.NotAfter.Format("2006-01-02"),
			Passed:   false,
		}
	}
	if check.MinDaysRemaining > 0 {
		daysLeft := int(leaf.NotAfter.Sub(now).Hours() / 24)
		if daysLeft < check.MinDaysRemaining {
			return AssertionResult{
				Type:     "ssl_cert",
				Expected: fmt.Sprintf(">= %d days", check.MinDaysRemaining),
				Actual:   fmt.Sprintf("%d days", daysLeft),
				Passed:   false,
			}
		}
	}
	return AssertionResult{
		Type:     "ssl_cert",
		Expected: addr,
		Actual:   fmt.Sprintf("valid, expires %s", leaf.NotAfter.Format("2006-01-02")),
		Passed:   true,
	}
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

// CheckRedisPing issues a PING to a Redis server and expects +PONG.
func CheckRedisPing(check *schema.RedisCheck) AssertionResult {
	host := check.Host
	if host == "" {
		host = "localhost"
	}
	port := check.Port
	if port == 0 {
		port = 6379
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return AssertionResult{Type: "redis_ping", Expected: addr, Actual: err.Error(), Passed: false}
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second)) //nolint:errcheck

	// Optional AUTH
	if check.Password != "" {
		authCmd := fmt.Sprintf("*2\r\n$4\r\nAUTH\r\n$%d\r\n%s\r\n", len(check.Password), check.Password)
		if _, err := conn.Write([]byte(authCmd)); err != nil {
			return AssertionResult{Type: "redis_ping", Expected: addr, Actual: "auth write error: " + err.Error(), Passed: false}
		}
		buf := make([]byte, 128)
		n, _ := conn.Read(buf)
		if !strings.HasPrefix(string(buf[:n]), "+OK") {
			return AssertionResult{Type: "redis_ping", Expected: "+OK", Actual: strings.TrimSpace(string(buf[:n])), Passed: false}
		}
	}

	if _, err := conn.Write([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		return AssertionResult{Type: "redis_ping", Expected: addr, Actual: err.Error(), Passed: false}
	}
	buf := make([]byte, 64)
	n, err := conn.Read(buf)
	if err != nil {
		return AssertionResult{Type: "redis_ping", Expected: "+PONG", Actual: err.Error(), Passed: false}
	}
	reply := strings.TrimSpace(string(buf[:n]))
	if !strings.HasPrefix(reply, "+PONG") {
		return AssertionResult{Type: "redis_ping", Expected: "+PONG", Actual: reply, Passed: false}
	}
	return AssertionResult{Type: "redis_ping", Expected: addr, Actual: "PONG", Passed: true}
}

// CheckMemcachedVersion issues `version` to Memcached and expects a VERSION line.
func CheckMemcachedVersion(check *schema.MemcachedCheck) AssertionResult {
	host := check.Host
	if host == "" {
		host = "localhost"
	}
	port := check.Port
	if port == 0 {
		port = 11211
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return AssertionResult{Type: "memcached_version", Expected: addr, Actual: err.Error(), Passed: false}
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second)) //nolint:errcheck

	if _, err := conn.Write([]byte("version\r\n")); err != nil {
		return AssertionResult{Type: "memcached_version", Expected: addr, Actual: err.Error(), Passed: false}
	}
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		return AssertionResult{Type: "memcached_version", Expected: "VERSION", Actual: err.Error(), Passed: false}
	}
	reply := strings.TrimSpace(string(buf[:n]))
	if !strings.HasPrefix(reply, "VERSION") {
		return AssertionResult{Type: "memcached_version", Expected: "VERSION ...", Actual: reply, Passed: false}
	}
	return AssertionResult{Type: "memcached_version", Expected: addr, Actual: reply, Passed: true}
}

// CheckPostgresPing sends an SSLRequest to a Postgres server and verifies a protocol-valid response.
func CheckPostgresPing(check *schema.PostgresCheck) AssertionResult {
	host := check.Host
	if host == "" {
		host = "localhost"
	}
	port := check.Port
	if port == 0 {
		port = 5432
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return AssertionResult{Type: "postgres_ping", Expected: addr, Actual: err.Error(), Passed: false}
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second)) //nolint:errcheck

	// SSLRequest message: int32 length (8), int32 code (80877103)
	// Bytes: 00 00 00 08 04 D2 16 2F
	sslReq := []byte{0x00, 0x00, 0x00, 0x08, 0x04, 0xD2, 0x16, 0x2F}
	if _, err := conn.Write(sslReq); err != nil {
		return AssertionResult{Type: "postgres_ping", Expected: addr, Actual: "write error: " + err.Error(), Passed: false}
	}
	buf := make([]byte, 1)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return AssertionResult{Type: "postgres_ping", Expected: "S or N", Actual: fmt.Sprintf("read error: %v", err), Passed: false}
	}
	reply := buf[0]
	// 'S' = SSL supported, 'N' = SSL not supported, 'E' = error message follows (still valid postgres)
	if reply == 'S' || reply == 'N' || reply == 'E' {
		return AssertionResult{Type: "postgres_ping", Expected: addr, Actual: string(reply), Passed: true}
	}
	return AssertionResult{Type: "postgres_ping", Expected: "S/N/E", Actual: fmt.Sprintf("0x%02x", reply), Passed: false}
}

// CheckMySQLPing verifies a MySQL server sends a valid v10 handshake packet on connection.
func CheckMySQLPing(check *schema.MySQLCheck) AssertionResult {
	host := check.Host
	if host == "" {
		host = "localhost"
	}
	port := check.Port
	if port == 0 {
		port = 3306
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return AssertionResult{Type: "mysql_ping", Expected: addr, Actual: err.Error(), Passed: false}
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second)) //nolint:errcheck

	// MySQL server immediately sends a handshake packet:
	// [3 bytes: payload length][1 byte: sequence id][payload starts with 1 byte: protocol version]
	// Protocol version 10 (0x0a) is "v10", the current universally-used version.
	hdr := make([]byte, 5)
	n, err := conn.Read(hdr)
	if err != nil || n < 5 {
		return AssertionResult{Type: "mysql_ping", Expected: "handshake", Actual: fmt.Sprintf("read error: %v (n=%d)", err, n), Passed: false}
	}
	protocolVersion := hdr[4]
	if protocolVersion != 0x0a {
		return AssertionResult{Type: "mysql_ping", Expected: "protocol v10 (0x0a)", Actual: fmt.Sprintf("0x%02x", protocolVersion), Passed: false}
	}
	return AssertionResult{Type: "mysql_ping", Expected: addr, Actual: "v10", Passed: true}
}

// CheckGRPCHealth queries grpc.health.v1.Health/Check and passes if SERVING.
func CheckGRPCHealth(check *schema.GRPCHealthCheck) AssertionResult {
	timeout := check.Timeout.Duration
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var creds credentials.TransportCredentials
	if check.UseTLS {
		creds = credentials.NewTLS(nil)
	} else {
		creds = insecure.NewCredentials()
	}

	conn, err := grpc.NewClient(check.Address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return AssertionResult{
			Type:     "grpc_health",
			Expected: check.Address,
			Actual:   "dial error: " + err.Error(),
			Passed:   false,
		}
	}
	defer conn.Close()

	client := healthpb.NewHealthClient(conn)
	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: check.Service})
	if err != nil {
		return AssertionResult{
			Type:     "grpc_health",
			Expected: "SERVING",
			Actual:   "rpc error: " + err.Error(),
			Passed:   false,
		}
	}

	status := resp.GetStatus().String()
	return AssertionResult{
		Type:     "grpc_health",
		Expected: "SERVING",
		Actual:   status,
		Passed:   status == "SERVING",
	}
}

// isDockerAvailable returns true if the docker daemon is reachable.
func isDockerAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "docker", "info").Run() == nil
}

// CheckDockerContainerRunning checks if a named Docker container is running.
func CheckDockerContainerRunning(check *schema.DockerContainerCheck) AssertionResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "inspect", "--format={{.State.Running}}", check.Name).Output()
	if err != nil {
		return AssertionResult{Type: "docker_container_running", Expected: check.Name, Actual: "container not found or docker unavailable: " + err.Error(), Passed: false}
	}
	running := strings.TrimSpace(string(out))
	if running != "true" {
		return AssertionResult{Type: "docker_container_running", Expected: "true", Actual: running, Passed: false}
	}
	return AssertionResult{Type: "docker_container_running", Expected: check.Name, Actual: "running", Passed: true}
}

// CheckDockerImageExists checks if a Docker image exists locally.
func CheckDockerImageExists(check *schema.DockerImageCheck) AssertionResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := exec.CommandContext(ctx, "docker", "image", "inspect", check.Image).Run()
	if err != nil {
		return AssertionResult{Type: "docker_image_exists", Expected: check.Image, Actual: "image not found or docker unavailable: " + err.Error(), Passed: false}
	}
	return AssertionResult{Type: "docker_image_exists", Expected: check.Image, Actual: "exists", Passed: true}
}

// httpReachable makes an HTTP GET request and returns the status code,
// response time, and any error. expectedStatus=0 means any 2xx passes.
func httpReachable(url string, timeout time.Duration, expectedStatus int) (statusCode int, elapsed time.Duration, err error) {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	start := time.Now()
	resp, err := client.Get(url)
	elapsed = time.Since(start)
	if err != nil {
		return 0, elapsed, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, elapsed, nil
}

// is2xx returns true if the status code is in the 200-299 range.
func is2xx(code int) bool {
	return code >= 200 && code <= 299
}

// CheckURLReachable verifies an HTTP/HTTPS endpoint is accessible.
func CheckURLReachable(check *schema.URLReachableCheck) AssertionResult {
	statusCode, elapsed, err := httpReachable(check.URL, check.Timeout.Duration, 0)
	if err != nil {
		return AssertionResult{
			Type:     "url_reachable",
			Expected: fmt.Sprintf("%s reachable", check.URL),
			Actual:   fmt.Sprintf("connection failed: %v", err),
			Passed:   false,
		}
	}
	expected := "any 2xx"
	passed := is2xx(statusCode)
	if check.StatusCode != nil {
		expected = fmt.Sprintf("HTTP %d", *check.StatusCode)
		passed = statusCode == *check.StatusCode
	}
	return AssertionResult{
		Type:     "url_reachable",
		Expected: expected,
		Actual:   fmt.Sprintf("HTTP %d (%s)", statusCode, elapsed.Round(time.Millisecond)),
		Passed:   passed,
	}
}

// CheckServiceReachable verifies an external service dependency is accessible.
func CheckServiceReachable(check *schema.ServiceReachableCheck) AssertionResult {
	statusCode, elapsed, err := httpReachable(check.URL, check.Timeout.Duration, 0)
	if err != nil {
		return AssertionResult{
			Type:     "service_reachable",
			Expected: fmt.Sprintf("%s reachable", check.URL),
			Actual:   fmt.Sprintf("connection failed: %v", err),
			Passed:   false,
		}
	}
	return AssertionResult{
		Type:     "service_reachable",
		Expected: "any 2xx",
		Actual:   fmt.Sprintf("HTTP %d (%s)", statusCode, elapsed.Round(time.Millisecond)),
		Passed:   is2xx(statusCode),
	}
}

// CheckS3Bucket verifies an S3-compatible bucket is accessible via anonymous HEAD.
func CheckS3Bucket(check *schema.S3BucketCheck) AssertionResult {
	endpoint := check.Endpoint
	if endpoint == "" {
		endpoint = "s3.amazonaws.com"
	}
	var url string
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		url = fmt.Sprintf("%s/%s?location", endpoint, check.Bucket)
	} else {
		url = fmt.Sprintf("https://%s/%s?location", endpoint, check.Bucket)
	}

	statusCode, elapsed, err := httpReachable(url, 5*time.Second, 0)
	if err != nil {
		return AssertionResult{
			Type:     "s3_bucket",
			Expected: fmt.Sprintf("bucket %s accessible", check.Bucket),
			Actual:   fmt.Sprintf("connection failed: %v", err),
			Passed:   false,
		}
	}
	if statusCode == 403 {
		return AssertionResult{
			Type:     "s3_bucket",
			Expected: fmt.Sprintf("bucket %s accessible", check.Bucket),
			Actual:   "HTTP 403 Forbidden — bucket requires authentication; use http assertion with Go templates for authenticated access",
			Passed:   false,
		}
	}
	return AssertionResult{
		Type:     "s3_bucket",
		Expected: fmt.Sprintf("bucket %s accessible", check.Bucket),
		Actual:   fmt.Sprintf("HTTP %d (%s)", statusCode, elapsed.Round(time.Millisecond)),
		Passed:   is2xx(statusCode),
	}
}

// CheckVersion runs a shell command and regex-matches stdout.
func CheckVersion(check *schema.VersionCheck) AssertionResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", check.Command)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return AssertionResult{
			Type:     "version_check",
			Expected: fmt.Sprintf("pattern %q", check.Pattern),
			Actual:   fmt.Sprintf("command failed: %v", err),
			Passed:   false,
		}
	}

	re := regexp.MustCompile(check.Pattern)
	output := strings.TrimSpace(stdout.String())
	if re.MatchString(output) {
		return AssertionResult{
			Type:     "version_check",
			Expected: fmt.Sprintf("pattern %q", check.Pattern),
			Actual:   output,
			Passed:   true,
		}
	}
	return AssertionResult{
		Type:     "version_check",
		Expected: fmt.Sprintf("pattern %q", check.Pattern),
		Actual:   fmt.Sprintf("output %q did not match", output),
		Passed:   false,
	}
}
