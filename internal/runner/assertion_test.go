package runner

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// CheckExitCode
// ---------------------------------------------------------------------------

func TestCheckExitCode_Pass(t *testing.T) {
	r := CheckExitCode(0, 0)
	if !r.Passed {
		t.Errorf("expected pass, got fail: actual=%s expected=%s", r.Actual, r.Expected)
	}
	if r.Type != "exit_code" {
		t.Errorf("expected type 'exit_code', got %q", r.Type)
	}
}

func TestCheckExitCode_Fail(t *testing.T) {
	r := CheckExitCode(1, 0)
	if r.Passed {
		t.Errorf("expected fail, got pass")
	}
	if r.Actual != "1" || r.Expected != "0" {
		t.Errorf("unexpected values: actual=%s expected=%s", r.Actual, r.Expected)
	}
}

func TestCheckExitCode_NonZeroExpected(t *testing.T) {
	r := CheckExitCode(2, 2)
	if !r.Passed {
		t.Errorf("expected pass for exit code 2 == 2")
	}
}

func TestCheckExitCode_LargeCode(t *testing.T) {
	r := CheckExitCode(127, 127)
	if !r.Passed {
		t.Errorf("expected pass for exit code 127 == 127")
	}
}

// ---------------------------------------------------------------------------
// CheckStdoutContains
// ---------------------------------------------------------------------------

func TestCheckStdoutContains_Pass(t *testing.T) {
	r := CheckStdoutContains("hello world", "world")
	if !r.Passed {
		t.Errorf("expected pass")
	}
	if r.Type != "stdout_contains" {
		t.Errorf("expected type 'stdout_contains', got %q", r.Type)
	}
}

func TestCheckStdoutContains_Fail(t *testing.T) {
	r := CheckStdoutContains("hello world", "missing")
	if r.Passed {
		t.Errorf("expected fail")
	}
}

func TestCheckStdoutContains_EmptyStdout(t *testing.T) {
	r := CheckStdoutContains("", "anything")
	if r.Passed {
		t.Errorf("expected fail for empty stdout")
	}
}

func TestCheckStdoutContains_EmptySubstr(t *testing.T) {
	// Empty substring is always contained in any string (Go strings.Contains behavior)
	r := CheckStdoutContains("hello", "")
	if !r.Passed {
		t.Errorf("expected pass: empty substring is contained in any string")
	}
}

func TestCheckStdoutContains_BothEmpty(t *testing.T) {
	r := CheckStdoutContains("", "")
	if !r.Passed {
		t.Errorf("expected pass: empty substring in empty string")
	}
}

func TestCheckStdoutContains_Multiline(t *testing.T) {
	stdout := "line1\nline2\nline3"
	r := CheckStdoutContains(stdout, "line2")
	if !r.Passed {
		t.Errorf("expected pass for multiline stdout")
	}
}

func TestCheckStdoutContains_ExactMatch(t *testing.T) {
	r := CheckStdoutContains("exact", "exact")
	if !r.Passed {
		t.Errorf("expected pass for exact match")
	}
}

// ---------------------------------------------------------------------------
// CheckStdoutMatches
// ---------------------------------------------------------------------------

func TestCheckStdoutMatches_Pass(t *testing.T) {
	r := CheckStdoutMatches("version 1.2.3", `version \d+\.\d+\.\d+`)
	if !r.Passed {
		t.Errorf("expected pass")
	}
	if r.Type != "stdout_matches" {
		t.Errorf("expected type 'stdout_matches', got %q", r.Type)
	}
}

func TestCheckStdoutMatches_Fail(t *testing.T) {
	r := CheckStdoutMatches("hello world", `^\d+$`)
	if r.Passed {
		t.Errorf("expected fail")
	}
}

func TestCheckStdoutMatches_EmptyStdout(t *testing.T) {
	r := CheckStdoutMatches("", `\w+`)
	if r.Passed {
		t.Errorf("expected fail for empty stdout against word pattern")
	}
}

func TestCheckStdoutMatches_EmptyStdoutMatchesEmptyPattern(t *testing.T) {
	r := CheckStdoutMatches("", "")
	if !r.Passed {
		t.Errorf("expected pass: empty pattern matches empty string")
	}
}

