package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/reporter"
	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

// RunOptions controls test execution behavior.
type RunOptions struct {
	Tags        []string
	ExcludeTags []string
	FailFast    bool
	DryRun      bool
	Timeout     time.Duration // per-test override (0 = use config)
}

// SuiteResult holds the aggregate outcome.
type SuiteResult struct {
	Project         string
	Total           int
	Passed          int
	Failed          int
	Skipped         int
	AllowedFailures int
	Duration        time.Duration
	Tests           []TestResult
}

// TestResult holds one test's outcome.
type TestResult struct {
	Name           string
	Passed         bool
	Skipped        bool
	AllowedFailure bool // true if test failed but allow_failure was set
	Duration       time.Duration
	Assertions     []AssertionResult
	Error          error
	Attempts       int // number of attempts made (1 = no retry)
}

// Runner executes smoke tests from a config.
type Runner struct {
	Config    *schema.SmokeConfig
	Reporter  reporter.Reporter
	ConfigDir string
}

// Run executes all tests per the options and returns the suite result.
func (r *Runner) Run(opts RunOptions) (*SuiteResult, error) {
	start := time.Now()

	// Run prerequisites
	if len(r.Config.Prereqs) > 0 {
		timeout := r.Config.Settings.Timeout.Duration
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		results := CheckPrerequisites(r.Config.Prereqs, timeout)
		for _, pr := range results {
			r.Reporter.PrereqStart(pr.Name)
			r.Reporter.PrereqResult(reporter.PrereqResultData{
				Name:   pr.Name,
				Passed: pr.Passed,
				Output: pr.Output,
				Hint:   pr.Hint,
				Error:  pr.Error,
			})
		}
		// Check for prereq failures
		for _, pr := range results {
			if !pr.Passed {
				return nil, fmt.Errorf("prerequisite %q failed: %v", pr.Name, pr.Error)
			}
		}
	}

	// Filter tests by tags
	tests := filterTests(r.Config.Tests, opts.Tags, opts.ExcludeTags)

	suite := &SuiteResult{
		Project: r.Config.Project,
		Total:   len(tests),
	}

	failFast := opts.FailFast || r.Config.Settings.FailFast

	if r.Config.Settings.Parallel && !failFast {
		r.runParallel(tests, opts, suite)
	} else {
		r.runSequential(tests, opts, suite, failFast)
	}

	suite.Duration = time.Since(start)

	// Report summary
	r.Reporter.Summary(reporter.SuiteResultData{
		Project:         suite.Project,
		Total:           suite.Total,
		Passed:          suite.Passed,
		Failed:          suite.Failed,
		Skipped:         suite.Skipped,
		AllowedFailures: suite.AllowedFailures,
		Duration:        suite.Duration,
	})

	return suite, nil
}

func (r *Runner) runSequential(tests []schema.Test, opts RunOptions, suite *SuiteResult, failFast bool) {
	stopped := false
	for _, t := range tests {
		if stopped {
			tr := TestResult{Name: t.Name, Skipped: true}
			suite.Tests = append(suite.Tests, tr)
			suite.Skipped++
			r.Reporter.TestStart(t.Name)
			r.Reporter.TestResult(reporter.TestResultData{Name: t.Name, Skipped: true})
			continue
		}

		tr := r.runTest(t, opts)
		suite.Tests = append(suite.Tests, tr)
		if tr.Passed {
			suite.Passed++
		} else if tr.Skipped {
			suite.Skipped++
		} else if tr.AllowedFailure {
			suite.AllowedFailures++
		} else {
			suite.Failed++
			if failFast {
				stopped = true
			}
		}
	}
}

func (r *Runner) runParallel(tests []schema.Test, opts RunOptions, suite *SuiteResult) {
	results := make([]TestResult, len(tests))
	var wg sync.WaitGroup

	for i, t := range tests {
		wg.Add(1)
		go func(idx int, test schema.Test) {
			defer wg.Done()
			results[idx] = r.runTest(test, opts)
		}(i, t)
	}
	wg.Wait()

	for _, tr := range results {
		suite.Tests = append(suite.Tests, tr)
		if tr.Passed {
			suite.Passed++
		} else if tr.Skipped {
			suite.Skipped++
		} else if tr.AllowedFailure {
			suite.AllowedFailures++
		} else {
			suite.Failed++
		}
	}
}

