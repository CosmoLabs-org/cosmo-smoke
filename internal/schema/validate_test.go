package schema

import (
	"strings"
	"testing"
)

func TestValidate_ValidConfig(t *testing.T) {
	exitCode := 0
	cfg := &SmokeConfig{
		Version: 1,
		Project: "myapp",
		Tests: []Test{
			{Name: "test1", Run: "echo hi", Expect: Expect{ExitCode: &exitCode}},
		},
	}
	if err := Validate(cfg); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_MissingProject(t *testing.T) {
	exitCode := 0
	cfg := &SmokeConfig{
		Version: 1,
		Tests: []Test{
			{Name: "test1", Run: "echo hi", Expect: Expect{ExitCode: &exitCode}},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "project name is required") {
		t.Errorf("error = %q, want mention of project", err.Error())
	}
}

func TestValidate_InvalidVersion(t *testing.T) {
	exitCode := 0
	cfg := &SmokeConfig{
		Version: 2,
		Project: "myapp",
		Tests: []Test{
			{Name: "test1", Run: "echo hi", Expect: Expect{ExitCode: &exitCode}},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unsupported version") {
		t.Errorf("error = %q, want mention of version", err.Error())
	}
}

func TestValidate_NoTests(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "myapp",
		Tests:   []Test{},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "at least one test") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestValidate_TestMissingName(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "myapp",
		Tests: []Test{
			{Run: "echo hi"},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestValidate_TestMissingRun(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "myapp",
		Tests: []Test{
			{Name: "test1"},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "run command is required") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 0,
		Tests:   []Test{},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	// Should have: bad version, missing project, no tests
	if len(ve.Errors) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(ve.Errors), ve.Errors)
	}
}

func TestValidate_RetryCountZero(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "myapp",
		Tests: []Test{
			{
				Name: "test1", Run: "echo hi",
				Retry: &RetryPolicy{Count: 0, Backoff: Duration{Duration: 1e9}},
			},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for retry.count = 0")
	}
	if !strings.Contains(err.Error(), "retry.count must be >= 1") {
		t.Errorf("error = %q, want mention of retry.count", err.Error())
	}
}

func TestValidate_RetryBackoffZero(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "myapp",
		Tests: []Test{
			{
				Name: "test1", Run: "echo hi",
				Retry: &RetryPolicy{Count: 3, Backoff: Duration{Duration: 0}},
			},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for retry.backoff = 0")
	}
	if !strings.Contains(err.Error(), "retry.backoff must be > 0") {
		t.Errorf("error = %q, want mention of retry.backoff", err.Error())
	}
}

func TestValidate_RetryValid(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "myapp",
		Tests: []Test{
			{
				Name: "test1", Run: "echo hi",
				Retry: &RetryPolicy{Count: 3, Backoff: Duration{Duration: 1e9}},
			},
		},
	}
	if err := Validate(cfg); err != nil {
		t.Errorf("unexpected error for valid retry block: %v", err)
	}
}

func TestValidate_DockerContainerRunning_MissingName(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "test",
		Tests: []Test{
			{Name: "t1", Run: "true", Expect: Expect{DockerContainer: &DockerContainerCheck{Name: ""}}},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Error("expected validation error for empty docker_container_running.name")
	}
}

func TestValidate_DockerImageExists_MissingImage(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "test",
		Tests: []Test{
			{Name: "t1", Run: "true", Expect: Expect{DockerImage: &DockerImageCheck{Image: ""}}},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Error("expected validation error for empty docker_image_exists.image")
	}
}

func TestValidate_OTelTraceRequiresJaegerURL(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "test",
		Tests: []Test{{
			Name: "otel",
			Expect: Expect{
				OTelTrace: &OTelTraceCheck{MinSpans: 1},
			},
		}},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for otel_trace without jaeger_url")
	}
	if !strings.Contains(err.Error(), "otel_trace.jaeger_url") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_OTelEnabledRequiresJaegerURL(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "test",
		OTel:    OTelConfig{Enabled: true},
		Tests:   []Test{{Name: "t", Run: "true"}},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for otel enabled without jaeger_url")
	}
	if !strings.Contains(err.Error(), "otel.jaeger_url") {
		t.Errorf("unexpected error: %v", err)
	}
}
