package mcp

import (
	"strings"
	"testing"
)

// --- generateExpectBlock detailed output tests ---

func TestGenerateExpectBlock_ExitCodeDetailed(t *testing.T) {
	got := generateExpectBlock("exit_code", map[string]interface{}{"code": float64(42)})
	if !strings.Contains(got, "exit_code: 42") {
		t.Errorf("expected exit_code: 42 in %q", got)
	}
	// Default
	got = generateExpectBlock("exit_code", nil)
	if !strings.Contains(got, "exit_code: 0") {
		t.Errorf("expected exit_code: 0 default in %q", got)
	}
}

func TestGenerateExpectBlock_StdoutContainsDetailed(t *testing.T) {
	got := generateExpectBlock("stdout_contains", map[string]interface{}{"text": "hello world"})
	if !strings.Contains(got, `stdout_contains: "hello world"`) {
		t.Errorf("expected quoted string in %q", got)
	}
	got = generateExpectBlock("stdout_contains", nil)
	if !strings.Contains(got, "stdout_contains:") {
		t.Errorf("expected stdout_contains in default %q", got)
	}
}

func TestGenerateExpectBlock_StdoutMatchesDetailed(t *testing.T) {
	got := generateExpectBlock("stdout_matches", map[string]interface{}{"pattern": `^v\d+`})
	if !strings.Contains(got, "stdout_matches:") {
		t.Errorf("expected stdout_matches in %q", got)
	}
}

func TestGenerateExpectBlock_StderrDetailed(t *testing.T) {
	got := generateExpectBlock("stderr_contains", map[string]interface{}{"text": "fatal"})
	if !strings.Contains(got, `stderr_contains: "fatal"`) {
		t.Errorf("expected stderr_contains in %q", got)
	}
	got = generateExpectBlock("stderr_matches", map[string]interface{}{"pattern": "err.*"})
	if !strings.Contains(got, "stderr_matches:") {
		t.Errorf("expected stderr_matches in %q", got)
	}
}

func TestGenerateExpectBlock_FileAndEnv(t *testing.T) {
	got := generateExpectBlock("file_exists", map[string]interface{}{"path": "/tmp/out.txt"})
	if !strings.Contains(got, `file_exists: "/tmp/out.txt"`) {
		t.Errorf("expected file_exists in %q", got)
	}
	got = generateExpectBlock("env_exists", map[string]interface{}{"var": "MY_VAR"})
	if !strings.Contains(got, `env_exists: "MY_VAR"`) {
		t.Errorf("expected env_exists in %q", got)
	}
}

func TestGenerateExpectBlock_PortListening(t *testing.T) {
	got := generateExpectBlock("port_listening", map[string]interface{}{
		"port":     float64(3000),
		"protocol": "udp",
		"host":     "0.0.0.0",
	})
	if !strings.Contains(got, "port: 3000") {
		t.Errorf("expected port 3000 in %q", got)
	}
	if !strings.Contains(got, "protocol: udp") {
		t.Errorf("expected protocol udp in %q", got)
	}
	if !strings.Contains(got, "host: 0.0.0.0") {
		t.Errorf("expected host in %q", got)
	}
}

func TestGenerateExpectBlock_HTTPDetailed(t *testing.T) {
	got := generateExpectBlock("http", map[string]interface{}{
		"url":         "http://localhost:9090/health",
		"status_code": float64(201),
	})
	if !strings.Contains(got, `url: "http://localhost:9090/health"`) {
		t.Errorf("expected url in %q", got)
	}
	if !strings.Contains(got, "status_code: 201") {
		t.Errorf("expected status_code 201 in %q", got)
	}
}

func TestGenerateExpectBlock_JSONField(t *testing.T) {
	got := generateExpectBlock("json_field", map[string]interface{}{
		"path":   "data.name",
		"equals": "Alice",
	})
	if !strings.Contains(got, `path: "data.name"`) {
		t.Errorf("expected path in %q", got)
	}
	if !strings.Contains(got, `equals: "Alice"`) {
		t.Errorf("expected equals in %q", got)
	}
	// Without equals
	got = generateExpectBlock("json_field", map[string]interface{}{"path": "status"})
	if strings.Contains(got, "equals:") {
		t.Errorf("should not have equals when empty, got %q", got)
	}
	if !strings.Contains(got, `path: "status"`) {
		t.Errorf("expected path in %q", got)
	}
}

