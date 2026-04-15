package schema

import (
	"testing"
	"time"
)

func TestParse_ValidConfig(t *testing.T) {
	yaml := `
version: 1
project: myapp
description: "Test suite"
settings:
  timeout: 30s
  fail_fast: true
  parallel: false
prerequisites:
  - name: "Go installed"
    check: "go version"
    hint: "Install Go from https://go.dev"
tests:
  - name: "Compiles"
    run: "go build ./..."
    expect:
      exit_code: 0
    tags: [build]
    timeout: 10s
  - name: "Has README"
    run: "echo hi"
    expect:
      file_exists: "README.md"
`
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("version = %d, want 1", cfg.Version)
	}
	if cfg.Project != "myapp" {
		t.Errorf("project = %q, want %q", cfg.Project, "myapp")
	}
	if cfg.Description != "Test suite" {
		t.Errorf("description = %q, want %q", cfg.Description, "Test suite")
	}
	if cfg.Settings.Timeout.Duration != 30*time.Second {
		t.Errorf("timeout = %v, want 30s", cfg.Settings.Timeout.Duration)
	}
	if !cfg.Settings.FailFast {
		t.Error("fail_fast should be true")
	}
	if cfg.Settings.Parallel {
		t.Error("parallel should be false")
	}
	if len(cfg.Prereqs) != 1 {
		t.Fatalf("prereqs count = %d, want 1", len(cfg.Prereqs))
	}
	if cfg.Prereqs[0].Name != "Go installed" {
		t.Errorf("prereq name = %q", cfg.Prereqs[0].Name)
	}
	if len(cfg.Tests) != 2 {
		t.Fatalf("tests count = %d, want 2", len(cfg.Tests))
	}
	if cfg.Tests[0].Name != "Compiles" {
		t.Errorf("test[0].name = %q", cfg.Tests[0].Name)
	}
	if cfg.Tests[0].Expect.ExitCode == nil || *cfg.Tests[0].Expect.ExitCode != 0 {
		t.Errorf("test[0].expect.exit_code should be 0")
	}
	if len(cfg.Tests[0].Tags) != 1 || cfg.Tests[0].Tags[0] != "build" {
		t.Errorf("test[0].tags = %v", cfg.Tests[0].Tags)
	}
	if cfg.Tests[0].Timeout.Duration != 10*time.Second {
		t.Errorf("test[0].timeout = %v, want 10s", cfg.Tests[0].Timeout.Duration)
	}
	if cfg.Tests[1].Expect.FileExists != "README.md" {
		t.Errorf("test[1].expect.file_exists = %q", cfg.Tests[1].Expect.FileExists)
	}
}

func TestParse_MinimalConfig(t *testing.T) {
	yaml := `
version: 1
project: minimal
tests:
  - name: "echo"
    run: "echo hello"
    expect:
      exit_code: 0
`
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Project != "minimal" {
		t.Errorf("project = %q", cfg.Project)
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	_, err := Parse([]byte(`{{{not yaml`))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParse_AllAssertionTypes(t *testing.T) {
	yaml := `
version: 1
project: assertions
tests:
  - name: "all"
    run: "echo test"
    expect:
      exit_code: 0
      stdout_contains: "test"
      stdout_matches: "^te.t$"
      stderr_contains: "warn"
      file_exists: "go.mod"
`
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e := cfg.Tests[0].Expect
	if e.ExitCode == nil || *e.ExitCode != 0 {
		t.Error("exit_code")
	}
	if e.StdoutContains != "test" {
		t.Error("stdout_contains")
	}
	if e.StdoutMatches != "^te.t$" {
		t.Error("stdout_matches")
	}
	if e.StderrContains != "warn" {
		t.Error("stderr_contains")
	}
	if e.FileExists != "go.mod" {
		t.Error("file_exists")
	}
}

func TestDuration_Unmarshal(t *testing.T) {
	yaml := `
version: 1
project: dur
settings:
  timeout: 2m30s
tests:
  - name: "t"
    run: "echo"
    expect:
      exit_code: 0
`
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := 2*time.Minute + 30*time.Second
	if cfg.Settings.Timeout.Duration != want {
		t.Errorf("timeout = %v, want %v", cfg.Settings.Timeout.Duration, want)
	}
}
