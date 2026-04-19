---
completed: "2026-04-18"
created: "2026-04-18"
goals_completed: 38
goals_total: 38
status: COMPLETED
title: cosmo-smoke v0.7 — Implementation Plan
---

# cosmo-smoke v0.7 — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add monorepo sub-config discovery, WebSocket assertion, and optional gRPC build tag to cosmo-smoke.

**Architecture:** Three independent features. Monorepo adds a new `internal/monorepo/` package for discovery + runner integration. WebSocket adds stdlib-only client + assertion. gRPC splits existing code into build-tag-gated files. All follow existing pattern: schema struct → check function → runner wiring.

**Tech Stack:** Go 1.26, Cobra, yaml.v3, stdlib net/http + crypto/sha1 + encoding/binary for WebSocket

**Design spec:** `docs/brainstorming/2026-04-18-v0-7-monorepo-websocket-grpc-buildtag.md`

---

## File Structure

```
internal/schema/schema.go       # WebSocketCheck struct, Settings monorepo fields
internal/schema/validate.go     # WebSocket + monorepo validation rules
internal/runner/assertion.go    # Remove gRPC code, add CheckWebSocket
internal/runner/assertion_ws.go # WebSocket client implementation
internal/runner/assertion_grpc.go       # //go:build grpc — gRPC code moved from assertion.go
internal/runner/assertion_grpc_stub.go  # //go:build !grpc — stub
internal/runner/runner.go       # RunMonorepo method, WebSocket wiring
internal/runner/assertion_ws_test.go    # WebSocket tests
internal/runner/assertion_grpc_test.go  # gRPC tests (moved, +build grpc)
internal/monorepo/discover.go   # SubConfig discovery
internal/monorepo/discover_test.go
cmd/run.go                     # --monorepo flag
CLAUDE.md                      # Update assertion table
```

---

## Chunk 1: WebSocket Assertion

### Task 1: Add WebSocket schema struct + validation

**Files:**
- Modify: `internal/schema/schema.go`
- Modify: `internal/schema/validate.go`

- [ ] **Step 1: Add WebSocketCheck struct to schema.go (after VersionCheck struct)**

```go
// WebSocketCheck verifies a WebSocket endpoint is reachable and responds as expected.
type WebSocketCheck struct {
	URL            string   `yaml:"url"`
	Send           string   `yaml:"send,omitempty"`
	ExpectContains string   `yaml:"expect_contains,omitempty"`
	ExpectMatches  string   `yaml:"expect_matches,omitempty"`
	Timeout        Duration `yaml:"timeout,omitempty"`
}
```

- [ ] **Step 2: Add WebSocket field to Expect struct**

```go
WebSocket *WebSocketCheck `yaml:"websocket,omitempty"`
```

- [ ] **Step 3: Add validation rules to validate.go (after VersionCheck block)**

```go
if e := t.Expect.WebSocket; e != nil {
	if !strings.HasPrefix(e.URL, "ws://") && !strings.HasPrefix(e.URL, "wss://") {
		errs = append(errs, fmt.Sprintf("%s: websocket.url must start with ws:// or wss://", prefix))
	}
	if e.ExpectMatches != "" {
		if _, err := regexp.Compile(e.ExpectMatches); err != nil {
			errs = append(errs, fmt.Sprintf("%s: websocket.expect_matches is invalid regex: %v", prefix, err))
		}
	}
}
```

- [ ] **Step 4: Build and verify**

Run: `go build ./...`
Expected: success

---

### Task 2: Implement WebSocket client + Check function

**Files:**
- Create: `internal/runner/assertion_ws.go`
- Modify: `internal/runner/runner.go`

- [ ] **Step 1: Create assertion_ws.go with stdlib WebSocket client**

