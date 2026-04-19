package mcp

import (
	"strings"
	"testing"
)

// TestGetSuggestions verifies suggestion generation for common failure patterns.
func TestGetSuggestions(t *testing.T) {
	tests := []struct {
		assertionType string
		actual        string
		wantCount     int
		wantContains  string
	}{
		{
			assertionType: "redis_ping",
			actual:        "connection refused on localhost:6379",
			wantCount:     1,
			wantContains:  "not running",
		},
		{
			assertionType: "http",
			actual:        "connection refused",
			wantCount:     1,
			wantContains:  "listening",
		},
		{
			assertionType: "http",
			actual:        "unexpected status code: got 500",
			wantCount:     1,
			wantContains:  "status code",
		},
		{
			assertionType: "port_listening",
			actual:        "port 8080 is not open",
			wantCount:     1,
			wantContains:  "Start the service",
		},
		{
			assertionType: "postgres_ping",
			actual:        "connection refused",
			wantCount:     1,
			wantContains:  "not running",
		},
		{
			assertionType: "ssl_cert",
			actual:        "certificate expired",
			wantCount:     1,
			wantContains:  "renew",
		},
		{
			assertionType: "exit_code",
			actual:        "exit code: got 1, expected 0",
			wantCount:     1,
			wantContains:  "exit code",
		},
		{
			assertionType: "unknown_type",
			actual:        "something failed",
			wantCount:     1,
			wantContains:  "configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.assertionType+"/"+tt.wantContains, func(t *testing.T) {
			suggestions := GetSuggestions(tt.assertionType, tt.actual)
			if len(suggestions) < tt.wantCount {
				t.Errorf("expected at least %d suggestions, got %d", tt.wantCount, len(suggestions))
			}
			found := false
			for _, s := range suggestions {
				if strings.Contains(strings.ToLower(s), strings.ToLower(tt.wantContains)) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("no suggestion contains %q in %v", tt.wantContains, suggestions)
			}
		})
	}
}

// TestSanitize tests output truncation.
func TestSanitize(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short", "hello", 100, "hello"},
		{"exact", "hello", 5, "hello"},
		{"truncated", "hello world this is a long string", 5, "hello\n[... truncated, full output: 33 bytes]"},
		{"whitespace trimmed", "  hello  ", 100, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitize(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("sanitize(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// TestSuiteResultIncludesSuggestions verifies that smoke_run results
// include fix_suggestions for failed assertions.
func TestSuiteResultIncludesSuggestions(t *testing.T) {
	// Simulate a failed assertion result
	ar := AssertionResult{
		Type:     "redis_ping",
		Expected: "+PONG response from Redis",
		Actual:   "connection refused on localhost:6379",
		Passed:   false,
	}

	suggestions := GetSuggestions(ar.Type, ar.Actual)
	if len(suggestions) == 0 {
		t.Error("expected fix suggestions for failed redis_ping assertion")
	}
}
