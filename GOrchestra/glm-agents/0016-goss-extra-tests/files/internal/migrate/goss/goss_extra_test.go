package goss

import (
	"testing"
)

// TestParseAllAssertionTypes verifies a Gossfile with every assertion type parses correctly.
func TestParseAllAssertionTypes(t *testing.T) {
	yaml := `
process:
  nginx:
    running: true
port:
  tcp:80:
    listening: true
command:
  "echo ok":
    exit-status: 0
file:
  /etc/hosts:
    exists: true
http:
  "http://localhost/health":
    status: 200
package:
  curl:
    installed: true
service:
  sshd:
    running: true
    enabled: true
user:
  root:
    exists: true
group:
  root:
    exists: true
dns:
  localhost:
    resolvable: true
addr:
  "tcp://127.0.0.1:443":
    reachable: true
interface:
  eth0:
    exists: true
mount:
  "/tmp":
    exists: true
kernel-param:
  net.ipv4.ip_forward:
    value: "1"
gossfile:
  "extra.yaml": {}
`
	gf, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	want := map[string]int{
		"process": 1, "port": 1, "command": 1, "file": 1,
		"http": 1, "package": 1, "service": 1, "user": 1,
		"group": 1, "dns": 1, "addr": 1, "interface": 1,
		"mount": 1, "kernel-param": 1, "gossfile": 1,
	}

	got := map[string]int{
		"process":      len(gf.Process),
		"port":         len(gf.Port),
		"command":      len(gf.Command),
		"file":         len(gf.File),
		"http":         len(gf.HTTP),
		"package":      len(gf.Package),
		"service":      len(gf.Service),
		"user":         len(gf.User),
		"group":        len(gf.Group),
		"dns":          len(gf.DNS),
		"addr":         len(gf.Addr),
		"interface":    len(gf.Interface),
		"mount":        len(gf.Mount),
		"kernel-param": len(gf.KernelParam),
		"gossfile":     len(gf.Gossfile),
	}

	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s count = %d, want %d", k, got[k], w)
		}
	}
}

// TestParseEmptyVarsSection parses a Gossfile where vars/params sections are empty.
func TestParseEmptyVarsSection(t *testing.T) {
	yaml := `
process:
  nginx:
    running: true
vars: {}
params: {}
`
	gf, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(gf.Process) != 1 {
		t.Errorf("Process count = %d, want 1", len(gf.Process))
	}
	if !boolVal(gf.Process["nginx"], "running") {
		t.Error("nginx.running should be true")
	}
}

// TestParseGossfileHTTPTests verifies parsing of HTTP test resources.
func TestParseGossfileHTTPTests(t *testing.T) {
	yaml := `
http:
  "https://example.com/health":
    status: 200
    body:
      - "ok"
      - "healthy"
    method: GET
  "https://example.com/api":
    status: 404
    method: POST
`
	gf, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(gf.HTTP) != 2 {
		t.Fatalf("HTTP count = %d, want 2", len(gf.HTTP))
	}

	// Verify first entry
	h1 := gf.HTTP["https://example.com/health"]
	if intVal(h1, "status") != 200 {
		t.Errorf("health status = %d, want 200", intVal(h1, "status"))
	}
	if stringVal(h1, "method") != "GET" {
		t.Errorf("health method = %q, want GET", stringVal(h1, "method"))
	}
	bodies := stringSlice(h1, "body")
	if len(bodies) != 2 || bodies[0] != "ok" || bodies[1] != "healthy" {
		t.Errorf("health body = %v, want [ok healthy]", bodies)
	}

	// Verify second entry
	h2 := gf.HTTP["https://example.com/api"]
	if intVal(h2, "status") != 404 {
		t.Errorf("api status = %d, want 404", intVal(h2, "status"))
	}
}

// TestParseGossfileProcessTests verifies parsing of process resources.
func TestParseGossfileProcessTests(t *testing.T) {
	yaml := `
process:
  nginx:
    running: true
  sshd:
    running: true
  cron:
    running: false
`
	gf, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(gf.Process) != 3 {
		t.Fatalf("Process count = %d, want 3", len(gf.Process))
	}
	if !boolVal(gf.Process["nginx"], "running") {
		t.Error("nginx.running should be true")
	}
	if !boolVal(gf.Process["sshd"], "running") {
		t.Error("sshd.running should be true")
	}
	if boolVal(gf.Process["cron"], "running") {
		t.Error("cron.running should be false")
	}
}

// TestConvertGossHTTPToSmokeAssertion translates a Goss HTTP test to a smoke HTTP assertion.
func TestConvertGossHTTPToSmokeAssertion(t *testing.T) {
	gf := &GossFile{
		HTTP: map[string]GossAttrs{
			"http://app.local/status": {
				"status": 200,
				"body":   []interface{}{"ok"},
				"method": "GET",
			},
		},
	}

	tests, warnings := Translate(gf, TranslateOptions{})
	if len(tests) != 1 {
		t.Fatalf("tests = %d, want 1", len(tests))
	}

	tt := tests[0]
	if tt.Expect.HTTP == nil {
		t.Fatal("expected HTTP check in assertion")
	}

	if tt.Expect.HTTP.URL != "http://app.local/status" {
		t.Errorf("URL = %q, want %q", tt.Expect.HTTP.URL, "http://app.local/status")
	}
	if tt.Expect.HTTP.StatusCode == nil || *tt.Expect.HTTP.StatusCode != 200 {
		t.Error("StatusCode should be 200")
	}
	if tt.Expect.HTTP.BodyContains != "ok" {
		t.Errorf("BodyContains = %q, want %q", tt.Expect.HTTP.BodyContains, "ok")
	}
	if tt.Expect.HTTP.Method != "GET" {
		t.Errorf("Method = %q, want %q", tt.Expect.HTTP.Method, "GET")
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %d", len(warnings))
	}
}

// TestConvertGossProcessToProcessRunning translates a Goss process test to process_running assertion.
func TestConvertGossProcessToProcessRunning(t *testing.T) {
	gf := &GossFile{
		Process: map[string]GossAttrs{
			"redis-server": {"running": true},
			"stopped-svc":  {"running": false},
		},
	}

	tests, _ := Translate(gf, TranslateOptions{})
	if len(tests) != 1 {
		t.Fatalf("tests = %d, want 1 (only running:true)", len(tests))
	}

	tt := tests[0]
	if tt.Expect.ProcessRunning != "redis-server" {
		t.Errorf("ProcessRunning = %q, want %q", tt.Expect.ProcessRunning, "redis-server")
	}
	if tt.Name != "process:redis-server running" {
		t.Errorf("Name = %q, want %q", tt.Name, "process:redis-server running")
	}
	if tt.Run != "true" {
		t.Errorf("Run = %q, want %q", tt.Run, "true")
	}
	if tt.Expect.ExitCode == nil || *tt.Expect.ExitCode != 0 {
		t.Error("ExitCode should be 0")
	}
}