```go
package runner

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

// websocketGUID is the magic GUID from RFC 6455 used for accept key computation.
const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// wsUpgrade performs a WebSocket handshake over a raw TCP connection.
func wsUpgrade(conn net.Conn, host, path string, timeout time.Duration) error {
	key := make([]byte, 16)
	rand.Read(key)
	clientKey := base64.StdEncoding.EncodeToString(key)

	upgradeReq := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n", path, host, clientKey)
	conn.SetDeadline(time.Now().Add(timeout))
	conn.Write([]byte(upgradeReq))

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("read upgrade response: %w", err)
	}

	resp := string(buf[:n])
	if !strings.Contains(resp, " 101 ") {
		return fmt.Errorf("upgrade failed: %s", strings.Split(resp, "\r\n")[0])
	}

	// Verify accept key
	expectedAccept := computeAcceptKey(clientKey)
	if !strings.Contains(resp, expectedAccept) {
		return fmt.Errorf("invalid accept key")
	}

	return nil
}

func computeAcceptKey(clientKey string) string {
	h := sha1.New()
	h.Write([]byte(clientKey + websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// wsReadFrame reads a single WebSocket frame and returns the payload as a string.
// Handles text (opcode 1), binary (opcode 2), close (opcode 8), and ping (opcode 9).
func wsReadFrame(conn net.Conn) (string, bool, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", false, fmt.Errorf("read frame header: %w", err)
	}

	opcode := header[0] & 0x0F
	masked := (header[1] & 0x80) != 0
	payloadLen := int64(header[1] & 0x7F)

	switch payloadLen {
	case 126:
		ext := make([]byte, 2)
		if _, err := io.ReadFull(conn, ext); err != nil {
			return "", false, err
		}
		payloadLen = int64(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err := io.ReadFull(conn, ext); err != nil {
			return "", false, err
		}
		payloadLen = int64(binary.BigEndian.Uint64(ext))
	}

	// Skip mask key if present (server frames should not be masked, but handle it)
	if masked {
		mask := make([]byte, 4)
		if _, err := io.ReadFull(conn, mask); err != nil {
			return "", false, err
		}
	}

	if opcode == 0x08 {
		// Close frame
		reason := ""
		if payloadLen > 0 {
			payload := make([]byte, payloadLen)
			if _, err := io.ReadFull(conn, payload); err != nil {
				return "", true, err
			}
			reason = string(payload)
		}
		return reason, true, nil
	}

	if opcode == 0x09 {
		// Ping — read payload, ignore (no pong response in this minimal client)
		if payloadLen > 0 {
			payload := make([]byte, payloadLen)
			if _, err := io.ReadFull(conn, payload); err != nil {
				return "", false, err
			}
		}
		return "", false, nil
	}

	// Text (1) or Binary (2) — read payload
	var payload []byte
	if payloadLen > 0 {
		payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(conn, payload); err != nil {
			return "", false, err
		}
	}

	return string(payload), false, nil
}

// wsSendMessage writes a masked text frame (required by RFC 6455 for client frames).
func wsSendMessage(conn net.Conn, msg string) error {
	frame := []byte{0x81} // FIN + text opcode
	len := len(msg)
	if len <= 125 {
		frame = append(frame, byte(0x80|len))
	} else if len <= 65535 {
		frame = append(frame, 0x80|126)
		frame = append(frame, byte(len>>8), byte(len))
	} else {
		frame = append(frame, 0x80|127)
		for i := 7; i >= 0; i-- {
			frame = append(frame, byte(len>>(i*8)))
		}
	}

	mask := make([]byte, 4)
	rand.Read(mask)
	frame = append(frame, mask...)

	masked := make([]byte, len)
	for i, b := range []byte(msg) {
		masked[i] = b ^ mask[i%4]
	}
	frame = append(frame, masked...)

	_, err := conn.Write(frame)
	return err
}

// CheckWebSocket verifies a WebSocket endpoint is reachable and optionally matches response.
func CheckWebSocket(check *schema.WebSocketCheck) AssertionResult {
	timeout := check.Timeout.Duration
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	// Parse URL to get host and path
	url := check.URL
	useTLS := strings.HasPrefix(url, "wss://")
	hostPath := strings.TrimPrefix(url, "ws://")
	hostPath = strings.TrimPrefix(hostPath, "wss://")
	parts := strings.SplitN(hostPath, "/", 2)
	host := parts[0]
	path := "/"
	if len(parts) == 2 {
		path = "/" + parts[1]
	}

	// Add default port if not specified
	if !strings.Contains(host, ":") {
		if useTLS {
			host = host + ":443"
		} else {
			host = host + ":80"
		}
	}

	// Connect
	start := time.Now()
	var conn net.Conn
	var err error
	if useTLS {
		conn, err = tlsDialWithTimeout(host, timeout)
	} else {
		conn, err = net.DialTimeout("tcp", host, timeout)
	}
	if err != nil {
		return AssertionResult{
			Type:     "websocket",
			Expected: fmt.Sprintf("%s reachable", check.URL),
			Actual:   fmt.Sprintf("connection failed: %v", err),
			Passed:   false,
		}
	}
	defer conn.Close()
	elapsed := time.Since(start)

	// Upgrade
	if err := wsUpgrade(conn, host, path, timeout); err != nil {
		return AssertionResult{
			Type:     "websocket",
			Expected: fmt.Sprintf("%s upgrade", check.URL),
			Actual:   fmt.Sprintf("upgrade failed: %v", err),
			Passed:   false,
		}
	}

	// Connect-only mode: no send and no expectations
	if check.Send == "" && check.ExpectContains == "" && check.ExpectMatches == "" {
		return AssertionResult{
			Type:     "websocket",
			Expected: fmt.Sprintf("%s connected", check.URL),
			Actual:   fmt.Sprintf("connected (%s)", elapsed.Round(time.Millisecond)),
			Passed:   true,
		}
	}

	// Send message if provided
	if check.Send != "" {
		if err := wsSendMessage(conn, check.Send); err != nil {
			return AssertionResult{
				Type:     "websocket",
				Expected: "send message",
				Actual:   fmt.Sprintf("send failed: %v", err),
				Passed:   false,
			}
		}
	}

	// Read response and match
	conn.SetDeadline(time.Now().Add(timeout))
	for {
		msg, closed, err := wsReadFrame(conn)
		if err != nil {
			return AssertionResult{
				Type:     "websocket",
				Expected: "receive message",
				Actual:   fmt.Sprintf("read failed: %v", err),
				Passed:   false,
			}
		}
		if closed {
			return AssertionResult{
				Type:     "websocket",
				Expected: "receive message",
				Actual:   fmt.Sprintf("server closed: %s", msg),
				Passed:   false,
			}
		}
		// Skip empty frames (e.g. pong responses)
		if msg == "" {
			continue
		}

		if check.ExpectContains != "" {
			if strings.Contains(msg, check.ExpectContains) {
				return AssertionResult{
					Type:     "websocket",
					Expected: fmt.Sprintf("contains %q", check.ExpectContains),
					Actual:   msg,
					Passed:   true,
				}
			}
			return AssertionResult{
				Type:     "websocket",
				Expected: fmt.Sprintf("contains %q", check.ExpectContains),
				Actual:   fmt.Sprintf("received %q did not contain", msg),
				Passed:   false,
			}
		}

		if check.ExpectMatches != "" {
			matched, _ := regexp.MatchString(check.ExpectMatches, msg)
			if matched {
				return AssertionResult{
					Type:     "websocket",
					Expected: fmt.Sprintf("matches %q", check.ExpectMatches),
					Actual:   msg,
					Passed:   true,
				}
			}
			return AssertionResult{
				Type:     "websocket",
				Expected: fmt.Sprintf("matches %q", check.ExpectMatches),
				Actual:   fmt.Sprintf("received %q did not match", msg),
				Passed:   false,
			}
		}
	}
}

func tlsDialWithTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	return tls.DialWithDialer(dialer, "tcp", addr, nil)
}
```