func TestGenerateExpectBlock_ResponseTime(t *testing.T) {
	got := generateExpectBlock("response_time_ms", map[string]interface{}{"ms": float64(200)})
	if !strings.Contains(got, "response_time_ms: 200") {
		t.Errorf("expected response_time_ms 200 in %q", got)
	}
}

func TestGenerateExpectBlock_SSLCert(t *testing.T) {
	got := generateExpectBlock("ssl_cert", map[string]interface{}{
		"host":               "mysite.com",
		"min_days_remaining": float64(60),
	})
	if !strings.Contains(got, `host: "mysite.com"`) {
		t.Errorf("expected host in %q", got)
	}
	if !strings.Contains(got, "min_days_remaining: 60") {
		t.Errorf("expected min_days_remaining in %q", got)
	}
}

func TestGenerateExpectBlock_RedisPing(t *testing.T) {
	got := generateExpectBlock("redis_ping", map[string]interface{}{
		"host": "redis-server",
		"port": float64(6380),
	})
	if !strings.Contains(got, `host: "redis-server"`) {
		t.Errorf("expected host in %q", got)
	}
	if !strings.Contains(got, "port: 6380") {
		t.Errorf("expected port 6380 in %q", got)
	}
}

func TestGenerateExpectBlock_Memcached(t *testing.T) {
	got := generateExpectBlock("memcached_version", map[string]interface{}{
		"host": "cache.local",
		"port": float64(11212),
	})
	if !strings.Contains(got, `host: "cache.local"`) {
		t.Errorf("expected host in %q", got)
	}
}

func TestGenerateExpectBlock_Postgres(t *testing.T) {
	got := generateExpectBlock("postgres_ping", map[string]interface{}{
		"host": "db.internal",
		"port": float64(5433),
	})
	if !strings.Contains(got, "postgres_ping:") {
		t.Errorf("expected postgres_ping in %q", got)
	}
	if !strings.Contains(got, `host: "db.internal"`) {
		t.Errorf("expected host in %q", got)
	}
}

func TestGenerateExpectBlock_MySQL(t *testing.T) {
	got := generateExpectBlock("mysql_ping", map[string]interface{}{
		"host": "mysql.internal",
		"port": float64(3307),
	})
	if !strings.Contains(got, "mysql_ping:") {
		t.Errorf("expected mysql_ping in %q", got)
	}
}

func TestGenerateExpectBlock_GRPCHealth(t *testing.T) {
	got := generateExpectBlock("grpc_health", map[string]interface{}{
		"address": "api.internal:443",
	})
	if !strings.Contains(got, `address: "api.internal:443"`) {
		t.Errorf("expected address in %q", got)
	}
}

func TestGenerateExpectBlock_Docker(t *testing.T) {
	got := generateExpectBlock("docker_container_running", map[string]interface{}{
		"name": "my-postgres",
	})
	if !strings.Contains(got, `name: "my-postgres"`) {
		t.Errorf("expected container name in %q", got)
	}
	got = generateExpectBlock("docker_image_exists", map[string]interface{}{
		"image": "postgres:16",
	})
	if !strings.Contains(got, `image: "postgres:16"`) {
		t.Errorf("expected image in %q", got)
	}
}

func TestGenerateExpectBlock_URLReachable(t *testing.T) {
	got := generateExpectBlock("url_reachable", map[string]interface{}{
		"url":         "https://api.example.com/healthz",
		"status_code": float64(204),
	})
	if !strings.Contains(got, `url: "https://api.example.com/healthz"`) {
		t.Errorf("expected url in %q", got)
	}
	if !strings.Contains(got, "status_code: 204") {
		t.Errorf("expected status_code 204 in %q", got)
	}
}

func TestGenerateExpectBlock_ServiceReachable(t *testing.T) {
	got := generateExpectBlock("service_reachable", map[string]interface{}{
		"url": "https://payments.internal/v1/health",
	})
	if !strings.Contains(got, `url: "https://payments.internal/v1/health"`) {
		t.Errorf("expected url in %q", got)
	}
}

func TestGenerateExpectBlock_S3Bucket(t *testing.T) {
	got := generateExpectBlock("s3_bucket", map[string]interface{}{
		"bucket": "assets-prod",
		"region": "eu-west-1",
	})
	if !strings.Contains(got, `bucket: "assets-prod"`) {
		t.Errorf("expected bucket in %q", got)
	}
	if !strings.Contains(got, `region: "eu-west-1"`) {
		t.Errorf("expected region in %q", got)
	}
}

