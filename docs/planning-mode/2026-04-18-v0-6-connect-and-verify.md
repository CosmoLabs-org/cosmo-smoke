# cosmo-smoke v0.6 — Connect and Verify — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add 4 new assertion types (url_reachable, service_reachable, s3_bucket, version_check) and a pre-commit hook file to cosmo-smoke.

**Architecture:** Follow existing assertion pattern: schema struct in `schema.go` → Check function in `assertion.go` → wire in `runner.go`. New `httpReachable` shared helper for HTTP-based checks. All stdlib, no new deps.

**Tech Stack:** Go 1.26, Cobra, yaml.v3, stdlib net/http + regexp

**Design spec:** `docs/brainstorming/2026-04-18-v0-6-connect-and-verify.md`

---

## File Structure

```
internal/schema/schema.go          # Add 4 new structs + 4 new Expect fields + validation
internal/runner/assertion.go       # Add httpReachable helper + 4 Check functions
internal/runner/runner.go          # Wire 4 new assertion checks in runTestOnce()
internal/runner/assertion_reachable_test.go  # Tests for url_reachable, service_reachable, s3_bucket
internal/runner/assertion_version_test.go    # Tests for version_check
.pre-commit-hooks.yaml             # Pre-commit framework hook definition
```

---

## Chunk 1: Schema + Shared Helper

### Task 1: Add schema structs and wire into Expect

**Files:**
- Modify: `internal/schema/schema.go`

- [ ] **Step 1: Add new check structs after existing ones (after line ~158)**

```go
// URLReachableCheck verifies an HTTP/HTTPS endpoint is accessible.
// Lightweight connectivity check — use the http assertion for full response validation.
type URLReachableCheck struct {
	URL        string   `yaml:"url"`
	Timeout    Duration `yaml:"timeout,omitempty"`
	StatusCode *int     `yaml:"status_code,omitempty"` // 0 = any 2xx
}

// ServiceReachableCheck verifies an external service dependency is accessible.
// Semantically named wrapper around url_reachable for dependency documentation.
type ServiceReachableCheck struct {
	URL     string   `yaml:"url"`
	Timeout Duration `yaml:"timeout,omitempty"`
}

// S3BucketCheck verifies an S3-compatible bucket is accessible via anonymous HEAD.
// For authenticated access, use the http assertion with Go template env var references.
type S3BucketCheck struct {
	Bucket   string `yaml:"bucket"`
	Region   string `yaml:"region,omitempty"`   // default us-east-1
	Endpoint string `yaml:"endpoint,omitempty"` // default s3.amazonaws.com
}

// VersionCheck verifies an installed tool matches a required version pattern.
// Runs the command and regex-matches stdout. Unix-only (uses sh -c).
type VersionCheck struct {
	Command string `yaml:"command"` // shell command to run
	Pattern string `yaml:"pattern"` // Go regex to match against stdout
}
```

- [ ] **Step 2: Add fields to Expect struct (after DockerImage field, ~line 78)**

```go
URLReachable     *URLReachableCheck     `yaml:"url_reachable,omitempty"`
ServiceReachable *ServiceReachableCheck `yaml:"service_reachable,omitempty"`
S3Bucket         *S3BucketCheck         `yaml:"s3_bucket,omitempty"`
VersionCheck     *VersionCheck          `yaml:"version_check,omitempty"`
```

- [ ] **Step 3: Add validation rules to Validate function (or create if missing)**

```go
// In the validation function, add checks:
if e.URLReachable != nil {
    if !strings.HasPrefix(e.URLReachable.URL, "http://") && !strings.HasPrefix(e.URLReachable.URL, "https://") {
        errs = append(errs, fmt.Errorf("url_reachable.url must start with http:// or https://"))
    }
}
if e.ServiceReachable != nil {
    if !strings.HasPrefix(e.ServiceReachable.URL, "http://") && !strings.HasPrefix(e.ServiceReachable.URL, "https://") {
        errs = append(errs, fmt.Errorf("service_reachable.url must start with http:// or https://"))
    }
}
if e.S3Bucket != nil {
    if e.S3Bucket.Bucket == "" {
        errs = append(errs, fmt.Errorf("s3_bucket.bucket is required"))
    }
}
if e.VersionCheck != nil {
    if e.VersionCheck.Command == "" {
        errs = append(errs, fmt.Errorf("version_check.command is required"))
    }
    if _, err := regexp.Compile(e.VersionCheck.Pattern); err != nil {
        errs = append(errs, fmt.Errorf("version_check.pattern is invalid regex: %w", err))
    }
}
```

