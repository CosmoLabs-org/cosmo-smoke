package reporter

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestOTelReporter_TestResult_SendsSpan(t *testing.T) {
	var received sync.Mutex
	var bodies []json.RawMessage

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/v1/traces" {
			t.Errorf("path = %q, want /v1/traces", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		body, _ := io.ReadAll(r.Body)
		received.Lock()
		bodies = append(bodies, body)
		received.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	r := NewOTelReporter(ts.URL+"/v1/traces", "smoke-test", nil)
	r.TestResult(TestResultData{
		Name:     "api-health",
		Passed:   true,
		Duration: 150 * time.Millisecond,
		Assertions: []AssertionDetail{
			{Type: "exit_code", Passed: true},
		},
	})

	// Wait for async send
	time.Sleep(100 * time.Millisecond)

	received.Lock()
	if len(bodies) != 1 {
		t.Fatalf("expected 1 request, got %d", len(bodies))
	}
	received.Unlock()

	// Parse OTLP JSON
	var otlp struct {
		ResourceSpans []struct {
			ScopeSpans []struct {
				Spans []struct {
					Name       string `json:"name"`
					TraceID    string `json:"traceId"`
					SpanID     string `json:"spanId"`
					Attributes []struct {
						Key   string `json:"key"`
						Value struct {
							StringValue string `json:"stringValue"`
						} `json:"value"`
					} `json:"attributes"`
				} `json:"spans"`
			} `json:"scopeSpans"`
		} `json:"resourceSpans"`
	}
	if err := json.Unmarshal(bodies[0], &otlp); err != nil {
		t.Fatalf("parse OTLP JSON: %v", err)
	}
	if len(otlp.ResourceSpans) == 0 || len(otlp.ResourceSpans[0].ScopeSpans) == 0 {
		t.Fatal("no resource spans in OTLP payload")
	}
	spans := otlp.ResourceSpans[0].ScopeSpans[0].Spans
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	s := spans[0]
	if s.Name != "smoke-test/api-health" {
		t.Errorf("span name = %q, want smoke-test/api-health", s.Name)
	}
	if s.TraceID == "" {
		t.Error("span traceId is empty")
	}
	if s.SpanID == "" {
		t.Error("span spanId is empty")
	}
	// Check attributes for test result
	foundStatus := false
	for _, a := range s.Attributes {
		if a.Key == "smoke.passed" {
			foundStatus = true
			if a.Value.StringValue != "true" {
				t.Errorf("smoke.passed = %q, want true", a.Value.StringValue)
			}
		}
	}
	if !foundStatus {
		t.Error("missing smoke.passed attribute")
	}
}

func TestOTelReporter_Summary_SendsSpan(t *testing.T) {
	var received sync.Mutex
	var bodyCount int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.Lock()
		bodyCount++
		received.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	r := NewOTelReporter(ts.URL+"/v1/traces", "smoke-test", nil)
	r.Summary(SuiteResultData{
		Project: "myapp",
		Total:   5,
		Passed:  3,
		Failed:  2,
	})

	time.Sleep(100 * time.Millisecond)

	received.Lock()
	if bodyCount != 1 {
		t.Errorf("expected 1 request, got %d", bodyCount)
	}
	received.Unlock()
}

func TestOTelReporter_CustomHeaders(t *testing.T) {
	var gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	headers := map[string]string{"Authorization": "Bearer test-token"}
	r := NewOTelReporter(ts.URL+"/v1/traces", "smoke-test", headers)
	r.TestResult(TestResultData{Name: "test1", Passed: true})

	time.Sleep(100 * time.Millisecond)

	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q, want Bearer test-token", gotAuth)
	}
}

func TestOTelReporter_CollectorUnreachable_NoPanic(t *testing.T) {
	r := NewOTelReporter("http://127.0.0.1:1/v1/traces", "smoke-test", nil)
	// Should not panic when collector is unreachable
	r.TestResult(TestResultData{Name: "test1", Passed: true})
	r.Summary(SuiteResultData{Project: "myapp", Total: 1, Passed: 1})
	time.Sleep(100 * time.Millisecond)
}