Note: needs `"crypto/tls"` and `"regexp"` in imports.

- [ ] **Step 2: Wire WebSocket into runner.go (after VersionCheck block)**

```go
if t.Expect.WebSocket != nil {
	a := CheckWebSocket(t.Expect.WebSocket)
	assertions = append(assertions, a)
	if !a.Passed {
		allPassed = false
	}
}
```

- [ ] **Step 3: Build and verify**

Run: `go build ./...`
Expected: success

---

### Task 3: Write WebSocket tests

**Files:**
- Create: `internal/runner/assertion_ws_test.go`

- [ ] **Step 1: Create test file with stdlib WS test server**

```go
package runner

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

const testWSGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// wsTestServer creates an httptest server that upgrades to WebSocket and echoes messages.
func wsTestServer(handler func(conn net.Conn, msg string) string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Sec-WebSocket-Key")
		acceptKey := computeTestAcceptKey(key)
		hj, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, _, _ := hj.Hijack()
		resp := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
		conn.Write([]byte(resp))

		defer conn.Close()
		for {
			msg, closed, err := wsReadFrame(conn)
			if err != nil || closed {
				return
			}
			if msg == "" {
				continue
			}
			reply := handler(conn, msg)
			if reply != "" {
				wsWriteTextFrame(conn, reply)
			}
		}
	}))
}

func computeTestAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + testWSGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func wsWriteTextFrame(conn net.Conn, msg string) {
	frame := []byte{0x81}
	len := len(msg)
	if len <= 125 {
		frame = append(frame, byte(len))
	} else if len <= 65535 {
		frame = append(frame, 126, byte(len>>8), byte(len))
	}
	frame = append(frame, []byte(msg)...)
	conn.Write(frame)
}

func TestCheckWebSocket_ExpectContains_Pass(t *testing.T) {
	ts := wsTestServer(func(conn net.Conn, msg string) string {
		return "pong:" + msg
	})
	defer ts.Close()

	wsURL := "ws://" + strings.TrimPrefix(ts.URL, "http://")
	result := CheckWebSocket(&schema.WebSocketCheck{
		URL:            wsURL,
		Send:           "ping",
		ExpectContains: "pong",
		Timeout:        schema.Duration{Duration: 5 * time.Second},
	})
	if !result.Passed {
		t.Errorf("expected pass, got: %s", result.Actual)
	}
}

func TestCheckWebSocket_ExpectMatches_Pass(t *testing.T) {
	ts := wsTestServer(func(conn net.Conn, msg string) string {
		return `{"status":"connected","id":42}`
	})
	defer ts.Close()

	wsURL := "ws://" + strings.TrimPrefix(ts.URL, "http://")
	result := CheckWebSocket(&schema.WebSocketCheck{
		URL:           wsURL,
		Send:          "hello",
		ExpectMatches: `connected.*42`,
		Timeout:       schema.Duration{Duration: 5 * time.Second},
	})
	if !result.Passed {
		t.Errorf("expected pass, got: %s", result.Actual)
	}
}

func TestCheckWebSocket_NoMatch_Fail(t *testing.T) {
	ts := wsTestServer(func(conn net.Conn, msg string) string {
		return "hello world"
	})
	defer ts.Close()

	wsURL := "ws://" + strings.TrimPrefix(ts.URL, "http://")
	result := CheckWebSocket(&schema.WebSocketCheck{
		URL:            wsURL,
		Send:           "ping",
		ExpectContains: "pong",
		Timeout:        schema.Duration{Duration: 5 * time.Second},
	})
	if result.Passed {
		t.Error("expected fail")
	}
}

func TestCheckWebSocket_ConnectionRefused(t *testing.T) {
	result := CheckWebSocket(&schema.WebSocketCheck{
		URL:            "ws://127.0.0.1:1",
		ExpectContains: "anything",
		Timeout:        schema.Duration{Duration: 1 * time.Second},
	})
	if result.Passed {
		t.Error("expected fail for connection refused")
	}
}

func TestCheckWebSocket_ConnectOnly(t *testing.T) {
	ts := wsTestServer(func(conn net.Conn, msg string) string {
		return ""
	})
	defer ts.Close()

	wsURL := "ws://" + strings.TrimPrefix(ts.URL, "http://")
	result := CheckWebSocket(&schema.WebSocketCheck{
		URL:     wsURL,
		Timeout: schema.Duration{Duration: 5 * time.Second},
	})
	if !result.Passed {
		t.Errorf("expected pass for connect-only, got: %s", result.Actual)
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/runner/... -run TestCheckWebSocket -v`
Expected: all pass