func TestGenerateExpectBlock_VersionCheck(t *testing.T) {
	got := generateExpectBlock("version_check", map[string]interface{}{
		"command": "node --version",
		"pattern": `v\d+\.\d+`,
	})
	if !strings.Contains(got, `command: "node --version"`) {
		t.Errorf("expected command in %q", got)
	}
	if !strings.Contains(got, "pattern:") {
		t.Errorf("expected pattern in %q", got)
	}
}

func TestGenerateExpectBlock_WebSocket(t *testing.T) {
	got := generateExpectBlock("websocket", map[string]interface{}{
		"url":  "wss://ws.example.com/stream",
		"send": `{"type":"ping"}`,
	})
	if !strings.Contains(got, `url: "wss://ws.example.com/stream"`) {
		t.Errorf("expected url in %q", got)
	}
	if !strings.Contains(got, `send: "{\"type\":\"ping\"}"`) {
		t.Errorf("expected send in %q", got)
	}
	if !strings.Contains(got, "expect_contains:") {
		t.Errorf("expected expect_contains in %q", got)
	}
}

func TestGenerateExpectBlock_OTelTrace(t *testing.T) {
	got := generateExpectBlock("otel_trace", map[string]interface{}{
		"service_name": "payment-svc",
		"min_spans":    float64(3),
	})
	if !strings.Contains(got, `service_name: "payment-svc"`) {
		t.Errorf("expected service_name in %q", got)
	}
	if !strings.Contains(got, "min_spans: 3") {
		t.Errorf("expected min_spans 3 in %q", got)
	}
}

func TestGenerateExpectBlock_CredentialCheck(t *testing.T) {
	got := generateExpectBlock("credential_check", map[string]interface{}{
		"source": "file",
		"name":   "/run/secrets/db-password",
	})
	if !strings.Contains(got, `source: "file"`) {
		t.Errorf("expected source in %q", got)
	}
	if !strings.Contains(got, `name: "/run/secrets/db-password"`) {
		t.Errorf("expected name in %q", got)
	}
}

func TestGenerateExpectBlock_GraphQL(t *testing.T) {
	got := generateExpectBlock("graphql", map[string]interface{}{
		"url": "http://api.example.com/graphql",
	})
	if !strings.Contains(got, `url: "http://api.example.com/graphql"`) {
		t.Errorf("expected url in %q", got)
	}
}

// --- generateExpectBlock with empty/nil params ---

func TestGenerateExpectBlock_EmptyParams(t *testing.T) {
	types := []string{
		"exit_code", "stdout_contains", "stdout_matches",
		"stderr_contains", "stderr_matches", "file_exists", "env_exists",
		"port_listening", "process_running", "http", "json_field",
		"response_time_ms", "ssl_cert", "redis_ping", "memcached_version",
		"postgres_ping", "mysql_ping", "grpc_health",
		"docker_container_running", "docker_image_exists",
		"url_reachable", "service_reachable", "s3_bucket", "version_check",
		"websocket", "otel_trace", "credential_check", "graphql",
	}
	for _, at := range types {
		t.Run(at, func(t *testing.T) {
			got := generateExpectBlock(at, nil)
			if got == "" {
				t.Errorf("generateExpectBlock(%q, nil) returned empty string", at)
			}
			// Every block should contain its assertion type name
			if !strings.Contains(got, at) {
				t.Errorf("generateExpectBlock(%q, nil) = %q, should contain type name", at, got)
			}
		})
	}
}

func TestGenerateExpectBlock_UnknownType(t *testing.T) {
	got := generateExpectBlock("nonexistent_assertion", nil)
	if !strings.Contains(got, "nonexistent_assertion") {
		t.Errorf("unknown type should mention the type name, got %q", got)
	}
	if !strings.Contains(got, "#") {
		t.Errorf("unknown type should be a comment, got %q", got)
	}
}

// --- boolArg with various string values ---

