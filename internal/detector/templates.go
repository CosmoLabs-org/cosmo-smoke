package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

// intPtr returns a pointer to an int literal.
func intPtr(n int) *int { return &n }

// duration30s is a convenience 30-second Duration.
var duration30s = schema.Duration{Duration: 30 * time.Second}

// hasPkgScript checks whether package.json contains a scripts entry with the given key.
func hasPkgScript(dir, script string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}
	_, ok := pkg.Scripts[script]
	return ok
}

// GenerateConfig creates a SmokeConfig from detected project types.
// projectName is derived from the directory name.
func GenerateConfig(dir string, types []ProjectType) *schema.SmokeConfig {
	cfg := &schema.SmokeConfig{
		Version: 1,
		Project: filepath.Base(dir),
		Settings: schema.Settings{
			Timeout:  duration30s,
			FailFast: true,
		},
	}

	for _, t := range types {
		switch t {
		case Go:
			cfg.Prereqs = append(cfg.Prereqs, schema.Prerequisite{
				Name:  "Go installed",
				Check: "go version",
				Hint:  "Install Go from https://go.dev/dl/",
			})
			cfg.Tests = append(cfg.Tests,
				schema.Test{
					Name: "Compiles",
					Run:  "go build ./...",
					Expect: schema.Expect{
						ExitCode: intPtr(0),
					},
				},
				schema.Test{
					Name: "Tests pass",
					Run:  "go test ./...",
					Expect: schema.Expect{
						ExitCode: intPtr(0),
					},
				},
			)

		case Node:
			if HasBun(dir) {
				cfg.Prereqs = append(cfg.Prereqs, schema.Prerequisite{
					Name:  "Bun installed",
					Check: "bun --version",
					Hint:  "Install Bun from https://bun.sh",
				})
				cfg.Tests = append(cfg.Tests, schema.Test{
					Name: "Dependencies",
					Run:  "bun install",
					Expect: schema.Expect{
						ExitCode: intPtr(0),
					},
				})
			} else {
				cfg.Prereqs = append(cfg.Prereqs, schema.Prerequisite{
					Name:  "Node installed",
					Check: "node --version",
					Hint:  "Install Node from https://nodejs.org",
				})
				cfg.Tests = append(cfg.Tests, schema.Test{
					Name: "Dependencies",
					Run:  "npm install",
					Expect: schema.Expect{
						ExitCode: intPtr(0),
					},
				})
			}
			// Add lint test only if lint script exists in package.json.
			if hasPkgScript(dir, "lint") {
				lintCmd := "npm run lint"
				if HasBun(dir) {
					lintCmd = "bun run lint"
				}
				cfg.Tests = append(cfg.Tests, schema.Test{
					Name: "Lint",
					Run:  lintCmd,
					Expect: schema.Expect{
						ExitCode: intPtr(0),
					},
				})
			}

		case Python:
			cfg.Prereqs = append(cfg.Prereqs, schema.Prerequisite{
				Name:  "Python installed",
				Check: "python3 --version",
				Hint:  "Install Python from https://python.org",
			})
			cfg.Tests = append(cfg.Tests, schema.Test{
				Name: "Import check",
				Run:  `python3 -c "import sys; print(sys.version)"`,
				Expect: schema.Expect{
					ExitCode: intPtr(0),
				},
			})

		case Docker:
			cfg.Tests = append(cfg.Tests, schema.Test{
				Name: "Docker build",
				Run:  "docker build -t test .",
				Expect: schema.Expect{
					ExitCode: intPtr(0),
				},
			})

		case Rust:
			cfg.Prereqs = append(cfg.Prereqs, schema.Prerequisite{
				Name:  "Cargo installed",
				Check: "cargo --version",
				Hint:  "Install Rust from https://rustup.rs",
			})
			cfg.Tests = append(cfg.Tests,
				schema.Test{
					Name: "Compiles",
					Run:  "cargo build",
					Expect: schema.Expect{
						ExitCode: intPtr(0),
					},
				},
				schema.Test{
					Name: "Tests",
					Run:  "cargo test",
					Expect: schema.Expect{
						ExitCode: intPtr(0),
					},
				},
			)
		}
	}

	return cfg
}
