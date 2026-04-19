package runner

import (
	"crypto/rand"
	"encoding/hex"
)

// TraceContext holds a W3C trace context for distributed tracing correlation.
type TraceContext struct {
	traceID [16]byte
	Enabled bool
}

// SpanContext holds a single span within a trace.
type SpanContext struct {
	traceID [16]byte
	spanID  [8]byte
}

// NewTraceContext creates a new trace context with a random 128-bit trace ID.
func NewTraceContext() *TraceContext {
	var tc TraceContext
	rand.Read(tc.traceID[:])
	tc.Enabled = true
	return &tc
}

// TraceID returns the hex-encoded 32-character trace ID.
func (tc *TraceContext) TraceID() string {
	return hex.EncodeToString(tc.traceID[:])
}

// NewSpan creates a child span with a random 64-bit span ID under this trace.
func (tc *TraceContext) NewSpan() *SpanContext {
	var sc SpanContext
	sc.traceID = tc.traceID
	rand.Read(sc.spanID[:])
	return &sc
}

// SpanID returns the hex-encoded 16-character span ID.
func (sc *SpanContext) SpanID() string {
	return hex.EncodeToString(sc.spanID[:])
}

// Traceparent returns the W3C traceparent header value (version 00, flags 01).
func (sc *SpanContext) Traceparent() string {
	return "00-" + hex.EncodeToString(sc.traceID[:]) + "-" + hex.EncodeToString(sc.spanID[:]) + "-01"
}
