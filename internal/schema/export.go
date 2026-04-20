package schema

import (
	"encoding/json"
)

// SchemaOutput is the top-level structure exported by `smoke schema`.
type SchemaOutput struct {
	Version        string            `json:"version"`
	AssertionTypes []AssertionSchema `json:"assertion_types"`
}

// AssertionSchema describes one assertion type's fields.
type AssertionSchema struct {
	Name   string       `json:"name"`
	YAML   string       `json:"yaml_field"`
	Fields []FieldInfo  `json:"fields"`
}

// FieldInfo describes one field within an assertion.
type FieldInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// ExportSchema returns a JSON-serializable description of all assertion types.
func ExportSchema() *SchemaOutput {
	return &SchemaOutput{
		Version: "1",
		AssertionTypes: []AssertionSchema{
			{
				Name: "exit_code", YAML: "exit_code",
				Fields: []FieldInfo{{Name: "value", Type: "int", Required: true}},
			},
			{Name: "stdout_contains", YAML: "stdout_contains", Fields: []FieldInfo{{Name: "value", Type: "string", Required: true}}},
			{Name: "stdout_matches", YAML: "stdout_matches", Fields: []FieldInfo{{Name: "pattern", Type: "string (regex)", Required: true}}},
			{Name: "stderr_contains", YAML: "stderr_contains", Fields: []FieldInfo{{Name: "value", Type: "string", Required: true}}},
			{Name: "stderr_matches", YAML: "stderr_matches", Fields: []FieldInfo{{Name: "pattern", Type: "string (regex)", Required: true}}},
			{Name: "file_exists", YAML: "file_exists", Fields: []FieldInfo{{Name: "path", Type: "string", Required: true}}},
			{Name: "env_exists", YAML: "env_exists", Fields: []FieldInfo{{Name: "name", Type: "string", Required: true}}},
			{
				Name: "port_listening", YAML: "port_listening",
				Fields: []FieldInfo{
					{Name: "port", Type: "int", Required: true},
					{Name: "protocol", Type: "string", Required: false},
					{Name: "host", Type: "string", Required: false},
				},
			},
			{Name: "process_running", YAML: "process_running", Fields: []FieldInfo{{Name: "name", Type: "string", Required: true}}},
			{
				Name: "http", YAML: "http",
				Fields: []FieldInfo{
					{Name: "url", Type: "string", Required: true},
					{Name: "method", Type: "string", Required: false},
					{Name: "headers", Type: "map[string]string", Required: false},
					{Name: "body", Type: "string", Required: false},
					{Name: "timeout", Type: "duration", Required: false},
					{Name: "status_code", Type: "int", Required: false},
					{Name: "body_contains", Type: "string", Required: false},
					{Name: "body_matches", Type: "string (regex)", Required: false},
					{Name: "header_contains", Type: "map[string]string", Required: false},
				},
			},
			{
				Name: "json_field", YAML: "json_field",
				Fields: []FieldInfo{
					{Name: "path", Type: "string (JSONPath)", Required: true},
					{Name: "equals", Type: "string", Required: false},
					{Name: "contains", Type: "string", Required: false},
					{Name: "matches", Type: "string (regex)", Required: false},
				},
			},
			{Name: "response_time_ms", YAML: "response_time_ms", Fields: []FieldInfo{{Name: "value", Type: "int", Required: true}}},
			{
				Name: "ssl_cert", YAML: "ssl_cert",
				Fields: []FieldInfo{
					{Name: "host", Type: "string", Required: true},
					{Name: "port", Type: "int", Required: false},
					{Name: "min_days_remaining", Type: "int", Required: false},
					{Name: "allow_self_signed", Type: "bool", Required: false},
				},
			},
			{
				Name: "redis_ping", YAML: "redis_ping",
				Fields: []FieldInfo{
					{Name: "host", Type: "string", Required: false},
					{Name: "port", Type: "int", Required: false},
					{Name: "password", Type: "string", Required: false},
				},
			},
			{
				Name: "memcached_version", YAML: "memcached_version",
				Fields: []FieldInfo{
					{Name: "host", Type: "string", Required: false},
					{Name: "port", Type: "int", Required: false},
				},
			},
			{
				Name: "postgres_ping", YAML: "postgres_ping",
				Fields: []FieldInfo{
					{Name: "host", Type: "string", Required: false},
					{Name: "port", Type: "int", Required: false},
				},
			},
			{
				Name: "mysql_ping", YAML: "mysql_ping",
				Fields: []FieldInfo{
					{Name: "host", Type: "string", Required: false},
					{Name: "port", Type: "int", Required: false},
				},
			},
			{
				Name: "grpc_health", YAML: "grpc_health",
				Fields: []FieldInfo{
					{Name: "address", Type: "string (host:port)", Required: true},
					{Name: "service", Type: "string", Required: false},
					{Name: "use_tls", Type: "bool", Required: false},
					{Name: "timeout", Type: "duration", Required: false},
				},
			},
			{
				Name: "docker_container_running", YAML: "docker_container_running",
				Fields: []FieldInfo{
					{Name: "name", Type: "string", Required: true},
				},
			},
			{
				Name: "docker_image_exists", YAML: "docker_image_exists",
				Fields: []FieldInfo{
					{Name: "image", Type: "string", Required: true},
				},
			},
			{
				Name: "url_reachable", YAML: "url_reachable",
				Fields: []FieldInfo{
					{Name: "url", Type: "string", Required: true},
					{Name: "timeout", Type: "duration", Required: false},
					{Name: "status_code", Type: "int", Required: false},
				},
			},
			{
				Name: "service_reachable", YAML: "service_reachable",
				Fields: []FieldInfo{
					{Name: "url", Type: "string", Required: true},
					{Name: "timeout", Type: "duration", Required: false},
				},
			},
			{
				Name: "s3_bucket", YAML: "s3_bucket",
				Fields: []FieldInfo{
					{Name: "bucket", Type: "string", Required: true},
					{Name: "region", Type: "string", Required: false},
					{Name: "endpoint", Type: "string", Required: false},
				},
			},
			{
				Name: "version_check", YAML: "version_check",
				Fields: []FieldInfo{
					{Name: "command", Type: "string", Required: true},
					{Name: "pattern", Type: "string (regex)", Required: true},
				},
			},
			{
				Name: "websocket", YAML: "websocket",
				Fields: []FieldInfo{
					{Name: "url", Type: "string (ws:// or wss://)", Required: true},
					{Name: "send", Type: "string", Required: false},
					{Name: "expect_contains", Type: "string", Required: false},
					{Name: "expect_matches", Type: "string (regex)", Required: false},
					{Name: "timeout", Type: "duration", Required: false},
					{Name: "headers", Type: "map[string]string", Required: false},
				},
			},
			{
				Name: "otel_trace", YAML: "otel_trace",
				Fields: []FieldInfo{
					{Name: "backend", Type: "string (jaeger|tempo|honeycomb|datadog)", Required: false},
					{Name: "jaeger_url", Type: "string", Required: false},
					{Name: "service_name", Type: "string", Required: false},
					{Name: "min_spans", Type: "int", Required: false},
					{Name: "timeout", Type: "duration", Required: false},
					{Name: "api_key", Type: "string", Required: false},
					{Name: "dd_app_key", Type: "string", Required: false},
				},
			},
			{
				Name: "credential_check", YAML: "credential_check",
				Fields: []FieldInfo{
					{Name: "source", Type: "string (env|file|exec)", Required: true},
					{Name: "name", Type: "string", Required: true},
					{Name: "contains", Type: "string", Required: false},
				},
			},
			{
				Name: "graphql", YAML: "graphql",
				Fields: []FieldInfo{
					{Name: "url", Type: "string", Required: true},
					{Name: "query", Type: "string", Required: false},
					{Name: "status_code", Type: "int", Required: false},
					{Name: "expect_types", Type: "[]string", Required: false},
					{Name: "expect_contains", Type: "string", Required: false},
					{Name: "timeout", Type: "duration", Required: false},
				},
			},
		},
	}
}

// ExportSchemaJSON returns the schema as formatted JSON.
func ExportSchemaJSON() ([]byte, error) {
	return json.MarshalIndent(ExportSchema(), "", "  ")
}
