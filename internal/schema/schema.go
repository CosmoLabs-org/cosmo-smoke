package schema

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// SmokeConfig is the top-level configuration parsed from .smoke.yaml.
type SmokeConfig struct {
	Version     int            `yaml:"version"`
	Project     string         `yaml:"project"`
	Description string         `yaml:"description,omitempty"`
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
	Name    string   `yaml:"name"`
	Run     string   `yaml:"run"`
	Expect  Expect   `yaml:"expect"`
	Tags    []string `yaml:"tags,omitempty"`
	Timeout Duration `yaml:"timeout,omitempty"`
	Cleanup string   `yaml:"cleanup,omitempty"`
}

// Expect defines the assertions for a test.
type Expect struct {
	ExitCode       *int   `yaml:"exit_code,omitempty"`
	StdoutContains string `yaml:"stdout_contains,omitempty"`
	StdoutMatches  string `yaml:"stdout_matches,omitempty"`
	StderrContains string `yaml:"stderr_contains,omitempty"`
	FileExists     string `yaml:"file_exists,omitempty"`
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
func Load(path string) (*SmokeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	return Parse(data)
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