---

## Chunk 2: Optional gRPC Build Tag

### Task 4: Split gRPC code into build-tagged files

**Files:**
- Create: `internal/runner/assertion_grpc.go`
- Create: `internal/runner/assertion_grpc_stub.go`
- Modify: `internal/runner/assertion.go`

- [ ] **Step 1: Create assertion_grpc.go with `//go:build grpc` tag**

Copy the following from `assertion.go` into `assertion_grpc.go`:
- Build tag: `//go:build grpc`
- Package declaration
- Imports: `context`, `fmt`, `time`, `schema`, `grpc`, `credentials`, `credentials/insecure`, `healthpb`
- The `CheckGRPCHealth` function (lines 582-626 of current assertion.go)

- [ ] **Step 2: Create assertion_grpc_stub.go with `//go:build !grpc` tag**

```go
//go:build !grpc

package runner

import (
	"fmt"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

// CheckGRPCHealth returns an error when built without the grpc tag.
func CheckGRPCHealth(check *schema.GRPCHealthCheck) AssertionResult {
	return AssertionResult{
		Type:     "grpc_health",
		Expected: check.Address,
		Actual:   "grpc_health not available — rebuild with -tags grpc",
		Passed:   false,
	}
}
```

- [ ] **Step 3: Remove gRPC code from assertion.go**

