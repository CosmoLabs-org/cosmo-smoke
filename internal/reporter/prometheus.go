package reporter

import (
	"fmt"
	"io"
	"regexp"
)

var nonLabelChars = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// sanitizeLabel replaces characters that are not valid in Prometheus label
// values ([a-zA-Z0-9_]) with underscores.
func sanitizeLabel(s string) string {
	return nonLabelChars.ReplaceAllString(s, "_")
}

// Prometheus buffers test results and emits a Prometheus text-format metrics
// block at suite end. The output is a snapshot (gauges), suitable for piping
// to a Pushgateway or scraping by an exporter.
type Prometheus struct {
	w     io.Writer
	tests []TestResultData
}

// NewPrometheus creates a Prometheus reporter writing to w.
func NewPrometheus(w io.Writer) *Prometheus {
	return &Prometheus{w: w}
}

func (p *Prometheus) PrereqStart(_ string) {}

func (p *Prometheus) PrereqResult(_ PrereqResultData) {}

func (p *Prometheus) TestStart(_ string) {}

func (p *Prometheus) TestResult(r TestResultData) {
	p.tests = append(p.tests, r)
}

func (p *Prometheus) Summary(s SuiteResultData) {
	w := p.w

	// --- Summary metrics ---

	fmt.Fprintf(w, "# HELP smoke_test_total Total tests in this run.\n")
	fmt.Fprintf(w, "# TYPE smoke_test_total gauge\n")
	fmt.Fprintf(w, "smoke_test_total %d\n", s.Total)
	fmt.Fprintln(w)

	fmt.Fprintf(w, "# HELP smoke_test_passed_total Tests that passed.\n")
	fmt.Fprintf(w, "# TYPE smoke_test_passed_total gauge\n")
	fmt.Fprintf(w, "smoke_test_passed_total %d\n", s.Passed)
	fmt.Fprintln(w)

	fmt.Fprintf(w, "# HELP smoke_test_failed_total Tests that failed (excluding allowed failures).\n")
	fmt.Fprintf(w, "# TYPE smoke_test_failed_total gauge\n")
	fmt.Fprintf(w, "smoke_test_failed_total %d\n", s.Failed)
	fmt.Fprintln(w)

	fmt.Fprintf(w, "# HELP smoke_test_allowed_failure_total Tests that failed but had allow_failure=true.\n")
	fmt.Fprintf(w, "# TYPE smoke_test_allowed_failure_total gauge\n")
	fmt.Fprintf(w, "smoke_test_allowed_failure_total %d\n", s.AllowedFailures)
	fmt.Fprintln(w)

	durationSeconds := s.Duration.Seconds()
	fmt.Fprintf(w, "# HELP smoke_test_duration_seconds Total suite duration.\n")
	fmt.Fprintf(w, "# TYPE smoke_test_duration_seconds gauge\n")
	fmt.Fprintf(w, "smoke_test_duration_seconds %g\n", durationSeconds)

	if len(p.tests) == 0 {
		return
	}

	// --- Per-test status ---

	fmt.Fprintln(w)
	fmt.Fprintf(w, "# HELP smoke_test_status Per-test pass (1) / fail (0).\n")
	fmt.Fprintf(w, "# TYPE smoke_test_status gauge\n")
	for _, t := range p.tests {
		status := 0
		if t.Passed {
			status = 1
		}
		allowedStr := "false"
		if t.AllowedFailure {
			allowedStr = "true"
		}
		fmt.Fprintf(w, "smoke_test_status{name=%q,allowed_failure=%q} %d\n",
			sanitizeLabel(t.Name), allowedStr, status)
	}

	// --- Per-test duration ---

	fmt.Fprintln(w)
	fmt.Fprintf(w, "# HELP smoke_test_duration_seconds_per_test Per-test duration.\n")
	fmt.Fprintf(w, "# TYPE smoke_test_duration_seconds_per_test gauge\n")
	for _, t := range p.tests {
		fmt.Fprintf(w, "smoke_test_duration_seconds_per_test{name=%q} %g\n",
			sanitizeLabel(t.Name), t.Duration.Seconds())
	}
}