- [ ] **Step 4: Build and verify compilation**

Run: `go build ./...`
Expected: success

---

### Task 2: Add httpReachable shared helper

**Files:**
- Modify: `internal/runner/assertion.go`

- [ ] **Step 1: Add httpReachable helper function (after existing Check functions)**

```go
// httpReachable makes an HTTP GET request and returns the status code,
// response time, and any error. expectedStatus=0 means any 2xx passes.
func httpReachable(url string, timeout time.Duration, expectedStatus int) (statusCode int, elapsed time.Duration, err error) {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	start := time.Now()
	resp, err := client.Get(url)
	elapsed = time.Since(start)
	if err != nil {
		return 0, elapsed, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, elapsed, nil
}

// is2xx returns true if the status code is in the 200-299 range.
func is2xx(code int) bool {
	return code >= 200 && code <= 299
}
```

- [ ] **Step 2: Add imports if not present**

Ensure `net/http` is in the import block.

- [ ] **Step 3: Build and verify**

Run: `go build ./...`
Expected: success

---

## Chunk 2: Assertion Implementations

### Task 3: Implement url_reachable + service_reachable + s3_bucket assertions

**Files:**
- Modify: `internal/runner/assertion.go`
- Create: `internal/runner/assertion_reachable_test.go`

- [ ] **Step 1: Write failing tests for url_reachable**

Create `internal/runner/assertion_reachable_test.go`:

```go
package runner

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

func TestCheckURLReachable_Pass(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	result := CheckURLReachable(&schema.URLReachableCheck{URL: ts.URL})
	if !result.Passed {
		t.Errorf("expected pass, got: %s", result.Actual)
	}
}

func TestCheckURLReachable_Fail5xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer ts.Close()

	result := CheckURLReachable(&schema.URLReachableCheck{URL: ts.URL})
	if result.Passed {
		t.Error("expected fail for 503")
	}
}

func TestCheckURLReachable_SpecificStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	}))
	defer ts.Close()

	code := 201
	result := CheckURLReachable(&schema.URLReachableCheck{URL: ts.URL, StatusCode: &code})
	if !result.Passed {
		t.Errorf("expected pass for 201, got: %s", result.Actual)
	}
}

func TestCheckURLReachable_InvalidURL(t *testing.T) {
	result := CheckURLReachable(&schema.URLReachableCheck{URL: "http://invalid.invalid.invalid"})
	if result.Passed {
		t.Error("expected fail for invalid URL")
	}
}

func TestCheckServiceReachable_Pass(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	result := CheckServiceReachable(&schema.ServiceReachableCheck{URL: ts.URL})
	if !result.Passed {
		t.Errorf("expected pass, got: %s", result.Actual)
	}
}

func TestCheckServiceReachable_Fail(t *testing.T) {
	result := CheckServiceReachable(&schema.ServiceReachableCheck{URL: "http://invalid.invalid.invalid", Timeout: schema.Duration{Duration: 1 * time.Second}})
	if result.Passed {
		t.Error("expected fail for unreachable service")
	}
}

func TestCheckS3Bucket_Pass(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate S3 HEAD bucket response
		w.WriteHeader(200)
	}))
	defer ts.Close()

	result := CheckS3Bucket(&schema.S3BucketCheck{
		Bucket:   "test-bucket",
		Endpoint: ts.URL[7:], // strip "http://"
	})
	if !result.Passed {
		t.Errorf("expected pass, got: %s", result.Actual)
	}
}

func TestCheckS3Bucket_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	result := CheckS3Bucket(&schema.S3BucketCheck{
		Bucket:   "missing-bucket",
		Endpoint: ts.URL[7:],
	})
	if result.Passed {
		t.Error("expected fail for 404")
	}
}

func TestCheckS3Bucket_AuthRequired(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	}))
	defer ts.Close()

	result := CheckS3Bucket(&schema.S3BucketCheck{
		Bucket:   "private-bucket",
		Endpoint: ts.URL[7:],
	})
	if result.Passed {
		t.Error("expected fail for 403")
	}
	if result.Expected == "" || !contains(result.Expected, "authentication") {
		t.Error("expected hint about authentication in output")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/runner/... -run "TestCheckURLReachable|TestCheckServiceReachable|TestCheckS3Bucket" -v`
