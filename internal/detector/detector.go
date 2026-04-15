package detector

import (
	"os"
	"path/filepath"
)

// ProjectType identifies the kind of project detected.
type ProjectType string

const (
	Go     ProjectType = "go"
	Node   ProjectType = "node"
	Python ProjectType = "python"
	Docker ProjectType = "docker"
	Rust   ProjectType = "rust"
)

// exists returns true if the given path exists under dir.
func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

// Detect scans dir for project type markers and returns all detected types.
func Detect(dir string) []ProjectType {
	var types []ProjectType

	if exists(dir, "go.mod") {
		types = append(types, Go)
	}
	if exists(dir, "package.json") {
		types = append(types, Node)
	}
	if exists(dir, "pyproject.toml") || exists(dir, "requirements.txt") || exists(dir, "setup.py") {
		types = append(types, Python)
	}
	if exists(dir, "Dockerfile") || exists(dir, "docker-compose.yml") {
		types = append(types, Docker)
	}
	if exists(dir, "Cargo.toml") {
		types = append(types, Rust)
	}

	return types
}

// HasBun returns true if the Node project uses bun (bun.lock present).
func HasBun(dir string) bool {
	return exists(dir, "bun.lock")
}
