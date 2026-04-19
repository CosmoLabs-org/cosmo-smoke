package runner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

// jaegerJSON builds a Jaeger API response body with the given span count.
func jaegerJSON(traceID string, spanCount int) string {
	type span struct {
		TraceID string `json:"traceID"`
		SpanID  string `json:"spanID"`
	}
	spans := make([]span, spanCount)
	for i := range spans {
		spans[i] = span{TraceID: traceID, SpanID: fmt.Sprintf("span%d", i)}
	}
	type trace struct {
		TraceID string `json:"traceID"`
		Spans   []span `json:"spans"`
	}
	data := trace{TraceID: traceID, Spans: spans}
	b, _ := json.Marshal(map[string]interface{}{"data": []trace{data}})
	return string(b)
}

func TestCheckOTelTrace_TraceFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jaegerJSON("abc123", 1))
	}))
	defer ts.Close()

	check := &schema.OTelTraceCheck{
		JaegerURL:   ts.URL,
		ServiceName: "myservice",
		MinSpans:    1,
		Timeout:     schema.Duration{Duration: 2 * time.Second},
	}
	result := CheckOTelTrace(check, "abc123", ts.Client())
	if !result.Passed {
		t.Errorf("expected pass, got fail: %s", result.Actual)
	}
}

func TestCheckOTelTrace_MinSpansNotMet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jaegerJSON("abc123", 1))
	}))
	defer ts.Close()

	check := &schema.OTelTraceCheck{
		JaegerURL:   ts.URL,
		ServiceName: "myservice",
		MinSpans:    5,
		Timeout:     schema.Duration{Duration: 100 * time.Millisecond},
	}
	result := CheckOTelTrace(check, "abc123", ts.Client())
	if result.Passed {
		t.Error("expected fail, got pass")
	}
}

func TestCheckOTelTrace_TimeoutNoSpans(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[]}`)
	}))
	defer ts.Close()

	check := &schema.OTelTraceCheck{
		JaegerURL:   ts.URL,
		ServiceName: "myservice",
		MinSpans:    1,
		Timeout:     schema.Duration{Duration: 100 * time.Millisecond},
	}
	result := CheckOTelTrace(check, "abc123", ts.Client())
	if result.Passed {
		t.Error("expected fail, got pass")
	}
}

func TestCheckOTelTrace_CollectorUnreachable(t *testing.T) {
	// Port 1 is not listening — forces connection refused.
	check := &schema.OTelTraceCheck{
		JaegerURL:   "http://127.0.0.1:1",
		ServiceName: "myservice",
		MinSpans:    1,
		Timeout:     schema.Duration{Duration: 100 * time.Millisecond},
	}
	result := CheckOTelTrace(check, "abc123", http.DefaultClient)
	if result.Passed {
		t.Error("expected fail, got pass")
	}
}

func TestCheckOTelTrace_PollingRetries(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount < 3 {
			fmt.Fprint(w, `{"data":[]}`)
			return
		}
		fmt.Fprint(w, jaegerJSON("retry-trace", 2))
	}))
	defer ts.Close()

	check := &schema.OTelTraceCheck{
		JaegerURL:   ts.URL,
		ServiceName: "myservice",
		MinSpans:    1,
		Timeout:     schema.Duration{Duration: 5 * time.Second},
	}
	result := CheckOTelTrace(check, "retry-trace", ts.Client())
	if !result.Passed {
		t.Errorf("expected pass after retries, got fail: %s", result.Actual)
	}
	if callCount < 3 {
		t.Errorf("expected >= 3 calls, got %d", callCount)
	}
}
