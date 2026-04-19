package reporter

// MultiReporter fans out events to multiple reporters.
type MultiReporter struct {
	reporters []Reporter
}

// NewMultiReporter creates a reporter that delegates to all given reporters.
func NewMultiReporter(reporters ...Reporter) *MultiReporter {
	return &MultiReporter{reporters: reporters}
}

func (m *MultiReporter) PrereqStart(name string) {
	for _, r := range m.reporters {
		r.PrereqStart(name)
	}
}

func (m *MultiReporter) PrereqResult(r PrereqResultData) {
	for _, rep := range m.reporters {
		rep.PrereqResult(r)
	}
}

func (m *MultiReporter) TestStart(name string) {
	for _, r := range m.reporters {
		r.TestStart(name)
	}
}

func (m *MultiReporter) TestResult(r TestResultData) {
	for _, rep := range m.reporters {
		rep.TestResult(r)
	}
}

func (m *MultiReporter) Summary(s SuiteResultData) {
	for _, r := range m.reporters {
		r.Summary(s)
	}
}
