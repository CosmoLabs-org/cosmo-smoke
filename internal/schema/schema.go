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
	Timeout  Duration `yaml:"timeout,omitempty"`
	FailFast bool     `yaml:"fail_fast,omitempty"`
	Parallel bool     `yaml:"parallel,omitempty"`
}

// Prerequisite is a command that must succeed before tests run.
type Prerequisite struct {
	Name  string `yaml:"name"`
	Check string `yaml:"check"`
	Hint  string `yaml:"hint,omitempty"`
}

// Test defines a single smoke test.
type Test struct {
	Name         string   `yaml:"name"`
	Run          string   `yaml:"run"`
	Expect       Expect   `yaml:"expect"`
	Tags         []string `yaml:"tags,omitempty"`
	Timeout      Duration `yaml:"timeout,omitempty"`
	Cleanup      string   `yaml:"cleanup,omitempty"`
	AllowFailure bool     `yaml:"allow_failure,omitempty"`
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
	JSONField      *JSONFieldCheck `yaml:"json_field,omitempty"`
	ResponseTimeMs *int            `yaml:"response_time_ms,omitempty"` // Fail if test duration exceeds this many ms
}

// PortCheck defines parameters for checking if a port is open and listening.
type PortCheck struct {
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol,omitempty"`
	Host     string `yaml:"host,omitempty"`
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

// JSONFieldCheck defines parameters for asserting on JSON fields in stdout.
type JSONFieldCheck struct {
	Path     string `yaml:"path"`
	Equals   string `yaml:"equals,omitempty"`
	Contains string `yaml:"contains,omitempty"`
	Matches  string `yaml:"matches,omitempty"`
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