Remove from `assertion.go`:
- gRPC imports: `google.golang.org/grpc`, `google.golang.org/grpc/credentials`, `google.golang.org/grpc/credentials/insecure`, `healthpb`
- The `CheckGRPCHealth` function body
- The `isDockerAvailable` function stays (used by Docker checks)

- [ ] **Step 4: Build without grpc tag (default)**

Run: `go build -o /tmp/smoke-nogrpc . && ls -lh /tmp/smoke-nogrpc`
Expected: builds successfully, binary ~8MB

- [ ] **Step 5: Build with grpc tag**

Run: `go build -tags grpc -o /tmp/smoke-grpc . && ls -lh /tmp/smoke-grpc`
Expected: builds successfully, binary ~13.9MB

---

### Task 5: Move gRPC tests and add stub test

**Files:**
- Create: `internal/runner/assertion_grpc_test.go` (with `//go:build grpc`)
- Create: `internal/runner/assertion_grpc_stub_test.go` (with `//go:build !grpc`)

- [ ] **Step 1: Find and move existing gRPC tests**

Search for gRPC-related test functions in `internal/runner/`. If found, move them to `assertion_grpc_test.go` with the `//go:build grpc` tag.

- [ ] **Step 2: Create stub test**

```go
//go:build !grpc

package runner

import (
	"strings"
	"testing"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

func TestCheckGRPCHealth_StubReturns(t *testing.T) {
	result := CheckGRPCHealth(&schema.GRPCHealthCheck{Address: "localhost:9090"})
	if result.Passed {
		t.Error("stub should not pass")
	}
	if !strings.Contains(result.Actual, "grpc") {
		t.Error("should mention grpc in output, got:", result.Actual)
	}
}
```

- [ ] **Step 3: Run tests both ways**

Run: `go test ./internal/runner/... -v`
Expected: all pass (stub test runs, gRPC tests skipped)

Run: `go test -tags grpc ./internal/runner/... -v`
Expected: all pass (gRPC tests run, stub test skipped)

---

## Chunk 3: Monorepo Sub-Config Discovery

### Task 6: Create monorepo discovery package

**Files:**
- Create: `internal/monorepo/discover.go`
- Create: `internal/monorepo/discover_test.go`

- [ ] **Step 1: Create discover.go**

