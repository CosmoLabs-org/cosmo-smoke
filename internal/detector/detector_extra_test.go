package detector

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExtra_DetectGoMod verifies Go detection from go.mod presence.
func TestExtra_DetectGoMod(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "go.mod")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Go {
		t.Fatalf("expected [go], got %v", types)
	}
}

// TestExtra_DetectNodePackageJSON verifies Node detection from package.json.
func TestExtra_DetectNodePackageJSON(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Node {
		t.Fatalf("expected [node], got %v", types)
	}
}

// TestExtra_DetectDockerDockerfile verifies Docker detection from Dockerfile.
func TestExtra_DetectDockerDockerfile(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "Dockerfile")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Docker {
		t.Fatalf("expected [docker], got %v", types)
	}
}

// TestExtra_DetectPythonRequirementsTxt verifies Python detection from requirements.txt.
func TestExtra_DetectPythonRequirementsTxt(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "requirements.txt")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Python {
		t.Fatalf("expected [python], got %v", types)
	}
}

// TestExtra_DetectPythonPyprojectToml verifies Python detection from pyproject.toml.
func TestExtra_DetectPythonPyprojectToml(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "pyproject.toml")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Python {
		t.Fatalf("expected [python], got %v", types)
	}
}

// TestExtra_DetectRustCargoToml verifies Rust detection from Cargo.toml.
func TestExtra_DetectRustCargoToml(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "Cargo.toml")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Rust {
		t.Fatalf("expected [rust], got %v", types)
	}
}

// TestExtra_DetectUnknownReturnsEmpty verifies that an unknown project
// (no marker files) returns an empty slice from Detect and a default
// (empty tests/prereqs) config from GenerateConfig.
func TestExtra_DetectUnknownReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	types := Detect(dir)
	if len(types) != 0 {
		t.Fatalf("expected no types, got %v", types)
	}

	cfg := GenerateConfig(dir, types)
	if len(cfg.Tests) != 0 {
		t.Errorf("expected 0 tests for unknown project, got %d", len(cfg.Tests))
	}
	if len(cfg.Prereqs) != 0 {
		t.Errorf("expected 0 prereqs for unknown project, got %d", len(cfg.Prereqs))
	}
	if cfg.Project != filepath.Base(dir) {
		t.Errorf("expected project name %q, got %q", filepath.Base(dir), cfg.Project)
	}
}

// TestExtra_DetectDockerCompose verifies Docker detection via docker-compose.yml.
func TestExtra_DetectDockerCompose(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "docker-compose.yml")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Docker {
		t.Fatalf("expected [docker], got %v", types)
	}
}

// TestExtra_DetectPythonSetupPy verifies Python detection from setup.py.
func TestExtra_DetectPythonSetupPy(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "setup.py")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Python {
		t.Fatalf("expected [python], got %v", types)
	}
}

// TestExtra_HasBunTrue verifies HasBun returns true when bun.lock exists.
func TestExtra_HasBunTrue(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "bun.lock")
	if !HasBun(dir) {
		t.Fatal("expected HasBun=true with bun.lock")
	}
}

// TestExtra_HasBunFalse verifies HasBun returns false when bun.lock is absent.
func TestExtra_HasBunFalse(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	if HasBun(dir) {
		t.Fatal("expected HasBun=false without bun.lock")
	}
}

// TestExtra_GenerateConfigGoDefaults checks default config fields for a Go project.
func TestExtra_GenerateConfigGoDefaults(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "go.mod")
	cfg := GenerateConfig(dir, []ProjectType{Go})

	if cfg.Version != 1 {
		t.Errorf("version: want 1, got %d", cfg.Version)
	}
	if !cfg.Settings.FailFast {
		t.Error("fail_fast should default to true")
	}
	if cfg.Settings.Timeout.Seconds() != 30 {
		t.Errorf("timeout: want 30s, got %v", cfg.Settings.Timeout)
	}
}

// TestExtra_GenerateConfigAllTypes verifies a project with all types detected
// produces prereqs and tests from every type.
func TestExtra_GenerateConfigAllTypes(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "go.mod")
	touch(t, dir, "package.json")
	touch(t, dir, "pyproject.toml")
	touch(t, dir, "Dockerfile")
	touch(t, dir, "Cargo.toml")

	allTypes := Detect(dir)
	if len(allTypes) != 5 {
		t.Fatalf("expected 5 types, got %d: %v", len(allTypes), allTypes)
	}

	cfg := GenerateConfig(dir, allTypes)
	// Go(1 prereq) + Node(1 prereq) + Python(1 prereq) + Rust(1 prereq) = 4
	if len(cfg.Prereqs) != 4 {
		t.Errorf("expected 4 prereqs, got %d: %v", len(cfg.Prereqs), cfg.Prereqs)
	}
	// Go(2) + Node(1) + Python(1) + Docker(1) + Rust(2) = 7
	if len(cfg.Tests) != 7 {
		t.Errorf("expected 7 tests, got %d: %v", len(cfg.Tests), cfg.Tests)
	}
}

// TestExtra_NodeWithLintScript verifies GenerateConfig adds a Lint test
// when the package.json contains a lint script.
func TestExtra_NodeWithLintScript(t *testing.T) {
	dir := t.TempDir()
	pkg := []byte(`{"scripts":{"lint":"eslint ."}}`)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), pkg, 0644); err != nil {
		t.Fatal(err)
	}
	cfg := GenerateConfig(dir, []ProjectType{Node})

	if len(cfg.Tests) != 2 {
		t.Fatalf("expected 2 tests (deps + lint), got %d", len(cfg.Tests))
	}
	if cfg.Tests[1].Name != "Lint" {
		t.Errorf("test[1]: want Lint, got %q", cfg.Tests[1].Name)
	}
}
