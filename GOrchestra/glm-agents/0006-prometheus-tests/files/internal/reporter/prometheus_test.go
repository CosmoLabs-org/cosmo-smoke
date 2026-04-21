//go:build ignore
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

func TestPrometheus_AllFailing(t *testing.T) {
	var buf bytes.Buffer
	r := NewPrometheus(&buf)

	r.TestResult(TestResultData{Name: "fail-a", Passed: false, Duration: 10 * time.Millisecond})
	r.TestResult(TestResultData{Name: "fail-b", Passed: false, Duration: 20 * time.Millisecond})
	r.TestResult(TestResultData{Name: "fail-c", Passed: false, Duration: 30 * time.Millisecond})
	r.Summary(SuiteResultData{Total: 3, Passed: 0, Failed: 3, Duration: 60 * time.Millisecond})

	out := buf.String()

	if !containsLine(out, "smoke_test_passed_total 0") {
		t.Errorf("expected passed_total 0 in:\n%s", out)
	}
	if !containsLine(out, "smoke_test_failed_total 3") {
		t.Errorf("expected failed_total 3 in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_status{name="fail_a",allowed_failure="false"} 0`) {
		t.Errorf("missing fail-a status line in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_status{name="fail_b",allowed_failure="false"} 0`) {
		t.Errorf("missing fail-b status line in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_status{name="fail_c",allowed_failure="false"} 0`) {
		t.Errorf("missing fail-c status line in:\n%s", out)
	}
}

func TestPrometheus_MixedPassFailSkip(t *testing.T) {
	var buf bytes.Buffer
	r := NewPrometheus(&buf)

	r.TestResult(TestResultData{Name: "pass1", Passed: true, Skipped: false, Duration: 10 * time.Millisecond})
	r.TestResult(TestResultData{Name: "fail1", Passed: false, Skipped: false, Duration: 5 * time.Millisecond})
	r.TestResult(TestResultData{Name: "skip1", Passed: false, Skipped: true, Duration: 0})
	r.Summary(SuiteResultData{Total: 3, Passed: 1, Failed: 1, Skipped: 1, Duration: 15 * time.Millisecond})

	out := buf.String()

	if !containsLine(out, "smoke_test_total 3") {
		t.Errorf("missing total 3 in:\n%s", out)
	}
	if !containsLine(out, "smoke_test_passed_total 1") {
		t.Errorf("missing passed_total 1 in:\n%s", out)
	}
	if !containsLine(out, "smoke_test_failed_total 1") {
		t.Errorf("missing failed_total 1 in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_status{name="pass1",allowed_failure="false"} 1`) {
		t.Errorf("missing pass1 status=1 in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_status{name="fail1",allowed_failure="false"} 0`) {
		t.Errorf("missing fail1 status=0 in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_status{name="skip1",allowed_failure="false"} 0`) {
		t.Errorf("missing skip1 status=0 in:\n%s", out)
	}
}

func TestPrometheus_ZeroDuration(t *testing.T) {
	var buf bytes.Buffer
	r := NewPrometheus(&buf)

	r.TestResult(TestResultData{Name: "instant-test", Passed: true, Duration: 0})
	r.Summary(SuiteResultData{Total: 1, Passed: 1, Failed: 0, Duration: 0})

	out := buf.String()

	if !containsLine(out, "smoke_test_duration_seconds 0") {
		t.Errorf("missing suite duration 0 in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_duration_seconds_per_test{name="instant_test"} 0`) {
		t.Errorf("missing per-test zero duration in:\n%s", out)
	}
	if !containsLine(out, `smoke_test_status{name="instant_test",allowed_failure="false"} 1`) {
		t.Errorf("missing instant-test status line in:\n%s", out)
	}
}

func TestPrometheus_LongNameSpecialChars(t *testing.T) {
	var buf bytes.Buffer
	r := NewPrometheus(&buf)

	longName := "a-really/long.test:name(with)[many]{special}chars&and symbols!@#$%^&*"
	r.TestResult(TestResultData{Name: longName, Passed: true, Duration: 100 * time.Millisecond})
	r.Summary(SuiteResultData{Total: 1, Passed: 1, Duration: 100 * time.Millisecond})

	out := buf.String()

	// All non [a-zA-Z0-9_] should become underscores
	sanitized := `smoke_test_status{name="a_really_long_test_name_with__many__special_chars_and_symbols________",allowed_failure="false"} 1`
	if !containsLine(out, sanitized) {
		t.Errorf("long name not sanitized correctly in:\n%s", out)
	}

	durationLine := `smoke_test_duration_seconds_per_test{name="a_really_long_test_name_with__many__special_chars_and_symbols________"} 0.1`
	if !containsLine(out, durationLine) {
		t.Errorf("sanitized duration line for long name missing in:\n%s", out)
	}
}