func TestCheckStdoutMatches_InvalidRegex(t *testing.T) {
	r := CheckStdoutMatches("some output", `[invalid`)
	if r.Passed {
		t.Errorf("expected fail for invalid regex")
	}
	// Actual should describe the error
	if r.Actual == "some output" {
		t.Errorf("expected Actual to contain error description, got stdout instead")
	}
}

func TestCheckStdoutMatches_AnchoredPattern(t *testing.T) {
	r := CheckStdoutMatches("ok", `^ok$`)
	if !r.Passed {
		t.Errorf("expected pass for anchored exact match")
	}
}

func TestCheckStdoutMatches_AnchoredPatternFail(t *testing.T) {
	r := CheckStdoutMatches("ok extra", `^ok$`)
	if r.Passed {
		t.Errorf("expected fail: anchored pattern should not match with extra content on same line")
	}
}

func TestCheckStdoutMatches_CaseInsensitiveNotDefault(t *testing.T) {
	// By default regex is case-sensitive
	r := CheckStdoutMatches("Hello", `hello`)
	if r.Passed {
		t.Errorf("expected fail: regex is case-sensitive by default")
	}
}

func TestCheckStdoutMatches_CaseFlagOverride(t *testing.T) {
	r := CheckStdoutMatches("Hello", `(?i)hello`)
	if !r.Passed {
		t.Errorf("expected pass with case-insensitive flag")
	}
}

// ---------------------------------------------------------------------------
// CheckStderrContains
// ---------------------------------------------------------------------------

func TestCheckStderrContains_Pass(t *testing.T) {
	r := CheckStderrContains("error: connection refused", "connection refused")
	if !r.Passed {
		t.Errorf("expected pass")
	}
	if r.Type != "stderr_contains" {
		t.Errorf("expected type 'stderr_contains', got %q", r.Type)
	}
}

func TestCheckStderrContains_Fail(t *testing.T) {
	r := CheckStderrContains("error: timeout", "connection refused")
	if r.Passed {
		t.Errorf("expected fail")
	}
}

func TestCheckStderrContains_EmptyStderr(t *testing.T) {
	r := CheckStderrContains("", "error")
	if r.Passed {
		t.Errorf("expected fail for empty stderr")
	}
}

func TestCheckStderrContains_EmptySubstr(t *testing.T) {
	r := CheckStderrContains("some error", "")
	if !r.Passed {
		t.Errorf("expected pass: empty substring is always contained")
	}
}

// ---------------------------------------------------------------------------
// CheckFileExists
// ---------------------------------------------------------------------------

