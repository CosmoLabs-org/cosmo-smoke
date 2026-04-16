package reporter

import (
	"fmt"
	"io"
)

// TAP produces Test Anything Protocol v14 output.
type TAP struct {
	w       io.Writer
	results []TestResultData
	count   int
}

// NewTAP creates a TAP reporter writing to w.
func NewTAP(w io.Writer) *TAP { return &TAP{w: w} }

func (t *TAP) PrereqStart(name string)            {}
func (t *TAP) PrereqResult(data PrereqResultData) {}
func (t *TAP) TestStart(name string)              { t.count++ }
func (t *TAP) TestResult(data TestResultData) {
	t.results = append(t.results, data)
}

func (t *TAP) Summary(data SuiteResultData) {
	fmt.Fprintf(t.w, "TAP version 14\n")
	fmt.Fprintf(t.w, "1..%d\n", len(t.results))
	for i, r := range t.results {
		if r.Skipped {
			fmt.Fprintf(t.w, "ok %d - %s # SKIP\n", i+1, r.Name)
		} else if r.Passed {
			fmt.Fprintf(t.w, "ok %d - %s\n", i+1, r.Name)
		} else if r.AllowedFailure {
			fmt.Fprintf(t.w, "not ok %d - %s # TODO allow_failure\n", i+1, r.Name)
			for _, a := range r.Assertions {
				if !a.Passed {
					fmt.Fprintf(t.w, "# %s: expected %s, got %s\n", a.Type, a.Expected, a.Actual)
				}
			}
		} else {
			fmt.Fprintf(t.w, "not ok %d - %s\n", i+1, r.Name)
			for _, a := range r.Assertions {
				if !a.Passed {
					fmt.Fprintf(t.w, "# %s: expected %s, got %s\n", a.Type, a.Expected, a.Actual)
				}
			}
		}
	}
}