```go
package monorepo

import (
	"os"
	"path/filepath"
	"strings"
)

// SubConfig represents a discovered .smoke.yaml file.
type SubConfig struct {
	Path    string // absolute path to .smoke.yaml
	Dir     string // directory containing it
	Project string // directory name as fallback project name
}

// defaultSkipDirs are always excluded from discovery.
var defaultSkipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true,
	"__pycache__": true, "dist": true, "build": true,
	"target": true, ".next": true, ".cache": true,
}

// Discover walks root and finds all .smoke.yaml files in subdirectories.
// Returns discovered configs sorted by path. Does not include root's own .smoke.yaml.
func Discover(root string, exclude []string) ([]SubConfig, error) {
	excludeSet := make(map[string]bool)
	for _, d := range exclude {
		excludeSet[filepath.Clean(d)] = true
	}

	var configs []SubConfig
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors, keep walking
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()

		// Skip default dirs
		if defaultSkipDirs[name] {
			return filepath.SkipDir
		}
		// Skip user-excluded dirs (match relative path)
		rel, _ := filepath.Rel(root, path)
		if excludeSet[filepath.Clean(rel)] {
			return filepath.SkipDir
		}
		// Skip root dir itself
		if path == root {
			return nil
		}

		configPath := filepath.Join(path, ".smoke.yaml")
		if _, err := os.Stat(configPath); err == nil {
			configs = append(configs, SubConfig{
				Path:    configPath,
				Dir:     path,
				Project: name,
			})
		}
		return nil
	})
	return configs, err
}
```

- [ ] **Step 2: Create discover_test.go**

```go
package monorepo

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestDiscover_FindsSubConfigs(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "api"), 0755)
	os.MkdirAll(filepath.Join(root, "worker"), 0755)
	os.WriteFile(filepath.Join(root, "api", ".smoke.yaml"), []byte("version: 1\nproject: api\ntests: []\n"), 0644)
	os.WriteFile(filepath.Join(root, "worker", ".smoke.yaml"), []byte("version: 1\nproject: worker\ntests: []\n"), 0644)

	configs, err := Discover(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(configs))
	}
	names := []string{filepath.Base(configs[0].Dir), filepath.Base(configs[1].Dir)}
	sort.Strings(names)
	if names[0] != "api" || names[1] != "worker" {
		t.Errorf("expected api+worker, got %v", names)
	}
}

func TestDiscover_SkipsIgnoredDirs(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules", "pkg"), 0755)
	os.WriteFile(filepath.Join(root, "node_modules", "pkg", ".smoke.yaml"), []byte("version: 1\nproject: pkg\ntests: []\n"), 0644)

	configs, err := Discover(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 0 {
		t.Errorf("expected 0 configs (node_modules skipped), got %d", len(configs))
	}
}

func TestDiscover_CustomExclude(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "api"), 0755)
	os.MkdirAll(filepath.Join(root, "internal"), 0755)
	os.WriteFile(filepath.Join(root, "api", ".smoke.yaml"), []byte("version: 1\nproject: api\ntests: []\n"), 0644)
	os.WriteFile(filepath.Join(root, "internal", ".smoke.yaml"), []byte("version: 1\nproject: internal\ntests: []\n"), 0644)

	configs, err := Discover(root, []string{"internal"})
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 1 || filepath.Base(configs[0].Dir) != "api" {
		t.Errorf("expected 1 config (api only), got %v", configs)
	}
}

func TestDiscover_DeepNesting(t *testing.T) {
	root := t.TempDir()
	deepDir := filepath.Join(root, "services", "team-a", "api")
	os.MkdirAll(deepDir, 0755)
	os.WriteFile(filepath.Join(deepDir, ".smoke.yaml"), []byte("version: 1\nproject: deep\ntests: []\n"), 0644)

	configs, err := Discover(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 1 || filepath.Base(configs[0].Dir) != "api" {
		t.Errorf("expected 1 deep config, got %v", configs)
	}
}

func TestDiscover_NoSmokeFiles(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "api"), 0755)
	// No .smoke.yaml anywhere

	configs, err := Discover(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 0 {
		t.Errorf("expected 0 configs, got %d", len(configs))
	}
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/monorepo/... -v`
Expected: all pass

---

### Task 7: Add monorepo schema + CLI flag + runner integration

**Files:**
- Modify: `internal/schema/schema.go`
- Modify: `internal/runner/runner.go`
- Modify: `cmd/run.go`

- [ ] **Step 1: Add monorepo fields to Settings in schema.go**

```go
Monorepo        bool     `yaml:"monorepo,omitempty"`
MonorepoExclude []string `yaml:"monorepo_exclude,omitempty"`
```