func TestCheckFileExists_AbsolutePathExists(t *testing.T) {
	tmp, err := os.CreateTemp("", "smoke-assert-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	r := CheckFileExists(tmp.Name(), "/some/config/dir")
	if !r.Passed {
		t.Errorf("expected pass for existing absolute path %s", tmp.Name())
	}
	if r.Type != "file_exists" {
		t.Errorf("expected type 'file_exists', got %q", r.Type)
	}
}

func TestCheckFileExists_AbsolutePathMissing(t *testing.T) {
	r := CheckFileExists("/nonexistent/path/to/file.txt", "/some/config/dir")
	if r.Passed {
		t.Errorf("expected fail for nonexistent absolute path")
	}
}

func TestCheckFileExists_RelativePathExists(t *testing.T) {
	dir := t.TempDir()
	filename := "output.txt"
	fullPath := filepath.Join(dir, filename)
	if err := os.WriteFile(fullPath, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	r := CheckFileExists(filename, dir)
	if !r.Passed {
		t.Errorf("expected pass for relative path resolved against configDir: %s", fullPath)
	}
	// Expected and Actual should both be the resolved absolute path
	if r.Expected != fullPath {
		t.Errorf("expected resolved path %q, got %q", fullPath, r.Expected)
	}
}

func TestCheckFileExists_RelativePathMissing(t *testing.T) {
	dir := t.TempDir()
	r := CheckFileExists("ghost.txt", dir)
	if r.Passed {
		t.Errorf("expected fail for missing relative path")
	}
}

func TestCheckFileExists_RelativeSubdirExists(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	fname := filepath.Join(subdir, "result.json")
	if err := os.WriteFile(fname, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	r := CheckFileExists("sub/result.json", dir)
	if !r.Passed {
		t.Errorf("expected pass for nested relative path")
	}
}

func TestCheckFileExists_AbsolutePathIgnoresConfigDir(t *testing.T) {
	// When given an absolute path, configDir should be ignored entirely
	tmp, err := os.CreateTemp("", "smoke-abs-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	r := CheckFileExists(tmp.Name(), "/completely/unrelated/dir")
	if !r.Passed {
		t.Errorf("expected pass: absolute path should ignore configDir")
	}
	if r.Expected != tmp.Name() {
		t.Errorf("expected resolved path to equal original absolute path, got %q", r.Expected)
	}
}

func TestCheckFileExists_EmptyConfigDir(t *testing.T) {
	// Relative path with empty configDir resolves to just the filename (relative to cwd)
	// This should not panic
	r := CheckFileExists("nonexistent.txt", "")
	// Result depends on whether file exists in cwd, but it must not panic
	_ = r
}

// ---------------------------------------------------------------------------
// CheckStderrMatches
// ---------------------------------------------------------------------------

func TestCheckStderrMatches_Pass(t *testing.T) {
	r := CheckStderrMatches("error: line 42: undefined", `line \d+`)
	if !r.Passed {
		t.Errorf("expected pass")
	}
	if r.Type != "stderr_matches" {
		t.Errorf("expected type 'stderr_matches', got %q", r.Type)
	}
}

func TestCheckStderrMatches_Fail(t *testing.T) {
	r := CheckStderrMatches("error: unknown", `line \d+`)
	if r.Passed {
		t.Errorf("expected fail")
	}
}

func TestCheckStderrMatches_InvalidRegex(t *testing.T) {
	r := CheckStderrMatches("some error", "[invalid(regex")
	if r.Passed {
		t.Errorf("expected fail for invalid regex")
	}
	if r.Actual == "" {
		t.Errorf("expected error message in Actual")
	}
}

func TestCheckStderrMatches_EmptyStderr(t *testing.T) {
	r := CheckStderrMatches("", "error")
	if r.Passed {
		t.Errorf("expected fail for empty stderr")
	}
}

func TestCheckStderrMatches_EmptyPattern(t *testing.T) {
	r := CheckStderrMatches("some error", "")
	if !r.Passed {
		t.Errorf("expected pass: empty pattern matches everything")
	}
}

// ---------------------------------------------------------------------------
// CheckEnvExists
// ---------------------------------------------------------------------------

func TestCheckEnvExists_Pass(t *testing.T) {
	os.Setenv("SMOKE_TEST_VAR", "value123")
	defer os.Unsetenv("SMOKE_TEST_VAR")

	r := CheckEnvExists("SMOKE_TEST_VAR")
	if !r.Passed {
		t.Errorf("expected pass for set env var")
	}
	if r.Type != "env_exists" {
		t.Errorf("expected type 'env_exists', got %q", r.Type)
	}
	if r.Actual != "value123" {
		t.Errorf("expected Actual to be 'value123', got %q", r.Actual)
	}
}

func TestCheckEnvExists_Fail(t *testing.T) {
	os.Unsetenv("SMOKE_NONEXISTENT_VAR")

	r := CheckEnvExists("SMOKE_NONEXISTENT_VAR")
	if r.Passed {
		t.Errorf("expected fail for unset env var")
	}
	if r.Actual != "" {
		t.Errorf("expected empty Actual for unset var, got %q", r.Actual)
	}
}

func TestCheckEnvExists_EmptyValue(t *testing.T) {
	os.Setenv("SMOKE_EMPTY_VAR", "")
	defer os.Unsetenv("SMOKE_EMPTY_VAR")

	r := CheckEnvExists("SMOKE_EMPTY_VAR")
	if r.Passed {
		t.Errorf("expected fail: empty string should count as not set")
	}
}
