package mcp

// assertionDocs maps assertion type names to their documentation.
var assertionDocs = map[string]*ExplainResult{
	"exit_code": {
		Type:        "exit_code",
		Description: "Verifies the process exit code matches the expected value",
		Fields: []ExplainField{
			{Name: "exit_code", Type: "int", Required: false, Default: "0", Description: "Expected exit code"},
		},
		Example: `tests:
  - name: command succeeds
    run: echo hello
    expect:
      exit_code: 0`,
	},
	"stdout_contains": {
		Type:        "stdout_contains",
		Description: "Checks that stdout contains a specific substring",
		Fields: []ExplainField{
			{Name: "stdout_contains", Type: "string", Required: true, Description: "Substring to search for in stdout"},
		},
		Example: `tests:
  - name: output contains hello
    run: echo "hello world"
    expect:
      stdout_contains: "hello"`,
	},
	"stdout_matches": {
		Type:        "stdout_matches",
		Description: "Checks that stdout matches a regular expression",
		Fields: []ExplainField{
			{Name: "stdout_matches", Type: "string", Required: true, Description: "Regex pattern to match against stdout"},
		},
		Example: `tests:
  - name: version output matches pattern
    run: myapp --version
    expect:
      stdout_matches: "v\\d+\\.\\d+\\.\\d+"`,
	},
	"stderr_contains": {
		Type:        "stderr_contains",
		Description: "Checks that stderr contains a specific substring",
		Fields: []ExplainField{
			{Name: "stderr_contains", Type: "string", Required: true, Description: "Substring to search for in stderr"},
		},
		Example: `tests:
  - name: error message in stderr
    run: myapp --invalid
    expect:
      stderr_contains: "unknown flag"`,
	},
	"stderr_matches": {
		Type:        "stderr_matches",
		Description: "Checks that stderr matches a regular expression",
		Fields: []ExplainField{
			{Name: "stderr_matches", Type: "string", Required: true, Description: "Regex pattern to match against stderr"},
		},
		Example: `tests:
  - name: error matches pattern
    run: myapp --invalid
    expect:
      stderr_matches: "error:.*flag"`,
	},
	"file_exists": {
		Type:        "file_exists",
		Description: "Checks that a file exists relative to the config directory",
		Fields: []ExplainField{
			{Name: "file_exists", Type: "string", Required: true, Description: "File path relative to config directory"},
		},
		Example: `tests:
  - name: config file created
    run: myapp init
    expect:
      file_exists: "config.yaml"`,
	},
	"env_exists": {
		Type:        "env_exists",
		Description: "Checks that an environment variable is set",
		Fields: []ExplainField{
			{Name: "env_exists", Type: "string", Required: true, Description: "Name of the environment variable to check"},
		},
		Example: `tests:
  - name: HOME is set
    run: "true"
    expect:
      env_exists: "HOME"`,
	},
	"port_listening": {
		Type:        "port_listening",
		Description: "Checks that a TCP or UDP port is open and accepting connections",
		Fields: []ExplainField{
			{Name: "port", Type: "int", Required: true, Description: "Port number to check"},
			{Name: "protocol", Type: "string", Required: false, Default: "tcp", Description: "Protocol: tcp or udp"},
			{Name: "host", Type: "string", Required: false, Default: "localhost", Description: "Host to connect to"},
		},
		Example: `tests:
  - name: web server listening
    expect:
      port_listening:
        port: 8080
        protocol: tcp`,
	},
	"process_running": {
		Type:        "process_running",
		Description: "Checks that a named process is currently running",
		Fields: []ExplainField{
			{Name: "process_running", Type: "string", Required: true, Description: "Exact process name (matched with pgrep -x / tasklist)"},
		},
		Example: `tests:
  - name: nginx is running
    expect:
      process_running: "nginx"`,
	},
	"http": {
		Type:        "http",
		Description: "HTTP endpoint check supporting status code, body, and header assertions",
		Fields: []ExplainField{
			{Name: "url", Type: "string", Required: true, Description: "URL to request"},
			{Name: "method", Type: "string", Required: false, Default: "GET", Description: "HTTP method"},
			{Name: "status_code", Type: "int", Required: false, Default: "200", Description: "Expected HTTP status code"},
			{Name: "body_contains", Type: "string", Required: false, Description: "Substring to find in response body"},
			{Name: "body_matches", Type: "string", Required: false, Description: "Regex to match against response body"},
			{Name: "header_contains", Type: "string", Required: false, Description: "Expected header in 'Name: Value' format"},
		},
		Example: `tests:
  - name: health endpoint returns 200
    expect:
      http:
        url: "http://localhost:8080/health"
        status_code: 200
        body_contains: "ok"`,
	},
	"json_field": {
		Type:        "json_field",
		Description: "Asserts on a JSON field from stdout using dot-notation path",
		Fields: []ExplainField{
			{Name: "path", Type: "string", Required: true, Description: "JSONPath using dot notation (e.g. 'data.name')"},
			{Name: "equals", Type: "string", Required: false, Description: "Exact value match"},
			{Name: "contains", Type: "string", Required: false, Description: "Substring match on the field value"},
			{Name: "matches", Type: "string", Required: false, Description: "Regex match on the field value"},
		},
		Example: `tests:
  - name: API returns correct version
    run: curl -s http://localhost:8080/api/version
    expect:
      json_field:
        path: "version"
        equals: "1.0.0"`,
	},
	"response_time_ms": {
		Type:        "response_time_ms",
		Description: "Asserts total test duration does not exceed threshold in milliseconds",
		Fields: []ExplainField{
			{Name: "response_time_ms", Type: "int", Required: true, Description: "Maximum allowed duration in milliseconds"},
		},
		Example: `tests:
  - name: API responds quickly
    run: curl -s http://localhost:8080/api
    expect:
      response_time_ms: 500`,
	},
	"ssl_cert": {
		Type:        "ssl_cert",
		Description: "TLS certificate validity and expiry check",
		Fields: []ExplainField{
			{Name: "host", Type: "string", Required: true, Description: "Hostname to check"},
			{Name: "port", Type: "int", Required: false, Default: "443", Description: "TLS port"},
			{Name: "min_days_remaining", Type: "int", Required: false, Default: "7", Description: "Minimum days before expiry"},
			{Name: "allow_self_signed", Type: "bool", Required: false, Default: "false", Description: "Accept self-signed certificates"},
		},
		Example: `tests:
  - name: TLS cert is valid
    expect:
      ssl_cert:
        host: "example.com"
        min_days_remaining: 30`,
	},
	"redis_ping": {
		Type:        "redis_ping",
		Description: "Redis PING command returns +PONG via RESP protocol",
		Fields: []ExplainField{
			{Name: "host", Type: "string", Required: false, Default: "localhost", Description: "Redis host"},
			{Name: "port", Type: "int", Required: false, Default: "6379", Description: "Redis port"},
			{Name: "password", Type: "string", Required: false, Description: "Redis AUTH password"},
		},
		Example: `tests:
  - name: Redis is reachable
    expect:
      redis_ping:
        host: "localhost"
        port: 6379`,
	},
	"memcached_version": {
		Type:        "memcached_version",
		Description: "Memcached version command returns VERSION response",
		Fields: []ExplainField{
			{Name: "host", Type: "string", Required: false, Default: "localhost", Description: "Memcached host"},
			{Name: "port", Type: "int", Required: false, Default: "11211", Description: "Memcached port"},
		},
		Example: `tests:
  - name: Memcached is reachable
    expect:
      memcached_version:
        host: "localhost"`,
	},
	"postgres_ping": {
		Type:        "postgres_ping",
		Description: "Postgres server SSLRequest handshake returns valid protocol byte",
		Fields: []ExplainField{
			{Name: "host", Type: "string", Required: false, Default: "localhost", Description: "Postgres host"},
			{Name: "port", Type: "int", Required: false, Default: "5432", Description: "Postgres port"},
		},
		Example: `tests:
  - name: Postgres is reachable
    expect:
      postgres_ping:
        host: "localhost"`,
	},
	"mysql_ping": {
		Type:        "mysql_ping",
		Description: "MySQL server sends valid v10 handshake packet on connection",
		Fields: []ExplainField{
			{Name: "host", Type: "string", Required: false, Default: "localhost", Description: "MySQL host"},
			{Name: "port", Type: "int", Required: false, Default: "3306", Description: "MySQL port"},
		},
		Example: `tests:
  - name: MySQL is reachable
    expect:
      mysql_ping:
        host: "localhost"`,
	},
	"grpc_health": {
		Type:        "grpc_health",
		Description: "grpc.health.v1 Health/Check returns SERVING status (requires -tags grpc build)",
		Fields: []ExplainField{
			{Name: "address", Type: "string", Required: true, Description: "gRPC server address (host:port)"},
			{Name: "service", Type: "string", Required: false, Description: "Specific service name to check"},
			{Name: "use_tls", Type: "bool", Required: false, Default: "false", Description: "Use TLS connection"},
			{Name: "timeout", Type: "string", Required: false, Default: "10s", Description: "Connection timeout"},
		},
		Example: `tests:
  - name: gRPC health check
    expect:
      grpc_health:
        address: "localhost:9090"
        service: "myservice.MyService"`,
		Notes: "Requires building with -tags grpc to include gRPC dependencies",
	},
	"docker_container_running": {
		Type:        "docker_container_running",
		Description: "Checks that a named Docker container is currently running",
		Fields: []ExplainField{
			{Name: "name", Type: "string", Required: true, Description: "Docker container name"},
		},
		Example: `tests:
  - name: postgres container is running
    expect:
      docker_container_running:
        name: "my-postgres"`,
	},
	"docker_image_exists": {
		Type:        "docker_image_exists",
		Description: "Checks that a Docker image exists locally",
		Fields: []ExplainField{
			{Name: "image", Type: "string", Required: true, Description: "Docker image name (e.g. 'nginx:latest')"},
		},
		Example: `tests:
  - name: base image exists
    expect:
      docker_image_exists:
        image: "nginx:alpine"`,
	},
	"url_reachable": {
		Type:        "url_reachable",
		Description: "HTTP/HTTPS connectivity check with optional status code validation",
		Fields: []ExplainField{
			{Name: "url", Type: "string", Required: true, Description: "URL to check"},
			{Name: "timeout", Type: "string", Required: false, Default: "10s", Description: "Request timeout"},
			{Name: "status_code", Type: "int", Required: false, Default: "200", Description: "Expected status code"},
		},
		Example: `tests:
  - name: external API is reachable
    expect:
      url_reachable:
        url: "https://api.example.com/health"
        timeout: "5s"`,
	},
	"service_reachable": {
		Type:        "service_reachable",
		Description: "External service dependency connectivity check",
		Fields: []ExplainField{
			{Name: "url", Type: "string", Required: true, Description: "Service URL to check"},
			{Name: "timeout", Type: "string", Required: false, Default: "10s", Description: "Request timeout"},
		},
		Example: `tests:
  - name: payment service reachable
    expect:
      service_reachable:
        url: "https://payments.internal/health"`,
	},
	"s3_bucket": {
		Type:        "s3_bucket",
		Description: "S3-compatible bucket accessibility check via anonymous HEAD request",
		Fields: []ExplainField{
			{Name: "bucket", Type: "string", Required: true, Description: "Bucket name"},
			{Name: "region", Type: "string", Required: false, Default: "us-east-1", Description: "AWS region"},
			{Name: "endpoint", Type: "string", Required: false, Description: "Custom S3-compatible endpoint URL"},
		},
		Example: `tests:
  - name: assets bucket accessible
    expect:
      s3_bucket:
        bucket: "my-assets"
        region: "us-east-1"`,
	},
	"version_check": {
		Type:        "version_check",
		Description: "Tool version verification via shell command and regex pattern",
		Fields: []ExplainField{
			{Name: "command", Type: "string", Required: true, Description: "Shell command that outputs version info"},
			{Name: "pattern", Type: "string", Required: true, Description: "Regex to extract version from command output"},
		},
		Example: `tests:
  - name: Go version check
    expect:
      version_check:
        command: "go version"
        pattern: "go(\\d+\\.\\d+\\.\\d+)"`,
	},
	"otel_trace": {
		Type:        "otel_trace",
		Description: "OpenTelemetry trace verification with W3C traceparent propagation. Supports Jaeger, Tempo, Honeycomb, and Datadog backends",
		Fields: []ExplainField{
			{Name: "backend", Type: "string", Required: false, Default: "jaeger", Description: "Trace backend: jaeger, tempo, honeycomb, datadog"},
			{Name: "jaeger_url", Type: "string", Required: false, Description: "Jaeger/collector URL"},
			{Name: "service_name", Type: "string", Required: false, Description: "Service name to filter traces"},
			{Name: "min_spans", Type: "int", Required: false, Default: "1", Description: "Minimum number of expected spans"},
			{Name: "timeout", Type: "string", Required: false, Default: "10s", Description: "Trace lookup timeout"},
			{Name: "api_key", Type: "string", Required: false, Description: "API key for honeycomb/datadog"},
			{Name: "dd_app_key", Type: "string", Required: false, Description: "Datadog application key"},
		},
		Example: `otel:
  enabled: true
  jaeger_url: "http://jaeger:16686"
tests:
  - name: trace received
    run: curl http://localhost:8080/api
    expect:
      otel_trace:
        service_name: "my-api"
        min_spans: 2`,
	},
	"websocket": {
		Type:        "websocket",
		Description: "WebSocket connect-send-expect assertion using stdlib-only client",
		Fields: []ExplainField{
			{Name: "url", Type: "string", Required: true, Description: "WebSocket URL (ws:// or wss://)"},
			{Name: "send", Type: "string", Required: false, Description: "Message to send after connecting"},
			{Name: "expect_contains", Type: "string", Required: false, Description: "Substring to find in received message"},
			{Name: "expect_matches", Type: "string", Required: false, Description: "Regex to match against received message"},
			{Name: "timeout", Type: "string", Required: false, Default: "10s", Description: "Wait timeout for expected message"},
		},
		Example: `tests:
  - name: WebSocket echo
    expect:
      websocket:
        url: "ws://localhost:8080/ws"
        send: "hello"
        expect_contains: "hello"`,
	},
	"credential_check": {
		Type:        "credential_check",
		Description: "Verifies a credential is accessible without leaking its value",
		Fields: []ExplainField{
			{Name: "source", Type: "string", Required: true, Description: "Source type: env, file, or exec"},
			{Name: "name", Type: "string", Required: true, Description: "Variable name, file path, or command"},
			{Name: "contains", Type: "string", Required: false, Description: "Substring that must be present in the value"},
		},
		Example: `tests:
  - name: API key is set
    expect:
      credential_check:
        source: "env"
        name: "API_KEY"`,
	},
	"graphql": {
		Type:        "graphql",
		Description: "GraphQL introspection and query assertion",
		Fields: []ExplainField{
			{Name: "url", Type: "string", Required: true, Description: "GraphQL endpoint URL"},
			{Name: "query", Type: "string", Required: false, Description: "GraphQL query to execute"},
			{Name: "status_code", Type: "int", Required: false, Default: "200", Description: "Expected HTTP status code"},
			{Name: "expect_types", Type: "[]string", Required: false, Description: "Type names expected in schema"},
			{Name: "expect_contains", Type: "string", Required: false, Description: "Substring expected in response"},
			{Name: "timeout", Type: "string", Required: false, Default: "10s", Description: "Request timeout"},
		},
		Example: `tests:
  - name: GraphQL schema has User type
    expect:
      graphql:
        url: "http://localhost:8080/graphql"
        expect_types: ["User"]`,
	},
}