- [ ] **Step 2: Add --monorepo flag to cmd/run.go**

Add to var block:
```go
monorepo bool
```

Add to init():
```go
runCmd.Flags().BoolVar(&monorepo, "monorepo", false, "Auto-discover .smoke.yaml in subdirectories")
```

- [ ] **Step 3: Add RunMonorepo to runner.go**

```go
func (r *Runner) RunMonorepo(opts RunOptions, subConfigs []monorepo.SubConfig) (*SuiteResult, error) {
	start := time.Now()

	suite := &SuiteResult{
		Project: r.Config.Project,
		Total:   len(subConfigs),
	}

	for _, sc := range subConfigs {
		cfg, err := schema.Load(sc.Path)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", sc.Path, err)
		}
		subRunner := &Runner{
			Config:    cfg,
			Reporter:  r.Reporter,
			ConfigDir: sc.Dir,
		}
		result, err := subRunner.Run(opts)
		if err != nil {
			return nil, fmt.Errorf("running %s: %w", sc.Project, err)
		}
		suite.Tests = append(suite.Tests, result.Tests...)
		suite.Passed += result.Passed
		suite.Failed += result.Failed
		suite.Skipped += result.Skipped
		suite.AllowedFailures += result.AllowedFailures
		suite.Total += result.Total
	}

	suite.Duration = time.Since(start)
	r.Reporter.Summary(reporter.SuiteResultData{
		Project:         suite.Project,
		Total:           suite.Total,
		Passed:          suite.Passed,
		Failed:          suite.Failed,
		Skipped:         suite.Skipped,
		AllowedFailures: suite.AllowedFailures,
		Duration:        suite.Duration,
	})
	return suite, nil
}
```

- [ ] **Step 4: Wire monorepo mode in cmd/run.go runSmoke function**

After config validation, before creating the runner:
```go
// Check monorepo mode
if monorepo || cfg.Settings.Monorepo {
	configs, err := monorepo.Discover(configDir, cfg.Settings.MonorepoExclude)
	if err != nil {
		return fmt.Errorf("discovering sub-configs: %w", err)
	}
	if len(configs) == 0 {
		return fmt.Errorf("no smoke configs found in %s", configDir)
	}
	r := &runner.Runner{Config: cfg, Reporter: rep, ConfigDir: configDir}
	result, err := r.RunMonorepo(runner.RunOptions{...}, configs)
	// handle result
}
```

- [ ] **Step 5: Build and run full test suite**

Run: `go build ./... && go test ./...`
Expected: success, all tests pass

---

## Chunk 4: Wiring + Release

### Task 8: Update CLAUDE.md + full test suite

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update assertion table with WebSocket**

Add to the Network Assertions table in CLAUDE.md:
```
| WebSocket | `{url, send?, expect_contains?, expect_matches?, timeout?}` | WebSocket connect-send-expect assertion |
```

- [ ] **Step 2: Add monorepo to commands section**

Update `smoke run` flags in CLAUDE.md:
```
      --monorepo            Auto-discover .smoke.yaml in subdirectories
```

- [ ] **Step 3: Update test count**

Update test count in CLAUDE.md to reflect new total.

- [ ] **Step 4: Run full suite**

Run: `go test ./...`
Expected: all pass (~260+ tests)

- [ ] **Step 5: Run self-smoke**

Run: `go build -ldflags "-s -w -X github.com/CosmoLabs-org/cosmo-smoke/cmd.Version=0.7.0" -o smoke . && ./smoke run`
Expected: 6 tests pass

---

## Release Checklist

- [ ] Run `smoke run` self-smoke
- [ ] Run `go test ./...` (all pass)
- [ ] Update CLAUDE.md assertion table + test count
- [ ] Update README.md, USAGE.md, FEATURES.md with v0.7 features
- [ ] Changelog: `ccs changelog finalize v0.7.0 "monorepo-websocket-grpc-buildtag"`
- [ ] Version: `ccs version-track bump minor`
- [ ] Tag: `git tag v0.7.0`
- [ ] Push: `git push origin master --tags`
