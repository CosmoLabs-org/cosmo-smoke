package reporter

import (
	"encoding/json"
	"io"
)

// JSON collects all results and emits JSON on Summary.
type JSON struct {
	w       io.Writer
	prereqs []PrereqResultData
	tests   []TestResultData
}

// NewJSON creates a JSON reporter writing to w.
func NewJSON(w io.Writer) *JSON {
	return &JSON{w: w}
}

func (j *JSON) PrereqStart(_ string) {}

func (j *JSON) PrereqResult(r PrereqResultData) {
	j.prereqs = append(j.prereqs, r)
}

func (j *JSON) TestStart(_ string) {}

func (j *JSON) TestResult(r TestResultData) {
	j.tests = append(j.tests, r)
}

type jsonPrereq struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Output string `json:"output,omitempty"`
	Hint   string `json:"hint,omitempty"`
	Error  string `json:"error,omitempty"`
}

type jsonTest struct {
	Name           string            `json:"name"`
	Passed         bool              `json:"passed"`
	Skipped        bool              `json:"skipped,omitempty"`
	AllowedFailure bool              `json:"allowed_failure,omitempty"`
	DurationMs     int64             `json:"duration_ms"`
	Assertions     []AssertionDetail `json:"assertions"`
	Error          string            `json:"error,omitempty"`
}

type jsonOutput struct {
	Project         string       `json:"project"`
	Total           int          `json:"total"`
	Passed          int          `json:"passed"`
	Failed          int          `json:"failed"`
	Skipped         int          `json:"skipped"`
	AllowedFailures int          `json:"allowed_failures"`
	DurationMs      int64        `json:"duration_ms"`
	Prerequisites   []jsonPrereq `json:"prerequisites,omitempty"`
	Tests           []jsonTest   `json:"tests"`
}

func (j *JSON) Summary(s SuiteResultData) {
	out := jsonOutput{
		Project:         s.Project,
		Total:           s.Total,
		Passed:          s.Passed,
		Failed:          s.Failed,
		Skipped:         s.Skipped,
		AllowedFailures: s.AllowedFailures,
		DurationMs:      s.Duration.Milliseconds(),
	}

	for _, p := range j.prereqs {
		jp := jsonPrereq{
			Name:   p.Name,
			Passed: p.Passed,
			Output: p.Output,
			Hint:   p.Hint,
		}
		if p.Error != nil {
			jp.Error = p.Error.Error()
		}
		out.Prerequisites = append(out.Prerequisites, jp)
	}

	for _, t := range j.tests {
		jt := jsonTest{
			Name:           t.Name,
			Passed:         t.Passed,
			Skipped:        t.Skipped,
			AllowedFailure: t.AllowedFailure,
			DurationMs:     t.Duration.Milliseconds(),
			Assertions:     t.Assertions,
		}
		if t.Error != nil {
			jt.Error = t.Error.Error()
		}
		out.Tests = append(out.Tests, jt)
	}

	enc := json.NewEncoder(j.w)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}
