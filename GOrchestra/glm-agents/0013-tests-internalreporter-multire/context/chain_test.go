package reporter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChain_SingleFormat_NoFilesCreated(t *testing.T) {
	rep, closers, err := Chain("terminal", os.Stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep == nil {
		t.Fatal("expected non-nil reporter")
	}
	if len(closers) != 0 {
		t.Fatalf("expected 0 closers, got %d", len(closers))
	}
}

func TestChain_MultipleFormats_CreatesFiles(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	rep, closers, err := Chain("terminal,json", os.Stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep == nil {
		t.Fatal("expected non-nil reporter")
	}
	if len(closers) != 1 {
		t.Fatalf("expected 1 closer, got %d", len(closers))
	}
	if _, err := os.Stat(filepath.Join(tmp, "smoke-results.json")); err != nil {
		t.Fatalf("expected smoke-results.json to exist: %v", err)
	}
	for _, c := range closers {
		c.Close()
	}
}

func TestChain_DeduplicatesFormats(t *testing.T) {
	var buf bytes.Buffer
	rep, closers, err := Chain("json,json,json", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(closers) != 0 {
		t.Fatalf("expected 0 closers (single format after dedup), got %d", len(closers))
	}
	_ = rep
}

func TestChain_UnknownFormat_ReturnsError(t *testing.T) {
	_, _, err := Chain("xml", os.Stdout)
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Fatalf("error should mention unknown format: %v", err)
	}
}

func TestChain_EmptyFormat_ReturnsError(t *testing.T) {
	_, _, err := Chain("", os.Stdout)
	if err == nil {
		t.Fatal("expected error for empty format")
	}
}

func TestChain_CommasOnly_ReturnsError(t *testing.T) {
	_, _, err := Chain(",,,", os.Stdout)
	if err == nil {
		t.Fatal("expected error for commas-only format")
	}
}

func TestChain_CaseInsensitive(t *testing.T) {
	rep, closers, err := Chain("JSON", os.Stdout)
	if err != nil {
		t.Fatalf("case-insensitive match should work: %v", err)
	}
	if len(closers) != 0 {
		t.Fatalf("expected 0 closers for single format, got %d", len(closers))
	}
	_ = rep
}

func TestChain_WhitespaceTrimmed(t *testing.T) {
	var buf bytes.Buffer
	rep, closers, err := Chain(" json , terminal ", &buf)
	if err != nil {
		t.Fatalf("whitespace trimming should work: %v", err)
	}
	if len(closers) != 1 {
		t.Fatalf("expected 1 closer (terminal to file), got %d", len(closers))
	}
	_ = rep
	for _, c := range closers {
		c.Close()
	}
}

func TestChain_TrailingComma(t *testing.T) {
	rep, closers, err := Chain("json,", os.Stdout)
	if err != nil {
		t.Fatalf("trailing comma should be handled: %v", err)
	}
	if len(closers) != 0 {
		t.Fatalf("expected 0 closers, got %d", len(closers))
	}
	_ = rep
}

func TestChain_FileNaming(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	tests := []struct {
		format   string
		filename string
	}{
		{"json", "smoke-results.json"},
		{"junit", "smoke-junit.xml"},
		{"prometheus", "smoke-metrics.prom"},
		{"tap", "smoke-tap.txt"},
	}
	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			_, closers, err := Chain("terminal,"+tc.format, os.Stdout)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			path := filepath.Join(tmp, tc.filename)
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("expected %s to exist: %v", tc.filename, err)
			}
			for _, c := range closers {
				c.Close()
			}
		})
	}
}
