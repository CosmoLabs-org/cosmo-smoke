package mcp

import (
	"context"
	"strings"
	"testing"
)

// TestHandleSmokeInitAutoDetect tests config generation for the current project.
func TestHandleSmokeInitAutoDetect(t *testing.T) {
	srv := NewServer()
	handler := srv.Handler("smoke_init")

	result, err := handler(context.Background(), map[string]interface{}{
		"directory": "../..",
	})
	if err != nil {
		t.Fatalf("smoke_init returned error: %v", err)
	}

	ir, ok := result.(*InitResult)
	if !ok {
		t.Fatalf("expected *InitResult, got %T", result)
	}

	if ir.YAML == "" {
		t.Error("expected non-empty YAML output")
	}
	if ir.Written {
		t.Error("expected written=false when write not specified")
	}
	// Should contain config markers
	if !strings.Contains(ir.YAML, "version") || !strings.Contains(ir.YAML, "tests") {
		t.Errorf("generated YAML doesn't look like a smoke config: %s", ir.YAML[:min(200, len(ir.YAML))])
	}
}

// TestHandleSmokeGenerateTestHTTP tests generating a single HTTP test.
func TestHandleSmokeGenerateTestHTTP(t *testing.T) {
	srv := NewServer()
	handler := srv.Handler("smoke_generate_test")

	result, err := handler(context.Background(), map[string]interface{}{
		"name":           "health check",
		"assertion_type": "http",
		"params": map[string]interface{}{
			"url":         "http://localhost:8080/health",
			"status_code": float64(200),
		},
	})
	if err != nil {
		t.Fatalf("smoke_generate_test returned error: %v", err)
	}

	gr, ok := result.(*GenerateTestResult)
	if !ok {
		t.Fatalf("expected *GenerateTestResult, got %T", result)
	}

	if !strings.Contains(gr.YAML, "health check") {
		t.Error("generated YAML should contain test name")
	}
	if !strings.Contains(gr.YAML, "http") {
		t.Error("generated YAML should contain http assertion")
	}
	if !strings.Contains(gr.YAML, "localhost:8080") {
		t.Error("generated YAML should contain the URL")
	}
}

// TestHandleSmokeGenerateTestPort tests generating a port_listening test.
func TestHandleSmokeGenerateTestPort(t *testing.T) {
	srv := NewServer()
	handler := srv.Handler("smoke_generate_test")

	result, err := handler(context.Background(), map[string]interface{}{
		"name":           "web server port",
		"assertion_type": "port_listening",
		"params": map[string]interface{}{
			"port": float64(8080),
		},
	})
	if err != nil {
		t.Fatalf("smoke_generate_test returned error: %v", err)
	}

	gr := result.(*GenerateTestResult)
	if !strings.Contains(gr.YAML, "port_listening") {
		t.Error("generated YAML should contain port_listening")
	}
	if !strings.Contains(gr.YAML, "8080") {
		t.Error("generated YAML should contain port 8080")
	}
}

// TestHandleSmokeGenerateTestMissingName tests error for missing name.
func TestHandleSmokeGenerateTestMissingName(t *testing.T) {
	srv := NewServer()
	handler := srv.Handler("smoke_generate_test")

	_, err := handler(context.Background(), map[string]interface{}{
		"assertion_type": "http",
	})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

// TestHandleSmokeGenerateTestMissingType tests error for missing assertion type.
func TestHandleSmokeGenerateTestMissingType(t *testing.T) {
	srv := NewServer()
	handler := srv.Handler("smoke_generate_test")

	_, err := handler(context.Background(), map[string]interface{}{
		"name": "some test",
	})
	if err == nil {
		t.Error("expected error for missing assertion_type")
	}
}
