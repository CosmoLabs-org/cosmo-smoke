package reporter

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// JUnit collects all results and emits JUnit XML on Summary.
type JUnit struct {
	w     io.Writer
	tests []TestResultData
}

// NewJUnit creates a JUnit reporter writing to w.
func NewJUnit(w io.Writer) *JUnit {
	return &JUnit{w: w}
}

func (j *JUnit) PrereqStart(_ string) {}

func (j *JUnit) PrereqResult(_ PrereqResultData) {}

func (j *JUnit) TestStart(_ string) {}

func (j *JUnit) TestResult(r TestResultData) {
	j.tests = append(j.tests, r)
}

// junitTestSuites is the root element.
type junitTestSuites struct {
	XMLName  xml.Name        `xml:"testsuites"`
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Time     string          `xml:"time,attr"`
	Suites   []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	Name      string            `xml:"name,attr"`
	Tests     int               `xml:"tests,attr"`
	Failures  int               `xml:"failures,attr"`
	Skipped   int               `xml:"skipped,attr,omitempty"`
	Time      string            `xml:"time,attr"`
	Timestamp string            `xml:"timestamp,attr,omitempty"`
	Hostname  string            `xml:"hostname,attr,omitempty"`
	Properties *junitProperties `xml:"properties,omitempty"`
	Cases     []junitTestCase   `xml:"testcase"`
}

type junitProperties struct {
	Props []junitProperty `xml:"property"`
}

type junitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type junitTestCase struct {
	Name    string         `xml:"name,attr"`
	Time    string         `xml:"time,attr"`
	Failure *junitFailure  `xml:"failure,omitempty"`
	Skipped *junitSkipped  `xml:"skipped,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

type junitSkipped struct{}

func formatSeconds(d float64) string {
	return fmt.Sprintf("%.3f", d)
}

func (j *JUnit) Summary(s SuiteResultData) {
	totalSeconds := s.Duration.Seconds()

	hostname, _ := os.Hostname()

	props := &junitProperties{
		Props: []junitProperty{
			{Name: "project", Value: s.Project},
			{Name: "passed", Value: fmt.Sprintf("%d", s.Passed)},
			{Name: "failed", Value: fmt.Sprintf("%d", s.Failed)},
			{Name: "skipped", Value: fmt.Sprintf("%d", s.Skipped)},
		},
	}

	suite := junitTestSuite{
		Name:      s.Project,
		Tests:     s.Total,
		Failures:  s.Failed,
		Skipped:   s.Skipped,
		Time:      formatSeconds(totalSeconds),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Hostname:  hostname,
		Properties: props,
	}

	for _, t := range j.tests {
		tc := junitTestCase{
			Name: t.Name,
			Time: formatSeconds(t.Duration.Seconds()),
		}

		if t.Skipped || t.AllowedFailure {
			tc.Skipped = &junitSkipped{}
		} else if !t.Passed {
			// Build failure message from failed assertions.
			var msgs []string
			for _, a := range t.Assertions {
				if !a.Passed {
					msgs = append(msgs, fmt.Sprintf("%s: expected %q not found in %q", a.Type, a.Expected, a.Actual))
				}
			}
			if t.Error != nil {
				msgs = append(msgs, t.Error.Error())
			}

			message := strings.Join(msgs, "; ")
			if message == "" {
				message = "test failed"
			}

			// Build detailed body from failed assertions.
			var lines []string
			for _, a := range t.Assertions {
				if !a.Passed {
					lines = append(lines, fmt.Sprintf("  %s:\n    Expected: %s\n    Actual:   %s", a.Type, a.Expected, a.Actual))
				}
			}
			body := strings.Join(lines, "\n")

			tc.Failure = &junitFailure{
				Message: message,
				Text:    body,
			}
		}

		suite.Cases = append(suite.Cases, tc)
	}

	root := junitTestSuites{
		Name:     "smoke",
		Tests:    s.Total,
		Failures: s.Failed,
		Time:     formatSeconds(totalSeconds),
		Suites:   []junitTestSuite{suite},
	}

	fmt.Fprintln(j.w, `<?xml version="1.0" encoding="UTF-8"?>`)
	enc := xml.NewEncoder(j.w)
	enc.Indent("", "  ")
	enc.Encode(root)
}
