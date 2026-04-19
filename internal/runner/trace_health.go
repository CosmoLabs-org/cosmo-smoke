package runner

import "math"

// TraceHealthTracker tracks otel_trace assertion results over a sliding window.
type TraceHealthTracker struct {
	window    []bool
	size      int
	confirmed int
}

// NewTraceHealthTracker creates a tracker with the given window size.
func NewTraceHealthTracker(windowSize int) *TraceHealthTracker {
	if windowSize < 1 {
		windowSize = 10
	}
	return &TraceHealthTracker{size: windowSize}
}

// Record adds a trace confirmation result to the sliding window.
func (t *TraceHealthTracker) Record(confirmed bool) {
	if len(t.window) >= t.size {
		// Evict oldest
		if t.window[0] {
			t.confirmed--
		}
		t.window = t.window[1:]
	}
	t.window = append(t.window, confirmed)
	if confirmed {
		t.confirmed++
	}
}

// Total returns the number of results in the current window.
func (t *TraceHealthTracker) Total() int {
	return len(t.window)
}

// Confirmed returns the number of confirmed (passing) trace results.
func (t *TraceHealthTracker) Confirmed() int {
	return t.confirmed
}

// HealthPct returns the percentage of confirmed traces (0-100).
// Returns 100 when empty (no data yet — healthy default).
func (t *TraceHealthTracker) HealthPct() float64 {
	if len(t.window) == 0 {
		return 100.0
	}
	return math.Round(float64(t.confirmed)/float64(len(t.window))*1000) / 10
}

// Degraded returns true when health is below the given threshold percentage.
// Returns false when empty (no data to judge).
func (t *TraceHealthTracker) Degraded(thresholdPct float64) bool {
	if len(t.window) == 0 {
		return false
	}
	return t.HealthPct() < thresholdPct
}

// Reset clears all tracked results.
func (t *TraceHealthTracker) Reset() {
	t.window = nil
	t.confirmed = 0
}
