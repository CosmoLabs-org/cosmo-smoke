package detector

import (
	"os"
	"path/filepath"
	"testing"
)

// touch creates an empty file at dir/name.
func touch(t *testing.T, dir, name string) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("touch %s: %v", name, err)
	}
	f.Close()
}

func TestDetect_Empty(t *testing.T) {
	dir := t.TempDir()
	types := Detect(dir)
	if len(types) != 0 {
		t.Errorf("empty dir: expected no types, got %v", types)
	}
}

func TestDetect_Go(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "go.mod")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Go {
		t.Errorf("go project: expected [go], got %v", types)
	}
}

func TestDetect_NodeBun(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "bun.lock")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Node {
		t.Errorf("node/bun project: expected [node], got %v", types)
	}
	if !HasBun(dir) {
		t.Error("HasBun: expected true with bun.lock present")
	}
}

func TestDetect_NodeNpm(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "package-lock.json")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Node {
		t.Errorf("node/npm project: expected [node], got %v", types)
	}
	if HasBun(dir) {
		t.Error("HasBun: expected false without bun.lock")
	}
}

func TestDetect_PythonPyproject(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "pyproject.toml")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Python {
		t.Errorf("python/pyproject: expected [python], got %v", types)
	}
}

func TestDetect_PythonRequirements(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "requirements.txt")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Python {
		t.Errorf("python/requirements: expected [python], got %v", types)
	}
}

func TestDetect_PythonSetup(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "setup.py")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Python {
		t.Errorf("python/setup.py: expected [python], got %v", types)
	}
}

func TestDetect_Docker(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "Dockerfile")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Docker {
		t.Errorf("docker/Dockerfile: expected [docker], got %v", types)
	}
}

func TestDetect_DockerCompose(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "docker-compose.yml")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Docker {
		t.Errorf("docker/compose: expected [docker], got %v", types)
	}
}

func TestDetect_Rust(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "Cargo.toml")
	types := Detect(dir)
	if len(types) != 1 || types[0] != Rust {
		t.Errorf("rust: expected [rust], got %v", types)
	}
}

func TestDetect_MultipleTypes(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "go.mod")
	touch(t, dir, "Dockerfile")
	types := Detect(dir)
	if len(types) != 2 {
		t.Fatalf("go+docker: expected 2 types, got %d: %v", len(types), types)
	}
	typeSet := map[ProjectType]bool{}
	for _, pt := range types {
		typeSet[pt] = true
	}
	if !typeSet[Go] || !typeSet[Docker] {
		t.Errorf("go+docker: expected Go and Docker, got %v", types)
	}
}

// --- GenerateConfig tests ---

func TestGenerateConfig_Go(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "go.mod")
	cfg := GenerateConfig(dir, []ProjectType{Go})

	if cfg.Version != 1 {
		t.Errorf("version: want 1, got %d", cfg.Version)
	}
	if cfg.Project != filepath.Base(dir) {
		t.Errorf("project name: want %q, got %q", filepath.Base(dir), cfg.Project)
	}
	if !cfg.Settings.FailFast {
		t.Error("fail_fast should be true")
	}
	if cfg.Settings.Timeout.Seconds() != 30 {
		t.Errorf("timeout: want 30s, got %v", cfg.Settings.Timeout)
	}
	if len(cfg.Prereqs) != 1 || cfg.Prereqs[0].Name != "Go installed" {
		t.Errorf("prereqs: want [Go installed], got %v", cfg.Prereqs)
	}
	if len(cfg.Tests) != 2 {
		t.Fatalf("tests: want 2, got %d", len(cfg.Tests))
	}
	if cfg.Tests[0].Name != "Compiles" {
		t.Errorf("test[0]: want Compiles, got %q", cfg.Tests[0].Name)
	}
	if cfg.Tests[1].Name != "Tests pass" {
		t.Errorf("test[1]: want 'Tests pass', got %q", cfg.Tests[1].Name)
	}
	for _, test := range cfg.Tests {
		if test.Expect.ExitCode == nil || *test.Expect.ExitCode != 0 {
			t.Errorf("test %q: want exit_code 0", test.Name)
		}
	}
}

func TestGenerateConfig_NodeBun(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "bun.lock")
	cfg := GenerateConfig(dir, []ProjectType{Node})

	if len(cfg.Prereqs) != 1 || cfg.Prereqs[0].Name != "Bun installed" {
		t.Errorf("prereqs: want [Bun installed], got %v", cfg.Prereqs)
	}
	if len(cfg.Tests) < 1 || cfg.Tests[0].Run != "bun install" {
		t.Errorf("test[0]: want 'bun install', got %q", cfg.Tests[0].Run)
	}
}