Expected: compilation errors (functions don't exist yet)

- [ ] **Step 3: Implement Check functions in assertion.go**

```go
func CheckURLReachable(check *schema.URLReachableCheck) AssertionResult {
	statusCode, elapsed, err := httpReachable(check.URL, check.Timeout.Duration, 0)
	if err != nil {
		return AssertionResult{
			Type:     "url_reachable",
			Expected: fmt.Sprintf("%s reachable", check.URL),
			Actual:   fmt.Sprintf("connection failed: %v", err),
			Passed:   false,
		}
	}
	expected := "any 2xx"
	passed := is2xx(statusCode)
	if check.StatusCode != nil {
		expected = fmt.Sprintf("HTTP %d", *check.StatusCode)
		passed = statusCode == *check.StatusCode
	}
	return AssertionResult{
		Type:     "url_reachable",
		Expected: expected,
		Actual:   fmt.Sprintf("HTTP %d (%s)", statusCode, elapsed.Round(time.Millisecond)),
		Passed:   passed,
	}
}

func CheckServiceReachable(check *schema.ServiceReachableCheck) AssertionResult {
	statusCode, elapsed, err := httpReachable(check.URL, check.Timeout.Duration, 0)
	if err != nil {
		return AssertionResult{
			Type:     "service_reachable",
			Expected: fmt.Sprintf("%s reachable", check.URL),
			Actual:   fmt.Sprintf("connection failed: %v", err),
			Passed:   false,
		}
	}
	return AssertionResult{
		Type:     "service_reachable",
		Expected: "any 2xx",
		Actual:   fmt.Sprintf("HTTP %d (%s)", statusCode, elapsed.Round(time.Millisecond)),
		Passed:   is2xx(statusCode),
	}
}

func CheckS3Bucket(check *schema.S3BucketCheck) AssertionResult {
	region := check.Region
	if region == "" {
		region = "us-east-1"
	}
	endpoint := check.Endpoint
	if endpoint == "" {
		endpoint = "s3.amazonaws.com"
	}
	// Use path-style URL for anonymous HEAD
	url := fmt.Sprintf("https://%s/%s?location", endpoint, check.Bucket)

	statusCode, elapsed, err := httpReachable(url, 5*time.Second, 0)
	if err != nil {
		return AssertionResult{
			Type:     "s3_bucket",
			Expected: fmt.Sprintf("bucket %s accessible", check.Bucket),
			Actual:   fmt.Sprintf("connection failed: %v", err),
			Passed:   false,
		}
	}
	if statusCode == 403 {
		return AssertionResult{
			Type:     "s3_bucket",
			Expected: fmt.Sprintf("bucket %s accessible", check.Bucket),
			Actual:   "HTTP 403 Forbidden — bucket requires authentication; use http assertion with Go templates for authenticated access",
			Passed:   false,
		}
	}
	return AssertionResult{
		Type:     "s3_bucket",
		Expected: fmt.Sprintf("bucket %s accessible", check.Bucket),
		Actual:   fmt.Sprintf("HTTP %d (%s)", statusCode, elapsed.Round(time.Millisecond)),
		Passed:   is2xx(statusCode),
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/runner/... -run "TestCheckURLReachable|TestCheckServiceReachable|TestCheckS3Bucket" -v`
Expected: all pass

---

### Task 4: Implement version_check assertion

**Files:**
- Modify: `internal/runner/assertion.go`
- Create: `internal/runner/assertion_version_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/runner/assertion_version_test.go`:

```go
package runner

import (
	"testing"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

func TestCheckVersion_Match(t *testing.T) {
	result := CheckVersion(&schema.VersionCheck{
		Command: "echo 'go version go1.22.0 linux/amd64'",
		Pattern: `go1\.2[0-9]`,
	})
	if !result.Passed {
		t.Errorf("expected match, got: %s", result.Actual)
	}
}

func TestCheckVersion_NoMatch(t *testing.T) {
	result := CheckVersion(&schema.VersionCheck{
		Command: "echo 'node v18.0.0'",
		Pattern: `v20\.[0-9]+`,
	})
	if result.Passed {
		t.Error("expected no match")
	}
}

func TestCheckVersion_CommandFailure(t *testing.T) {
	result := CheckVersion(&schema.VersionCheck{
		Command: "false",
		Pattern: `.*`,
	})
	if result.Passed {
		t.Error("expected fail for non-zero exit")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/runner/... -run TestCheckVersion -v`
Expected: compilation error

- [ ] **Step 3: Implement CheckVersion**

```go
func CheckVersion(check *schema.VersionCheck) AssertionResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", check.Command)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return AssertionResult{
			Type:     "version_check",
			Expected: fmt.Sprintf("pattern %q", check.Pattern),
			Actual:   fmt.Sprintf("command failed: %v", err),
			Passed:   false,
		}
	}

	re := regexp.MustCompile(check.Pattern)
	output := strings.TrimSpace(stdout.String())
	if re.MatchString(output) {
		return AssertionResult{
			Type:     "version_check",
			Expected: fmt.Sprintf("pattern %q", check.Pattern),
			Actual:   output,
			Passed:   true,
		}
	}
	return AssertionResult{
		Type:     "version_check",
		Expected: fmt.Sprintf("pattern %q", check.Pattern),
		Actual:   fmt.Sprintf("output %q did not match", output),
		Passed:   false,
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/runner/... -run TestCheckVersion -v`
Expected: all pass

---

## Chunk 3: Wiring + Pre-commit Hook

### Task 5: Wire new assertions into runner

**Files:**
- Modify: `internal/runner/runner.go`

- [ ] **Step 1: Add assertion evaluation blocks in runTestOnce() (after existing DockerImage block, ~line 399)**

```go
if t.Expect.URLReachable != nil {
	a := CheckURLReachable(t.Expect.URLReachable)
	assertions = append(assertions, a)
	if !a.Passed {
		allPassed = false
	}
}
if t.Expect.ServiceReachable != nil {
	a := CheckServiceReachable(t.Expect.ServiceReachable)
	assertions = append(assertions, a)
	if !a.Passed {
		allPassed = false
	}
}
if t.Expect.S3Bucket != nil {
	a := CheckS3Bucket(t.Expect.S3Bucket)
	assertions = append(assertions, a)
	if !a.Passed {
		allPassed = false
	}
}
if t.Expect.VersionCheck != nil {
	a := CheckVersion(t.Expect.VersionCheck)
	assertions = append(assertions, a)
	if !a.Passed {
		allPassed = false
	}
}
```

- [ ] **Step 2: Build and run full test suite**

Run: `go build ./... && go test ./...`
Expected: success, all existing + new tests pass

---

### Task 6: Add pre-commit hook file

**Files:**
- Create: `.pre-commit-hooks.yaml`

- [ ] **Step 1: Create the hook file**

```yaml
- id: smoke
  name: cosmo-smoke
  description: Run smoke tests from .smoke.yaml
  entry: smoke run --fail-fast
  language: system
  pass_filenames: false
  always_run: true
  stages: [pre-commit]
```

- [ ] **Step 2: Add validation test**

Add to `assertion_reachable_test.go` or create `precommit_test.go` at project root:

```go
package main

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPrecommitHooksYAML(t *testing.T) {
	data, err := os.ReadFile(".pre-commit-hooks.yaml")
	if err != nil {
		t.Fatalf("reading .pre-commit-hooks.yaml: %v", err)
	}
	var hooks []map[string]interface{}
	if err := yaml.Unmarshal(data, &hooks); err != nil {
		t.Fatalf("parsing YAML: %v", err)
	}
	if len(hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hooks))
	}
	if hooks[0]["id"] != "smoke" {
		t.Errorf("hook id = %v, want smoke", hooks[0]["id"])
	}
	if hooks[0]["entry"] != "smoke run --fail-fast" {
		t.Errorf("hook entry = %v", want correct entry", hooks[0]["entry"])
	}
}
```

- [ ] **Step 3: Run full suite**

Run: `go test ./...`
Expected: all pass

---

## Release Checklist

- [ ] Run `smoke run` self-smoke (6+ tests pass)
- [ ] Run `go test ./...` (246+ tests pass)
- [ ] Update CLAUDE.md assertion table with 4 new types
- [ ] Update CLAUDE.md test count from 64 to current
- [ ] Changelog: `ccs changelog finalize v0.6.0 "connect-and-verify"`
- [ ] Version: `ccs version-track bump minor`
- [ ] Tag: `git tag v0.6.0`
- [ ] Push: `git push origin master --tags`
