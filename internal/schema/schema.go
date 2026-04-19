package schema

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

// SmokeConfig is the top-level configuration parsed from .smoke.yaml.
type SmokeConfig struct {
	Version     int            `yaml:"version"`
	Project     string         `yaml:"project"`
	Description string         `yaml:"description,omitempty"`
	Includes    []string       `yaml:"includes,omitempty"`
	Settings    Settings       `yaml:"settings,omitempty"`
	Prereqs     []Prerequisite `yaml:"prerequisites,omitempty"`
	Tests       []Test         `yaml:"tests"`
}

// Settings controls global test behavior.
type Settings struct {
	Timeout         Duration `yaml:"timeout,omitempty"`
	FailFast        bool     `yaml:"fail_fast,omitempty"`
	Parallel        bool     `yaml:"parallel,omitempty"`
	Monorepo        bool     `yaml:"monorepo,omitempty"`
	MonorepoExclude []string `yaml:"monorepo_exclude,omitempty"`
}

// Prerequisite is a command that must succeed before tests run.
type Prerequisite struct {
	Name  string `yaml:"name"`
	Check string `yaml:"check"`
	Hint  string `yaml:"hint,omitempty"`
}

// RetryPolicy configures automatic retry for flaky tests.
type RetryPolicy struct {
	Count   int      `yaml:"count"`
	Backoff Duration `yaml:"backoff"`
}

// Test defines a single smoke test.
type Test struct {
	Name         string       `yaml:"name"`
	Run          string       `yaml:"run"`
	Expect       Expect       `yaml:"expect"`
	Tags         []string     `yaml:"tags,omitempty"`
	Timeout      Duration     `yaml:"timeout,omitempty"`
	Cleanup      string       `yaml:"cleanup,omitempty"`
	AllowFailure bool         `yaml:"allow_failure,omitempty"`
	Retry        *RetryPolicy `yaml:"retry,omitempty"`
	SkipIf       *SkipIf      `yaml:"skip_if,omitempty"`
}

// SkipIf defines conditions under which a test should be skipped.
type SkipIf struct {
	EnvUnset  string         `yaml:"env_unset,omitempty"`
	EnvEquals *EnvEqualsCond `yaml:"env_equals,omitempty"`
	FileMissing string       `yaml:"file_missing,omitempty"`
}

// EnvEqualsCond checks if an env var equals a specific value.
type EnvEqualsCond struct {
	Var   string `yaml:"var"`
	Value string `yaml:"value"`
}

// Expect defines the assertions for a test.
type Expect struct {
	ExitCode       *int            `yaml:"exit_code,omitempty"`
	StdoutContains string          `yaml:"stdout_contains,omitempty"`
	StdoutMatches  string          `yaml:"stdout_matches,omitempty"`
	StderrContains string          `yaml:"stderr_contains,omitempty"`
	StderrMatches  string          `yaml:"stderr_matches,omitempty"`
	FileExists     string          `yaml:"file_exists,omitempty"`
	EnvExists      string          `yaml:"env_exists,omitempty"`
	PortListening  *PortCheck      `yaml:"port_listening,omitempty"`
	ProcessRunning string          `yaml:"process_running,omitempty"`
	HTTP           *HTTPCheck      `yaml:"http,omitempty"`
	JSONField      *JSONFieldCheck  `yaml:"json_field,omitempty"`
	ResponseTimeMs *int             `yaml:"response_time_ms,omitempty"` // Fail if test duration exceeds this many ms
	SSLCert        *SSLCertCheck    `yaml:"ssl_cert,omitempty"`
	Redis          *RedisCheck      `yaml:"redis_ping,omitempty"`
	Memcached      *MemcachedCheck  `yaml:"memcached_version,omitempty"`
	Postgres       *PostgresCheck   `yaml:"postgres_ping,omitempty"`
	MySQL          *MySQLCheck      `yaml:"mysql_ping,omitempty"`
	GRPCHealth      *GRPCHealthCheck      `yaml:"grpc_health,omitempty"`
	DockerContainer  *DockerContainerCheck  `yaml:"docker_container_running,omitempty"`
	DockerImage      *DockerImageCheck      `yaml:"docker_image_exists,omitempty"`
	URLReachable     *URLReachableCheck     `yaml:"url_reachable,omitempty"`
	ServiceReachable *ServiceReachableCheck `yaml:"service_reachable,omitempty"`
	S3Bucket         *S3BucketCheck         `yaml:"s3_bucket,omitempty"`
	VersionCheck     *VersionCheck          `yaml:"version_check,omitempty"`
	WebSocket        *WebSocketCheck        `yaml:"websocket,omitempty"`
}

