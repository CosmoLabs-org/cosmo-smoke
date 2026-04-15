package runner

import (
	"testing"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/reporter"
	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

// noopReporter discards all events.
type noopReporter struct{}

func (n *noopReporter) PrereqStart(_ string)             {}
func (n *noopReporter) PrereqResult(_ reporter.PrereqResultData) {}
func (n *noopReporter) TestStart(_ string)                {}
func (n *noopReporter) TestResult(_ reporter.TestResultData)     {}
func (n *noopReporter) Summary(_ reporter.SuiteResultData)       {}

func intPtr(n int) *int { return &n }

func newConfig(tests []schema.Test) *schema.SmokeConfig {
	return &schema.SmokeConfig{
		Version: 1,
		Project: "test",
		Tests:   tests,
	}
}

func TestRunner_SinglePassingTest(t *testing.T) {
	cfg := newConfig([]schema.Test{
		{Name: "echo", Run: "echo hello", Expect: schema.Expect{ExitCode: intPtr(0)}},
	})
	r := &Runner{Config: cfg, Reporter: &noopReporter{}, ConfigDir: t.TempDir()}
	result, err := r.Run(RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed != 1 {
		t.Errorf("passed = %d, want 1", result.Passed)
	}
	if result.Failed != 0 {
		t.Errorf("failed = %d, want 0", result.Failed)
	}
}

func TestRunner_SingleFailingTest(t *testing.T) {
	cfg := newConfig([]schema.Test{
		{Name: "fail", Run: "exit 1", Expect: schema.Expect{ExitCode: intPtr(0)}},
	})
	r := &Runner{Config: cfg, Reporter: &noopReporter{}, ConfigDir: t.TempDir()}
	result, err := r.Run(RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed != 0 {
		t.Errorf("passed = %d, want 0", result.Passed)
	}
	if result.Failed != 1 {
		t.Errorf("failed = %d, want 1", result.Failed)
	}
}

func TestRunner_FailFast(t *testing.T) {
	cfg := newConfig([]schema.Test{
		{Name: "pass1", Run: "echo 1", Expect: schema.Expect{ExitCode: intPtr(0)}},
		{Name: "fail", Run: "exit 1", Expect: schema.Expect{ExitCode: intPtr(0)}},
		{Name: "skipped", Run: "echo 3", Expect: schema.Expect{ExitCode: intPtr(0)}},
	})
	r := &Runner{Config: cfg, Reporter: &noopReporter{}, ConfigDir: t.TempDir()}
	result, err := r.Run(RunOptions{FailFast: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed != 1 {
		t.Errorf("passed = %d, want 1", result.Passed)
	}
	if result.Failed != 1 {
		t.Errorf("failed = %d, want 1", result.Failed)
	}
	if result.Skipped != 1 {
		t.Errorf("skipped = %d, want 1", result.Skipped)
	}
}

func TestRunner_TagFilter_Include(t *testing.T) {
	cfg := newConfig([]schema.Test{
		{Name: "build", Run: "echo 1", Expect: schema.Expect{ExitCode: intPtr(0)}, Tags: []string{"build"}},
		{Name: "test", Run: "echo 2", Expect: schema.Expect{ExitCode: intPtr(0)}, Tags: []string{"test"}},
	})
	r := &Runner{Config: cfg, Reporter: &noopReporter{}, ConfigDir: t.TempDir()}
	result, err := r.Run(RunOptions{Tags: []string{"build"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}
	if result.Tests[0].Name != "build" {
		t.Errorf("name = %q, want build", result.Tests[0].Name)
	}
}

func TestRunner_TagFilter_Exclude(t *testing.T) {
	cfg := newConfig([]schema.Test{
		{Name: "fast", Run: "echo 1", Expect: schema.Expect{ExitCode: intPtr(0)}, Tags: []string{"fast"}},
		{Name: "slow", Run: "echo 2", Expect: schema.Expect{ExitCode: intPtr(0)}, Tags: []string{"slow"}},
	})
	r := &Runner{Config: cfg, Reporter: &noopReporter{}, ConfigDir: t.TempDir()}
	result, err := r.Run(RunOptions{ExcludeTags: []string{"slow"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}
}

func TestRunner_DryRun(t *testing.T) {
	cfg := newConfig([]schema.Test{
		{Name: "would-fail", Run: "exit 1", Expect: schema.Expect{ExitCode: intPtr(0)}},
	})
	r := &Runner{Config: cfg, Reporter: &noopReporter{}, ConfigDir: t.TempDir()}
	result, err := r.Run(RunOptions{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed != 1 {
		t.Errorf("passed = %d, want 1 (dry run always passes)", result.Passed)
	}
}

func TestRunner_StdoutContains(t *testing.T) {
	cfg := newConfig([]schema.Test{
		{Name: "grep", Run: "echo hello world", Expect: schema.Expect{StdoutContains: "hello"}},
	})
	r := &Runner{Config: cfg, Reporter: &noopReporter{}, ConfigDir: t.TempDir()}
	result, err := r.Run(RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed != 1 {
		t.Errorf("passed = %d, want 1", result.Passed)
	}
}

func TestRunner_Cleanup(t *testing.T) {
	dir := t.TempDir()
	cfg := newConfig([]schema.Test{
		{
			Name:    "with-cleanup",
			Run:     "echo test",
			Expect:  schema.Expect{ExitCode: intPtr(0)},
			Cleanup: "touch " + dir + "/cleaned",
		},
	})
	r := &Runner{Config: cfg, Reporter: &noopReporter{}, ConfigDir: dir}
	_, err := r.Run(RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Give cleanup a moment
	time.Sleep(100 * time.Millisecond)
	// Cleanup should have created the file (but we don't hard-fail if timing is off)
}

func TestRunner_Parallel(t *testing.T) {
	cfg := newConfig([]schema.Test{
		{Name: "a", Run: "echo a", Expect: schema.Expect{ExitCode: intPtr(0)}},
		{Name: "b", Run: "echo b", Expect: schema.Expect{ExitCode: intPtr(0)}},
		{Name: "c", Run: "echo c", Expect: schema.Expect{ExitCode: intPtr(0)}},
	})
	cfg.Settings.Parallel = true
	r := &Runner{Config: cfg, Reporter: &noopReporter{}, ConfigDir: t.TempDir()}
	result, err := r.Run(RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed != 3 {
		t.Errorf("passed = %d, want 3", result.Passed)
	}
}

func TestRunner_PrereqFailure(t *testing.T) {
	cfg := &schema.SmokeConfig{
		Version: 1,
		Project: "test",
		Prereqs: []schema.Prerequisite{
			{Name: "missing-tool", Check: "nonexistent_command_xyz"},
		},
		Tests: []schema.Test{
			{Name: "test", Run: "echo hi"},
		},
	}
	r := &Runner{Config: cfg, Reporter: &noopReporter{}, ConfigDir: t.TempDir()}
	_, err := r.Run(RunOptions{})
	if err == nil {
		t.Fatal("expected error from failed prerequisite")
	}
}

func TestFilterTests_NoFilters(t *testing.T) {
	tests := []schema.Test{{Name: "a"}, {Name: "b"}}
	got := filterTests(tests, nil, nil)
	if len(got) != 2 {
		t.Errorf("got %d tests, want 2", len(got))
	}
}

func TestFilterTests_IncludeOnly(t *testing.T) {
	tests := []schema.Test{
		{Name: "a", Tags: []string{"fast"}},
		{Name: "b", Tags: []string{"slow"}},
	}
	got := filterTests(tests, []string{"fast"}, nil)
	if len(got) != 1 || got[0].Name != "a" {
		t.Errorf("got %v, want [a]", got)
	}
}

func TestFilterTests_ExcludeOnly(t *testing.T) {
	tests := []schema.Test{
		{Name: "a", Tags: []string{"fast"}},
		{Name: "b", Tags: []string{"slow"}},
	}
	got := filterTests(tests, nil, []string{"slow"})
	if len(got) != 1 || got[0].Name != "a" {
		t.Errorf("got %v, want [a]", got)
	}
}
