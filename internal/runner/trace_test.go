package runner

import (
	"regexp"
	"testing"
)

func TestNewTraceContext(t *testing.T) {
	tc := NewTraceContext()
	tid := tc.TraceID()

	if len(tid) != 32 {
		t.Errorf("TraceID length = %d, want 32", len(tid))
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(tid) {
		t.Errorf("TraceID = %q, want 32 lowercase hex chars", tid)
	}
	if !tc.Enabled {
		t.Error("Enabled = false, want true")
	}
}

func TestTraceContext_ChildSpan(t *testing.T) {
	tc := NewTraceContext()
	span := tc.NewSpan()
	sid := span.SpanID()

	if len(sid) != 16 {
		t.Errorf("SpanID length = %d, want 16", len(sid))
	}
	if !regexp.MustCompile(`^[0-9a-f]{16}$`).MatchString(sid) {
		t.Errorf("SpanID = %q, want 16 lowercase hex chars", sid)
	}
}

func TestSpanContext_Traceparent(t *testing.T) {
	tc := NewTraceContext()
	span := tc.NewSpan()
	tp := span.Traceparent()

	pattern := `^00-[0-9a-f]{32}-[0-9a-f]{16}-01$`
	if !regexp.MustCompile(pattern).MatchString(tp) {
		t.Errorf("Traceparent = %q, want format matching %s", tp, pattern)
	}
}

func TestTraceContext_DifferentTraceIDs(t *testing.T) {
	tc1 := NewTraceContext()
	tc2 := NewTraceContext()

	if tc1.TraceID() == tc2.TraceID() {
		t.Error("two trace contexts produced identical trace IDs")
	}
}

func TestTraceContext_DifferentSpanIDs(t *testing.T) {
	tc := NewTraceContext()
	span1 := tc.NewSpan()
	span2 := tc.NewSpan()

	if span1.SpanID() == span2.SpanID() {
		t.Error("two spans under the same trace produced identical span IDs")
	}
}

func TestDisabledTraceContext(t *testing.T) {
	var tc TraceContext

	if tc.Enabled {
		t.Error("zero-value TraceContext.Enabled = true, want false")
	}
	if tc.TraceID() != "00000000000000000000000000000000" {
		t.Errorf("zero-value TraceID = %q, want all zeros", tc.TraceID())
	}
}