// PortCheck defines parameters for checking if a port is open and listening.
type PortCheck struct {
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol,omitempty"`
	Host     string `yaml:"host,omitempty"`
}

// SSLCertCheck defines parameters for TLS certificate validation.
type SSLCertCheck struct {
	Host             string `yaml:"host"`
	Port             int    `yaml:"port,omitempty"`               // defaults to 443
	MinDaysRemaining int    `yaml:"min_days_remaining,omitempty"` // 0 = any non-expired cert passes
	AllowSelfSigned  bool   `yaml:"allow_self_signed,omitempty"`
}

// RedisCheck pings a Redis server with PING and verifies PONG reply.
type RedisCheck struct {
	Host     string `yaml:"host,omitempty"`     // default "localhost"
	Port     int    `yaml:"port,omitempty"`     // default 6379
	Password string `yaml:"password,omitempty"` // optional AUTH
}

// MemcachedCheck issues `version` to a Memcached server and expects a VERSION reply.
type MemcachedCheck struct {
	Host string `yaml:"host,omitempty"` // default "localhost"
	Port int    `yaml:"port,omitempty"` // default 11211
}

// PostgresCheck pings a Postgres server via SSLRequest handshake.
type PostgresCheck struct {
	Host string `yaml:"host,omitempty"` // default "localhost"
	Port int    `yaml:"port,omitempty"` // default 5432
}

// MySQLCheck verifies a MySQL server sends a valid handshake on connection.
type MySQLCheck struct {
	Host string `yaml:"host,omitempty"` // default "localhost"
	Port int    `yaml:"port,omitempty"` // default 3306
}

// DockerContainerCheck verifies a named Docker container is running.
type DockerContainerCheck struct {
	Name string `yaml:"name"`
}

// DockerImageCheck verifies a Docker image exists locally.
type DockerImageCheck struct {
	Image string `yaml:"image"`
}

// HTTPCheck defines parameters for HTTP endpoint assertions.
type HTTPCheck struct {
	URL            string            `yaml:"url"`
	Method         string            `yaml:"method,omitempty"`
	Headers        map[string]string `yaml:"headers,omitempty"`
	Body           string            `yaml:"body,omitempty"`
	Timeout        Duration          `yaml:"timeout,omitempty"`
	StatusCode     *int              `yaml:"status_code,omitempty"`
	BodyContains   string            `yaml:"body_contains,omitempty"`
	BodyMatches    string            `yaml:"body_matches,omitempty"`
	HeaderContains map[string]string `yaml:"header_contains,omitempty"`
}

// GRPCHealthCheck queries the grpc.health.v1.Health/Check endpoint.
type GRPCHealthCheck struct {
	Address string   `yaml:"address"`           // host:port
	Service string   `yaml:"service,omitempty"` // "" = overall server health
	UseTLS  bool     `yaml:"use_tls,omitempty"` // default false (insecure)
	Timeout Duration `yaml:"timeout,omitempty"` // default 5s
}

// JSONFieldCheck defines parameters for asserting on JSON fields in stdout.
type JSONFieldCheck struct {
	Path     string `yaml:"path"`
	Equals   string `yaml:"equals,omitempty"`
	Contains string `yaml:"contains,omitempty"`
	Matches  string `yaml:"matches,omitempty"`
}

// URLReachableCheck verifies an HTTP/HTTPS endpoint is accessible.
type URLReachableCheck struct {
	URL        string   `yaml:"url"`
	Timeout    Duration `yaml:"timeout,omitempty"`
	StatusCode *int     `yaml:"status_code,omitempty"`
}