func TestGenerateConfig_NodeNpm(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	cfg := GenerateConfig(dir, []ProjectType{Node})

	if len(cfg.Prereqs) != 1 || cfg.Prereqs[0].Name != "Node installed" {
		t.Errorf("prereqs: want [Node installed], got %v", cfg.Prereqs)
	}
	if len(cfg.Tests) < 1 || cfg.Tests[0].Run != "npm install" {
		t.Errorf("test[0]: want 'npm install', got %q", cfg.Tests[0].Run)
	}
}

func TestGenerateConfig_NodeWithLint(t *testing.T) {
	dir := t.TempDir()
	// Write a package.json with a lint script.
	pkgJSON := []byte(`{"scripts":{"lint":"eslint ."}}`)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), pkgJSON, 0644); err != nil {
		t.Fatal(err)
	}
	cfg := GenerateConfig(dir, []ProjectType{Node})

	// Expect: Dependencies + Lint
	if len(cfg.Tests) != 2 {
		t.Fatalf("tests: want 2 (dependencies + lint), got %d: %v", len(cfg.Tests), cfg.Tests)
	}
	if cfg.Tests[1].Name != "Lint" {
		t.Errorf("test[1]: want Lint, got %q", cfg.Tests[1].Name)
	}
}

func TestGenerateConfig_NodeNoLint(t *testing.T) {
	dir := t.TempDir()
	pkgJSON := []byte(`{"scripts":{"build":"tsc"}}`)
	if err := os.WriteFile(filepath.Join(dir, "package.json"), pkgJSON, 0644); err != nil {
		t.Fatal(err)
	}
	cfg := GenerateConfig(dir, []ProjectType{Node})

	// Only Dependencies, no Lint
	if len(cfg.Tests) != 1 {
		t.Errorf("tests: want 1 (no lint script), got %d: %v", len(cfg.Tests), cfg.Tests)
	}
}

func TestGenerateConfig_Python(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "pyproject.toml")
	cfg := GenerateConfig(dir, []ProjectType{Python})

	if len(cfg.Prereqs) != 1 || cfg.Prereqs[0].Name != "Python installed" {
		t.Errorf("prereqs: want [Python installed], got %v", cfg.Prereqs)
	}
	if len(cfg.Tests) != 1 || cfg.Tests[0].Name != "Import check" {
		t.Errorf("tests: want [Import check], got %v", cfg.Tests)
	}
}

func TestGenerateConfig_Docker(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "Dockerfile")
	cfg := GenerateConfig(dir, []ProjectType{Docker})

	if len(cfg.Prereqs) != 0 {
		t.Errorf("docker: expected no prereqs, got %v", cfg.Prereqs)
	}
	if len(cfg.Tests) != 1 || cfg.Tests[0].Name != "Docker build" {
		t.Errorf("tests: want [Docker build], got %v", cfg.Tests)
	}
}

func TestGenerateConfig_Rust(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "Cargo.toml")
	cfg := GenerateConfig(dir, []ProjectType{Rust})

	if len(cfg.Prereqs) != 1 || cfg.Prereqs[0].Name != "Cargo installed" {
		t.Errorf("prereqs: want [Cargo installed], got %v", cfg.Prereqs)
	}
	if len(cfg.Tests) != 2 {
		t.Fatalf("tests: want 2, got %d", len(cfg.Tests))
	}
	if cfg.Tests[0].Name != "Compiles" || cfg.Tests[1].Name != "Tests" {
		t.Errorf("test names: got %q, %q", cfg.Tests[0].Name, cfg.Tests[1].Name)
	}
}

func TestGenerateConfig_MultipleTypes(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "go.mod")
	touch(t, dir, "Dockerfile")
	cfg := GenerateConfig(dir, []ProjectType{Go, Docker})

	// Go: 1 prereq, 2 tests; Docker: 0 prereqs, 1 test
	if len(cfg.Prereqs) != 1 {
		t.Errorf("prereqs: want 1 (Go only), got %d", len(cfg.Prereqs))
	}
	if len(cfg.Tests) != 3 {
		t.Errorf("tests: want 3 (2 Go + 1 Docker), got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_Empty(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, nil)

	if cfg.Version != 1 {
		t.Errorf("version: want 1, got %d", cfg.Version)
	}
	if len(cfg.Prereqs) != 0 {
		t.Errorf("prereqs: want 0, got %d", len(cfg.Prereqs))
	}
	if len(cfg.Tests) != 0 {
		t.Errorf("tests: want 0, got %d", len(cfg.Tests))
	}
}