func (r *Runner) runTest(t schema.Test, opts RunOptions) TestResult {
	if t.Retry == nil || t.Retry.Count <= 1 {
		res := r.runTestOnce(t, opts)
		if res.Attempts == 0 {
			res.Attempts = 1
		}
		return res
	}
	var last TestResult
	backoff := t.Retry.Backoff.Duration
	for attempt := 1; attempt <= t.Retry.Count; attempt++ {
		last = r.runTestOnce(t, opts)
		if last.Passed {
			last.Attempts = attempt
			return last
		}
		if attempt < t.Retry.Count {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	last.Attempts = t.Retry.Count
	return last
}

func (r *Runner) runTestOnce(t schema.Test, opts RunOptions) TestResult {
	r.Reporter.TestStart(t.Name)
	start := time.Now()

	if opts.DryRun {
		tr := TestResult{Name: t.Name, Passed: true, Duration: time.Since(start)}
		r.Reporter.TestResult(toReporterResult(tr))
		return tr
	}

	// Determine timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = t.Timeout.Duration
	}
	if timeout == 0 {
		timeout = r.Config.Settings.Timeout.Duration
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Run cleanup via defer
	if t.Cleanup != "" {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			c := exec.CommandContext(ctx, "sh", "-c", t.Cleanup)
			c.Dir = r.ConfigDir
			c.Run()
		}()
	}

	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", t.Run)
	cmd.Dir = r.ConfigDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			tr := TestResult{
				Name:           t.Name,
				AllowedFailure: t.AllowFailure,
				Duration:       time.Since(start),
				Error:          err,
			}
			r.Reporter.TestResult(toReporterResult(tr))
			return tr
		}
	}

	// Evaluate assertions
	var assertions []AssertionResult
	allPassed := true

	if t.Expect.ExitCode != nil {
		a := CheckExitCode(exitCode, *t.Expect.ExitCode)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.StdoutContains != "" {
		a := CheckStdoutContains(stdout.String(), t.Expect.StdoutContains)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.StdoutMatches != "" {
		a := CheckStdoutMatches(stdout.String(), t.Expect.StdoutMatches)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.StderrContains != "" {
		a := CheckStderrContains(stderr.String(), t.Expect.StderrContains)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.StderrMatches != "" {
		a := CheckStderrMatches(stderr.String(), t.Expect.StderrMatches)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.EnvExists != "" {
		a := CheckEnvExists(t.Expect.EnvExists)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.FileExists != "" {
		a := CheckFileExists(t.Expect.FileExists, r.ConfigDir)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.PortListening != nil {
		p := t.Expect.PortListening
		a := CheckPortListening(p.Port, p.Protocol, p.Host)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.ProcessRunning != "" {
		a := CheckProcessRunning(t.Expect.ProcessRunning)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.HTTP != nil {
		httpResults := CheckHTTP(t.Expect.HTTP)
		for _, a := range httpResults {
			assertions = append(assertions, a)
			if !a.Passed {
				allPassed = false
			}
		}
	}
	if t.Expect.JSONField != nil {
		jsonResults := CheckJSONField(stdout.String(), t.Expect.JSONField)
		for _, a := range jsonResults {
			assertions = append(assertions, a)
			if !a.Passed {
				allPassed = false
			}
		}
	}
	if t.Expect.SSLCert != nil {
		a := CheckSSLCert(t.Expect.SSLCert)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.Redis != nil {
		a := CheckRedisPing(t.Expect.Redis)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.Memcached != nil {
		a := CheckMemcachedVersion(t.Expect.Memcached)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.Postgres != nil {
		a := CheckPostgresPing(t.Expect.Postgres)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.MySQL != nil {
		a := CheckMySQLPing(t.Expect.MySQL)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}
	if t.Expect.GRPCHealth != nil {
		a := CheckGRPCHealth(t.Expect.GRPCHealth)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}

	duration := time.Since(start)

	if t.Expect.ResponseTimeMs != nil {
		a := CheckResponseTime(int(duration.Milliseconds()), *t.Expect.ResponseTimeMs)
		assertions = append(assertions, a)
		if !a.Passed {
			allPassed = false
		}
	}

	tr := TestResult{
		Name:           t.Name,
		Passed:         allPassed,
		AllowedFailure: !allPassed && t.AllowFailure,
		Duration:       duration,
		Assertions:     assertions,
	}
	r.Reporter.TestResult(toReporterResult(tr))
	return tr
}

func toReporterResult(tr TestResult) reporter.TestResultData {
	var assertions []reporter.AssertionDetail
	for _, a := range tr.Assertions {
		assertions = append(assertions, reporter.AssertionDetail{
			Type:     a.Type,
			Expected: a.Expected,
			Actual:   a.Actual,
			Passed:   a.Passed,
		})
	}
	return reporter.TestResultData{
		Name:           tr.Name,
		Passed:         tr.Passed,
		Skipped:        tr.Skipped,
		AllowedFailure: tr.AllowedFailure,
		Duration:       tr.Duration,
		Assertions:     assertions,
		Error:          tr.Error,
	}
}

func filterTests(tests []schema.Test, include, exclude []string) []schema.Test {
	if len(include) == 0 && len(exclude) == 0 {
		return tests
	}

	var filtered []schema.Test
	for _, t := range tests {
		if len(include) > 0 && !hasAnyTag(t.Tags, include) {
			continue
		}
		if len(exclude) > 0 && hasAnyTag(t.Tags, exclude) {
			continue
		}
		filtered = append(filtered, t)
	}
	return filtered
}

func hasAnyTag(tags, targets []string) bool {
	for _, tag := range tags {
		for _, target := range targets {
			if strings.EqualFold(tag, target) {
				return true
			}
		}
	}
	return false
}