func TestBoolArg_StringValues(t *testing.T) {
	tests := []struct {
		name string
		args map[string]interface{}
		def  bool
		want bool
	}{
		{"string true", map[string]interface{}{"k": "true"}, false, false},
		{"string false", map[string]interface{}{"k": "false"}, true, true},
		{"string yes", map[string]interface{}{"k": "yes"}, false, false},
		{"string no", map[string]interface{}{"k": "no"}, true, true},
		{"string 1", map[string]interface{}{"k": "1"}, false, false},
		{"string 0", map[string]interface{}{"k": "0"}, true, true},
		{"empty string", map[string]interface{}{"k": ""}, false, false},
		{"nil map", nil, true, true},
		{"nil map false default", nil, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args map[string]interface{}
			if tt.args != nil {
				args = tt.args
			}
			got := boolArg(args, "k", tt.def)
			if got != tt.want {
				t.Errorf("boolArg() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- sanitize edge case tests ---

func TestSanitize_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"empty string", "", 10, ""},
		{"whitespace trimmed then truncated", "  hello world  ", 7, "hello w" + "\n[... truncated, full output: 11 bytes]"},
		{"single char", "x", 1, "x"},
		{"all whitespace", "   ", 10, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitize(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("sanitize(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// --- GetSuggestions extra tests ---

func TestGetSuggestions_MatchedRules(t *testing.T) {
	suggestions := GetSuggestions("http", "connection refused to host")
	if len(suggestions) == 0 {
		t.Error("expected suggestions for http connection refused")
	}
	found := false
	for _, s := range suggestions {
		if strings.Contains(s, "not listening") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected suggestion about listening, got %v", suggestions)
	}
}

func TestGetSuggestions_NoMatch(t *testing.T) {
	suggestions := GetSuggestions("http", "some unknown error that matches nothing specific")
	if len(suggestions) == 0 {
		t.Error("should return fallback suggestion even when no rule matches")
	}
	if suggestions[0] != "Check the configuration for this assertion type" {
		t.Errorf("unexpected fallback: %v", suggestions)
	}
}

func TestGetSuggestions_UnknownType(t *testing.T) {
	suggestions := GetSuggestions("totally_unknown_type", "anything")
	if len(suggestions) != 1 {
		t.Errorf("expected 1 fallback suggestion, got %d", len(suggestions))
	}
	if suggestions[0] != "Check the configuration for this assertion type" {
		t.Errorf("unexpected fallback: %v", suggestions)
	}
}

func TestGetSuggestions_AllTypes(t *testing.T) {
	types := []string{
		"redis_ping", "http", "port_listening", "postgres_ping",
		"mysql_ping", "memcached_version", "ssl_cert", "exit_code",
		"stdout_contains", "stderr_contains", "docker_container_running",
		"docker_image_exists", "grpc_health", "url_reachable",
		"websocket", "credential_check", "s3_bucket", "version_check",
		"graphql", "otel_trace",
	}
	for _, at := range types {
		t.Run(at, func(t *testing.T) {
			suggestions := GetSuggestions(at, "connection refused error")
			if len(suggestions) == 0 {
				t.Errorf("GetSuggestions(%q, ...) returned no suggestions", at)
			}
		})
	}
}

// --- generateTestYAML with environment variable template patterns ---

func TestGenerateTestYAML_EnvVarPattern(t *testing.T) {
	got := generateTestYAML("api with env", "http", map[string]interface{}{
		"url": "http://{{ .Env.HOST }}:{{ .Env.PORT }}/health",
	}, nil)
	if !strings.Contains(got, "api with env") {
		t.Error("should contain test name")
	}
	if !strings.Contains(got, "expect:") {
		t.Error("should contain expect block")
	}
	// The template pattern should be passed through as-is
	if !strings.Contains(got, "{{ .Env.HOST }}") {
		t.Errorf("template pattern should be preserved, got: %s", got)
	}
}

func TestGenerateTestYAML_MultipleEnvVars(t *testing.T) {
	got := generateTestYAML("env test", "service_reachable", map[string]interface{}{
		"url": "https://{{ .Env.API_HOST }}:{{ .Env.API_PORT }}/status",
	}, []string{"integration"})
	if !strings.Contains(got, "tags:") {
		t.Error("should contain tags block")
	}
	if !strings.Contains(got, "- integration") {
		t.Error("should contain integration tag")
	}
	if !strings.Contains(got, "{{ .Env.API_HOST }}") {
		t.Errorf("env var template should be preserved, got: %s", got)
	}
}

// --- generateTestYAML no tags ---

func TestGenerateTestYAML_NoTags(t *testing.T) {
	got := generateTestYAML("simple test", "exit_code", nil, nil)
	if strings.Contains(got, "tags:") {
		t.Errorf("should not contain tags block when no tags provided, got: %s", got)
	}
	if !strings.Contains(got, "simple test") {
		t.Error("should contain test name")
	}
}
