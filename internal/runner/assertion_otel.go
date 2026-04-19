package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

type jaegerResponse struct {
	Data []jaegerTrace `json:"data"`
}

type jaegerTrace struct {
	TraceID string       `json:"traceID"`
	Spans   []jaegerSpan `json:"spans"`
}

type jaegerSpan struct {
	TraceID       string `json:"traceID"`
	SpanID        string `json:"spanID"`
	OperationName string `json:"operationName"`
}

// CheckOTelTrace queries a Jaeger-compatible API to verify that a trace
// arrived at the collector with at least MinSpans spans within Timeout.
func CheckOTelTrace(check *schema.OTelTraceCheck, traceID string, client *http.Client) AssertionResult {
	timeout := check.Timeout.Duration
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	minSpans := check.MinSpans
	if minSpans == 0 {
		minSpans = 1
	}
	url := fmt.Sprintf("%s/api/traces/%s?service=%s", check.JaegerURL, traceID, check.ServiceName)

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := queryJaeger(client, url)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if len(resp.Data) > 0 && len(resp.Data[0].Spans) >= minSpans {
			return AssertionResult{
				Type:     "otel_trace",
				Expected: fmt.Sprintf(">=%d spans for trace %s", minSpans, traceID),
				Actual:   fmt.Sprintf("%d spans found", len(resp.Data[0].Spans)),
				Passed:   true,
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Final attempt after deadline expired.
	resp, err := queryJaeger(client, url)
	if err != nil {
		return AssertionResult{
			Type:     "otel_trace",
			Expected: fmt.Sprintf(">=%d spans for trace %s", minSpans, traceID),
			Actual:   fmt.Sprintf("collector error: %v", err),
			Passed:   false,
		}
	}
	if len(resp.Data) > 0 {
		return AssertionResult{
			Type:     "otel_trace",
			Expected: fmt.Sprintf(">=%d spans for trace %s", minSpans, traceID),
			Actual:   fmt.Sprintf("%d spans found within %s", len(resp.Data[0].Spans), timeout),
			Passed:   len(resp.Data[0].Spans) >= minSpans,
		}
	}
	return AssertionResult{
		Type:     "otel_trace",
		Expected: fmt.Sprintf(">=%d spans for trace %s", minSpans, traceID),
		Actual:   fmt.Sprintf("no spans found for trace %s within %s", traceID, timeout),
		Passed:   false,
	}
}

func queryJaeger(client *http.Client, url string) (*jaegerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var jr jaegerResponse
	if err := json.NewDecoder(resp.Body).Decode(&jr); err != nil {
		return nil, err
	}
	return &jr, nil
}
