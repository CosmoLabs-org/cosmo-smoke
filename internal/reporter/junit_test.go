package reporter

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
	"time"
)

// xmlTestSuites mirrors junitTestSuites for test parsing.
type xmlTestSuites struct {
	XMLName  xml.Name       `xml:"testsuites"`
	Name     string         `xml:"name,attr"`
	Tests    int            `xml:"tests,attr"`
	Failures int            `xml:"failures,attr"`
	Time     string         `xml:"time,attr"`
	Suites   []xmlTestSuite `xml:"testsuite"`
}

type xmlTestSuite struct {
	Name      string           `xml:"name,attr"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Skipped   int              `xml:"skipped,attr"`
	Time      string           `xml:"time,attr"`
	Timestamp string           `xml:"timestamp,attr,omitempty"`
	Hostname  string           `xml:"hostname,attr,omitempty"`
	Properties *xmlProperties  `xml:"properties,omitempty"`
	Cases     []xmlTestCase    `xml:"testcase"`
}

type xmlProperties struct {
	Props []xmlProperty `xml:"property"`
}

type xmlProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type xmlTestCase struct {
	Name    string       `xml:"name,attr"`
	Time    string       `xml:"time,attr"`
	Failure *xmlFailure  `xml:"failure"`
	Skipped *xmlSkipped  `xml:"skipped"`
}

type xmlFailure struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

type xmlSkipped struct{}

func parseJUnit(t *testing.T, buf *bytes.Buffer) xmlTestSuites {
	t.Helper()
	// Strip the XML declaration line before unmarshalling.
	raw := buf.String()
	idx := strings.Index(raw, "\n")
	if idx >= 0 {
		raw = raw[idx+1:]
	}
	var result xmlTestSuites
	if err := xml.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("invalid JUnit XML: %v\nOutput:\n%s", err, buf.String())
	}
	return result
}

func TestJUnit_ValidOutput(t *testing.T) {
	var buf bytes.Buffer
	r := NewJUnit(&buf)

	r.PrereqStart("Go installed")
	r.PrereqResult(PrereqResultData{Name: "Go installed", Passed: true, Output: "go1.26.2"})

	r.TestStart("Compiles")
	r.TestResult(TestResultData{
		Name:   "Compiles",
		Passed: true,
		Assertions: []AssertionDetail{
			{Type: "exit_code", Expected: "0", Actual: "0", Passed: true},
		},
		Duration: 800 * time.Millisecond,
	})

	r.TestStart("CLI works")
	r.TestResult(TestResultData{
		Name:   "CLI works",
		Passed: false,
		Assertions: []AssertionDetail{
			{Type: "stdout_contains", Expected: "Usage", Actual: "error: unknown command", Passed: false},
		},
		Duration: 500 * time.Millisecond,
		Error:    fmt.Errorf("exit code 1"),
	})

	r.Summary(SuiteResultData{
		Project:  "myapp",
		Total:    2,
		Passed:   1,
		Failed:   1,
		Duration: 1300 * time.Millisecond,
	})

	result := parseJUnit(t, &buf)

	// Top-level testsuites attributes.
	if result.Name != "smoke" {
		t.Errorf("testsuites name = %q, want %q", result.Name, "smoke")
	}
	if result.Tests != 2 {
		t.Errorf("testsuites tests = %d, want 2", result.Tests)
	}
	if result.Failures != 1 {
		t.Errorf("testsuites failures = %d, want 1", result.Failures)
	}

	// Single testsuite.
	if len(result.Suites) != 1 {
		t.Fatalf("suite count = %d, want 1", len(result.Suites))
	}
	suite := result.Suites[0]
	if suite.Name != "myapp" {
		t.Errorf("testsuite name = %q, want %q", suite.Name, "myapp")
	}
	if suite.Tests != 2 {
		t.Errorf("testsuite tests = %d, want 2", suite.Tests)
	}
	if suite.Failures != 1 {
		t.Errorf("testsuite failures = %d, want 1", suite.Failures)
	}

	if len(suite.Cases) != 2 {
		t.Fatalf("testcase count = %d, want 2", len(suite.Cases))
	}

	// Passing test — no failure element.
	tc0 := suite.Cases[0]
	if tc0.Name != "Compiles" {
		t.Errorf("testcase[0] name = %q, want %q", tc0.Name, "Compiles")
	}
	if tc0.Failure != nil {
		t.Errorf("testcase[0] should not have a failure element")
	}

	// Failing test — failure element with message and body.
	tc1 := suite.Cases[1]
	if tc1.Name != "CLI works" {
		t.Errorf("testcase[1] name = %q, want %q", tc1.Name, "CLI works")
	}
	if tc1.Failure == nil {
		t.Fatal("testcase[1] should have a failure element")
	}
	if !strings.Contains(tc1.Failure.Message, "stdout_contains") {
		t.Errorf("failure message = %q, should contain %q", tc1.Failure.Message, "stdout_contains")
	}
	if !strings.Contains(tc1.Failure.Text, "Expected") {
		t.Errorf("failure body = %q, should contain 'Expected'", tc1.Failure.Text)
	}
}

func TestJUnit_EmptyResults(t *testing.T) {
	var buf bytes.Buffer
	r := NewJUnit(&buf)
	r.Summary(SuiteResultData{Project: "empty", Total: 0})

	result := parseJUnit(t, &buf)
	if result.Tests != 0 {
		t.Errorf("tests = %d, want 0", result.Tests)
	}
	if len(result.Suites) != 1 {
		t.Fatalf("suite count = %d, want 1", len(result.Suites))
	}
	if result.Suites[0].Name != "empty" {
		t.Errorf("suite name = %q, want %q", result.Suites[0].Name, "empty")
	}
}

func TestJUnit_SkippedTest(t *testing.T) {
	var buf bytes.Buffer
	r := NewJUnit(&buf)

	r.TestResult(TestResultData{
		Name:     "Skipped test",
		Skipped:  true,
		Duration: 0,
	})

	r.Summary(SuiteResultData{
		Project: "proj",
		Total:   1,
		Skipped: 1,
	})

	result := parseJUnit(t, &buf)
	suite := result.Suites[0]
	if suite.Skipped != 1 {
		t.Errorf("skipped = %d, want 1", suite.Skipped)
	}
	if len(suite.Cases) != 1 {
		t.Fatalf("case count = %d, want 1", len(suite.Cases))
	}
	if suite.Cases[0].Skipped == nil {
		t.Error("testcase should have a skipped element")
	}
	if suite.Cases[0].Failure != nil {
		t.Error("skipped testcase should not have a failure element")
	}
}

func TestJUnit_MultipleFailedAssertions(t *testing.T) {
	var buf bytes.Buffer
	r := NewJUnit(&buf)

	r.TestResult(TestResultData{
		Name:   "Multi-fail",
		Passed: false,
		Assertions: []AssertionDetail{
			{Type: "exit_code", Expected: "0", Actual: "2", Passed: false},
			{Type: "stdout_contains", Expected: "OK", Actual: "", Passed: false},
		},
		Duration: 100 * time.Millisecond,
	})

	r.Summary(SuiteResultData{Project: "proj", Total: 1, Failed: 1})

	result := parseJUnit(t, &buf)
	tc := result.Suites[0].Cases[0]
	if tc.Failure == nil {
		t.Fatal("testcase should have a failure element")
	}
	// Both assertion types should appear in the message.
	if !strings.Contains(tc.Failure.Message, "exit_code") {
		t.Errorf("failure message should contain exit_code assertion: %q", tc.Failure.Message)
	}
	if !strings.Contains(tc.Failure.Message, "stdout_contains") {
		t.Errorf("failure message should contain stdout_contains assertion: %q", tc.Failure.Message)
	}
}

func TestJUnit_XMLDeclaration(t *testing.T) {
	var buf bytes.Buffer
	r := NewJUnit(&buf)
	r.Summary(SuiteResultData{Project: "p", Total: 0})

	out := buf.String()
	if !strings.HasPrefix(out, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Errorf("output should start with XML declaration, got: %q", out[:min(len(out), 60)])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestJUnit_HasTimestampAndHostname(t *testing.T) {
	var buf bytes.Buffer
	r := NewJUnit(&buf)
	r.Summary(SuiteResultData{Project: "meta-test", Total: 1, Passed: 1, Duration: 100 * time.Millisecond})

	result := parseJUnit(t, &buf)
	suite := result.Suites[0]

	if suite.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}
	if suite.Hostname == "" {
		t.Error("expected hostname to be set")
	}
}

func TestJUnit_HasProperties(t *testing.T) {
	var buf bytes.Buffer
	r := NewJUnit(&buf)
	r.Summary(SuiteResultData{Project: "prop-test", Total: 3, Passed: 2, Failed: 1, Duration: 200 * time.Millisecond})

	result := parseJUnit(t, &buf)
	suite := result.Suites[0]

	if suite.Properties == nil {
		t.Fatal("expected properties element")
	}
	propMap := make(map[string]string)
	for _, p := range suite.Properties.Props {
		propMap[p.Name] = p.Value
	}
	if propMap["project"] != "prop-test" {
		t.Errorf("project property = %q, want %q", propMap["project"], "prop-test")
	}
	if propMap["passed"] != "2" {
		t.Errorf("passed property = %q, want %q", propMap["passed"], "2")
	}
	if propMap["failed"] != "1" {
		t.Errorf("failed property = %q, want %q", propMap["failed"], "1")
	}
}
