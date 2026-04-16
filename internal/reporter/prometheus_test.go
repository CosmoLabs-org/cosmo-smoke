package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// lines is a helper that splits output into non-empty lines.
func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

// containsLine checks whether any line in the output equals the target.
func containsLine(output, target string) bool {
	for _, l := range lines(output) {
		if l == target {
			return true
		}
	}
	return false
}

func TestPrometheus_EmptySuite(t *testing.T) {
	var buf bytes.Buffer
	r := NewPrometheus(&buf)
	r.Summary(SuiteResultData{Total: 0, Passed: 0, Failed: 0, AllowedFailures: 0})

	out := buf.String()

	checks := []string{
		"smoke_test_total 0",
		"smoke_test_passed_total 0",
		"smoke_test_failed_total 0",
		"smoke_test_allowed_failure_total 0",
		"smoke_test_duration_seconds 0",
	}
	for _, want := range checks {
		if !containsLine(out, want) {
			t.Errorf("missing line %q in output:\n%s", want, out)
		}
	}

	// No per-test sections for empty suite.
	if strings.Contains(out, "smoke_test_status{") {
		t.Errorf("unexpected per-test status in empty suite output:\n%s", out)
	}
}

func TestPrometheus_SinglePassingTest(t *testing.T) {
	var buf bytes.Buffer
	r := NewPrometheus(&buf)

	r.TestStart("check-health")
	r.TestResult(TestResultData{
		Name:     "check-health",
		Passed:   true,
		Duration: 123 * time.Millisecond,
	})
	r.Summary(SuiteResultData{Total: 1, Passed: 1, Failed: 0, Duration: 123 * time.Millisecond})

	out := buf.String()

	if !containsLine(out, "smoke_test_total 1") {
		t.Errorf("missing smoke_test_total 1 in:\n%s", out)
	}
	if !containsLine(out, "smoke_test_passed_total 1") {
		t.Errorf("missing smoke_test_passed_total 1 in:\n%s", out)
	}
	// status=1 for passing test
	if !containsLine(out, `smoke_test_status{name="check_health",allowed_failure="false"} 1`) {
		t.Errorf("missing per-test status line in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_duration_seconds_per_test{name="check_health"} 0.123`) {
		t.Errorf("missing per-test duration line in:\n%s", out)
	}
}

func TestPrometheus_MixedPassFail(t *testing.T) {
	var buf bytes.Buffer
	r := NewPrometheus(&buf)

	r.TestResult(TestResultData{Name: "pass1", Passed: true, Duration: 10 * time.Millisecond})
	r.TestResult(TestResultData{Name: "pass2", Passed: true, Duration: 20 * time.Millisecond})
	r.TestResult(TestResultData{Name: "fail1", Passed: false, Duration: 5 * time.Millisecond})
	r.Summary(SuiteResultData{Total: 3, Passed: 2, Failed: 1, Duration: 35 * time.Millisecond})

	out := buf.String()

	if !containsLine(out, "smoke_test_total 3") {
		t.Errorf("missing smoke_test_total 3 in:\n%s", out)
	}
	if !containsLine(out, "smoke_test_passed_total 2") {
		t.Errorf("missing smoke_test_passed_total 2 in:\n%s", out)
	}
	if !containsLine(out, "smoke_test_failed_total 1") {
		t.Errorf("missing smoke_test_failed_total 1 in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_status{name="fail1",allowed_failure="false"} 0`) {
		t.Errorf("missing failing test status=0 in:\n%s", out)
	}
}

func TestPrometheus_NameSanitization(t *testing.T) {
	var buf bytes.Buffer
	r := NewPrometheus(&buf)

	r.TestResult(TestResultData{
		Name:     "my-test.1/foo",
		Passed:   true,
		Duration: 50 * time.Millisecond,
	})
	r.Summary(SuiteResultData{Total: 1, Passed: 1})

	out := buf.String()

	// Special chars replaced with underscores: "my-test.1/foo" → "my_test_1_foo"
	if !containsLine(out, `smoke_test_status{name="my_test_1_foo",allowed_failure="false"} 1`) {
		t.Errorf("name not sanitized correctly in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_duration_seconds_per_test{name="my_test_1_foo"} 0.05`) {
		t.Errorf("sanitized duration line missing in:\n%s", out)
	}
}

func TestPrometheus_AllowedFailure(t *testing.T) {
	var buf bytes.Buffer
	r := NewPrometheus(&buf)

	r.TestResult(TestResultData{
		Name:           "optional-check",
		Passed:         false,
		AllowedFailure: true,
		Duration:       15 * time.Millisecond,
	})
	r.Summary(SuiteResultData{
		Total:           1,
		Passed:          0,
		Failed:          0, // allowed failure not counted in Failed
		AllowedFailures: 1,
		Duration:        15 * time.Millisecond,
	})

	out := buf.String()

	// allowed_failure="true", status=0
	if !containsLine(out, `smoke_test_status{name="optional_check",allowed_failure="true"} 0`) {
		t.Errorf("allowed failure test not rendered correctly in:\n%s", out)
	}
	// Not counted in failed_total
	if !containsLine(out, "smoke_test_failed_total 0") {
		t.Errorf("allowed failure should not contribute to failed_total in:\n%s", out)
	}
	if !containsLine(out, "smoke_test_allowed_failure_total 1") {
		t.Errorf("missing allowed_failure_total 1 in:\n%s", out)
	}
}
