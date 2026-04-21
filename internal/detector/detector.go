package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ProjectType identifies the kind of project detected.
type ProjectType string

const (
	Go         ProjectType = "go"
	Node       ProjectType = "node"
	Python     ProjectType = "python"
	Docker     ProjectType = "docker"
	Rust       ProjectType = "rust"
	ReactNative ProjectType = "react-native"
	Flutter    ProjectType = "flutter"
	IOS        ProjectType = "ios"
	Android    ProjectType = "android"
)

// exists returns true if the given path exists under dir.
func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

// hasGlob returns true if any file matching the glob pattern exists under dir.
func hasGlob(dir, pattern string) bool {
	matches, _ := filepath.Glob(filepath.Join(dir, pattern))
	return len(matches) > 0
}

// hasType checks if a specific type is already in the list.
func hasType(types []ProjectType, want ProjectType) bool {
	for _, t := range types {
		if t == want {
			return true
		}
	}
	return false
}

// hasDepInPackageJSON checks if package.json has a specific dependency.
func hasDepInPackageJSON(dir, dep string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}
	_, ok := pkg.Dependencies[dep]
	if !ok {
		_, ok = pkg.DevDependencies[dep]
	}
	return ok
}

// hasFlutterDep checks if pubspec.yaml has a Flutter SDK dependency.
func hasFlutterDep(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "pubspec.yaml"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "sdk: flutter")
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

	// React Native: app.json + react-native dependency or metro.config.js
	if exists(dir, "app.json") {
		if hasDepInPackageJSON(dir, "react-native") || exists(dir, "metro.config.js") {
			types = append(types, ReactNative)
		}
	}
	// Flutter: pubspec.yaml with flutter dependency
	if exists(dir, "pubspec.yaml") && hasFlutterDep(dir) {
		types = append(types, Flutter)
	}
	// iOS native: xcodeproj/xcworkspace or Podfile (skip if RN/Flutter)
	if !hasType(types, ReactNative) && !hasType(types, Flutter) {
		if hasGlob(dir, "*.xcodeproj") || hasGlob(dir, "*.xcworkspace") || exists(dir, "Podfile") {
			types = append(types, IOS)
		}
	}
	// Android native: build.gradle without Go/Node (skip if RN/Flutter)
	if !hasType(types, ReactNative) && !hasType(types, Flutter) {
		if exists(dir, "build.gradle") || exists(dir, "build.gradle.kts") {
			if !exists(dir, "go.mod") && !exists(dir, "package.json") {
				types = append(types, Android)
			}
		}
	}

	return types
}

// HasBun returns true if the Node project uses bun (bun.lock present).
func HasBun(dir string) bool {
	return exists(dir, "bun.lock")
}
