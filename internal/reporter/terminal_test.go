package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestTerminal_PrereqPass(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminal(&buf)
	r.PrereqStart("Go installed")
	r.PrereqResult(PrereqResultData{Name: "Go installed", Passed: true, Output: "go1.26.2"})
	out := buf.String()
	if !strings.Contains(out, "Go installed") {
		t.Errorf("output missing prereq name: %q", out)
	}
	if !strings.Contains(out, "go1.26.2") {
		t.Errorf("output missing prereq output: %q", out)
	}
}

func TestTerminal_PrereqFail(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminal(&buf)
	r.PrereqStart("Docker")
	r.PrereqResult(PrereqResultData{Name: "Docker", Passed: false, Hint: "Install Docker"})
	out := buf.String()
	if !strings.Contains(out, "Docker") {
		t.Errorf("output missing prereq name: %q", out)
	}
	if !strings.Contains(out, "Install Docker") {
		t.Errorf("output missing hint: %q", out)
	}
}

func TestTerminal_TestPass(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminal(&buf)
	r.TestStart("Compiles")
	r.TestResult(TestResultData{
		Name:     "Compiles",
		Passed:   true,
		Duration: 150 * time.Millisecond,
	})
	out := buf.String()
	if !strings.Contains(out, "Compiles") {
		t.Errorf("output missing test name: %q", out)
	}
	if !strings.Contains(out, "150ms") {
		t.Errorf("output missing duration: %q", out)
	}
}

func TestTerminal_TestFail(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminal(&buf)
	r.TestStart("Exit check")
	r.TestResult(TestResultData{
		Name:   "Exit check",
		Passed: false,
		Assertions: []AssertionDetail{
			{Type: "exit_code", Expected: "0", Actual: "1", Passed: false},
		},
		Duration: 50 * time.Millisecond,
	})
	out := buf.String()
	if !strings.Contains(out, "exit_code") {
		t.Errorf("output missing assertion type: %q", out)
	}
}

func TestTerminal_TestSkipped(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminal(&buf)
	r.TestStart("Skipped test")
	r.TestResult(TestResultData{Name: "Skipped test", Skipped: true, Duration: 0})
	out := buf.String()
	if !strings.Contains(out, "Skipped test") {
		t.Errorf("output missing test name: %q", out)
	}
}

func TestTerminal_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminal(&buf)
	r.Summary(SuiteResultData{
		Total:    5,
		Passed:   3,
		Failed:   1,
		Skipped:  1,
		Duration: 2 * time.Second,
	})
	out := buf.String()
	if !strings.Contains(out, "5 tests") {
		t.Errorf("output missing total: %q", out)
	}
	if !strings.Contains(out, "3 passed") {
		t.Errorf("output missing passed: %q", out)
	}
	if !strings.Contains(out, "1 failed") {
		t.Errorf("output missing failed: %q", out)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Microsecond, "(500µs)"},
		{150 * time.Millisecond, "(150ms)"},
		{2500 * time.Millisecond, "(2.5s)"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