// ServiceReachableCheck verifies an external service dependency is accessible.
type ServiceReachableCheck struct {
	URL     string   `yaml:"url"`
	Timeout Duration `yaml:"timeout,omitempty"`
}

// S3BucketCheck verifies an S3-compatible bucket is accessible via anonymous HEAD.
type S3BucketCheck struct {
	Bucket   string `yaml:"bucket"`
	Region   string `yaml:"region,omitempty"`
	Endpoint string `yaml:"endpoint,omitempty"`
}

// VersionCheck verifies an installed tool matches a required version pattern.
type VersionCheck struct {
	Command string `yaml:"command"`
	Pattern string `yaml:"pattern"`
}

// WebSocketCheck verifies a WebSocket endpoint is reachable and responds as expected.
type WebSocketCheck struct {
	URL            string   `yaml:"url"`
	Send           string   `yaml:"send,omitempty"`
	ExpectContains string   `yaml:"expect_contains,omitempty"`
	ExpectMatches  string   `yaml:"expect_matches,omitempty"`
	Timeout        Duration `yaml:"timeout,omitempty"`
}

// Duration wraps time.Duration for YAML unmarshaling from strings like "5s".
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = parsed
	return nil
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}

// Load reads and parses a .smoke.yaml file from the given path.
// Supports Go templates ({{ .Env.FOO }}) and includes.
func Load(path string) (*SmokeConfig, error) {
	return loadWithDepth(path, 0)
}

func loadWithDepth(path string, depth int) (*SmokeConfig, error) {
	if depth > 10 {
		return nil, fmt.Errorf("include depth exceeded (max 10): circular includes?")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	// Process Go templates
	processed, err := processTemplate(data)
	if err != nil {
		return nil, fmt.Errorf("processing template: %w", err)
	}

	cfg, err := Parse(processed)
	if err != nil {
		return nil, err
	}

	// Process includes
	configDir := filepath.Dir(path)
	for _, inc := range cfg.Includes {
		incPath := inc
		if !filepath.IsAbs(inc) {
			incPath = filepath.Join(configDir, inc)
		}

		incCfg, err := loadWithDepth(incPath, depth+1)
		if err != nil {
			return nil, fmt.Errorf("loading include %q: %w", inc, err)
		}

		// Merge: included tests and prereqs are prepended
		cfg.Prereqs = append(incCfg.Prereqs, cfg.Prereqs...)
		cfg.Tests = append(incCfg.Tests, cfg.Tests...)
	}

	return cfg, nil
}

// processTemplate expands Go templates in the config.
// Available: .Env (environment variables map)
func processTemplate(data []byte) ([]byte, error) {
	tmpl, err := template.New("config").Parse(string(data))
	if err != nil {
		return nil, err
	}

	// Build template data
	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		if idx := bytes.IndexByte([]byte(e), '='); idx > 0 {
			envMap[e[:idx]] = e[idx+1:]
		}
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]any{
		"Env": envMap,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Parse parses raw YAML bytes into a SmokeConfig.
func Parse(data []byte) (*SmokeConfig, error) {
	var cfg SmokeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// LoadDefault finds and loads .smoke.yaml from the current directory.
func LoadDefault() (*SmokeConfig, error) {
	return Load(".smoke.yaml")
}

// MergeEnv loads an environment-specific config and deep-merges it onto base.
// Env-specific tests are appended (not replaced). Settings from env override base.
func MergeEnv(base *SmokeConfig, envPath string) (*SmokeConfig, error) {
	envCfg, err := Load(envPath)
	if err != nil {
		return nil, fmt.Errorf("loading env config %s: %w", envPath, err)
	}

	// Deep merge: env settings override base
	if envCfg.Settings.Timeout.Duration > 0 {
		base.Settings.Timeout = envCfg.Settings.Timeout
	}
	if envCfg.Settings.FailFast {
		base.Settings.FailFast = true
	}
	if envCfg.Settings.Parallel {
		base.Settings.Parallel = true
	}

	// Prepend env prereqs (they run before base prereqs)
	base.Prereqs = append(envCfg.Prereqs, base.Prereqs...)

	// Append env tests (they run after base tests)
	base.Tests = append(base.Tests, envCfg.Tests...)

	return base, nil
}
