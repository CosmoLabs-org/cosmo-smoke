//go:build ignore
package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestTAP_PassingTest(t *testing.T) {
	var buf bytes.Buffer
	r := NewTAP(&buf)

	r.TestStart("Compiles")
	r.TestResult(TestResultData{
		Name:   "Compiles",
		Passed: true,
		Assertions: []AssertionDetail{
			{Type: "exit_code", Expected: "0", Actual: "0", Passed: true},
		},
		Duration: 100 * time.Millisecond,
	})

	r.Summary(SuiteResultData{Total: 1, Passed: 1})

	got := buf.String()
	if !strings.Contains(got, "TAP version 14") {
		t.Error("missing TAP version header")
	}
	if !strings.Contains(got, "1..1") {
		t.Error("missing test plan line")
	}
	if !strings.Contains(got, "ok 1 - Compiles") {
		t.Errorf("expected passing test line, got:\n%s", got)
	}
}

func TestTAP_FailingTest(t *testing.T) {
	var buf bytes.Buffer
	r := NewTAP(&buf)

	r.TestStart("Fails")
	r.TestResult(TestResultData{
		Name:   "Fails",
		Passed: false,
		Assertions: []AssertionDetail{
			{Type: "exit_code", Expected: "0", Actual: "1", Passed: false},
		},
		Duration: 50 * time.Millisecond,
	})

	r.Summary(SuiteResultData{Total: 1, Failed: 1})

	got := buf.String()
	if !strings.Contains(got, "not ok 1 - Fails") {
		t.Errorf("expected failing test line, got:\n%s", got)
	}
	if !strings.Contains(got, "# exit_code: expected 0, got 1") {
		t.Errorf("expected diagnostic line, got:\n%s", got)
	}
}

func TestTAP_SkippedTest(t *testing.T) {
	var buf bytes.Buffer
	r := NewTAP(&buf)

	r.TestStart("Skip me")
	r.TestResult(TestResultData{
		Name:    "Skip me",
		Skipped: true,
	})

	r.Summary(SuiteResultData{Total: 1, Skipped: 1})

	got := buf.String()
	if !strings.Contains(got, "ok 1 - Skip me # SKIP") {
		t.Errorf("expected skip line, got:\n%s", got)
	}
}

func TestTAP_MultipleTests(t *testing.T) {
	var buf bytes.Buffer
	r := NewTAP(&buf)

	r.TestStart("Pass")
	r.TestResult(TestResultData{Name: "Pass", Passed: true})
	r.TestStart("Fail")
	r.TestResult(TestResultData{
		Name:   "Fail",
		Passed: false,
		Assertions: []AssertionDetail{
			{Type: "stdout_contains", Expected: "hello", Actual: "", Passed: false},
		},
	})
	r.TestStart("Skip")
	r.TestResult(TestResultData{Name: "Skip", Skipped: true})

	r.Summary(SuiteResultData{Total: 3, Passed: 1, Failed: 1, Skipped: 1})

	got := buf.String()
	if !strings.Contains(got, "1..3") {
		t.Error("expected plan 1..3")
	}
	if !strings.Contains(got, "ok 1 - Pass") {
		t.Error("missing ok 1")
	}
	if !strings.Contains(got, "not ok 2 - Fail") {
		t.Error("missing not ok 2")
	}
	if !strings.Contains(got, "ok 3 - Skip # SKIP") {
		t.Error("missing ok 3 SKIP")
	}
}

func TestTAP_EmptyResults(t *testing.T) {
	var buf bytes.Buffer
	r := NewTAP(&buf)
	r.Summary(SuiteResultData{})

	got := buf.String()
	if !strings.Contains(got, "1..0") {
		t.Errorf("expected empty plan, got:\n%s", got)
	}
}

func TestTAP_AllowedFailureTest(t *testing.T) {
	var buf bytes.Buffer
	r := NewTAP(&buf)

	r.TestStart("flaky network")
	r.TestResult(TestResultData{
		Name:           "flaky network",
		Passed:         false,
		AllowedFailure: true,
		Assertions: []AssertionDetail{
			{Type: "exit_code", Expected: "0", Actual: "1", Passed: false},
		},
		Duration: 50 * time.Millisecond,
	})

	r.Summary(SuiteResultData{Total: 1, AllowedFailures: 1})

	got := buf.String()
	if !strings.Contains(got, "not ok 1 - flaky network # TODO allow_failure") {
		t.Errorf("expected allowed-failure TAP line, got:\n%s", got)
	}
	if !strings.Contains(got, "# exit_code: expected 0, got 1") {
		t.Errorf("expected diagnostic line, got:\n%s", got)
	}
}
