---
title: "cosmo-smoke v0.6 — Connect and Verify"
created: 2026-04-18
status: approved
roadmap:
  - ROAD-018
  - ROAD-013
ideas:
  - IDEA-MO1FBWB9
  - IDEA-MO1FBQHL
---

# cosmo-smoke v0.6 — Connect and Verify

## Why

v0.5 expanded ecosystem coverage (Goss migration). v0.6 expands assertion coverage — the "universal smoke test" needs to answer not just "does it run?" but "does it connect to everything it needs?"

CosmoLabs has ~95 projects spanning HTTP APIs, gRPC microservices, Go/Node/Python runtimes, Docker deployments, and S3-backed storage. The current 20 assertion types cover process, HTTP, database, Docker, and protocol checks. The gaps are connectivity (can the service reach its dependencies?) and environment verification (are the right tools installed?).

## Design Decisions

1. **No new dependencies.** All assertions use stdlib: `net/http`, `regexp`, `os/exec`. The "minimal deps" principle is a competitive advantage.
2. **url_reachable as primitive.** `service_reachable` and `s3_bucket` both reduce to "make an HTTP request, assert on the response." A shared `httpReachable` internal function avoids duplication.
3. **Pre-commit via framework.** `.pre-commit-hooks.yaml` file in the repo. Zero code, maximum adoption.
4. **version_check via command execution.** Runs a shell command and regex-matches stdout. No need to parse version strings — the pattern does the work.

## Assertion Types

### 1. url_reachable

Generic HTTP/HTTPS connectivity check.

```yaml
expect:
  url_reachable: {url: "https://example.com", timeout: 5s, status_code: 200}
```

- `url` (required): HTTP or HTTPS URL
- `timeout` (optional, default 5s): request timeout
- `status_code` (optional): expected status code. Default: any 2xx passes

Internal: `net/http.Client` with timeout. Returns `AssertionResult` with actual status code and response time.

### 2. service_reachable (ROAD-018)

Semantic wrapper for external service dependencies.

```yaml
expect:
  service_reachable: {url: "https://api.stripe.com", timeout: 5s}
```

- `url` (required): service endpoint URL
- `timeout` (optional, default 5s): request timeout

Internally delegates to `url_reachable` with `status_code` defaulting to any 2xx. The separate type exists for semantic clarity in config files and to support `depends_on` DAGs in v0.7.

### 3. s3_bucket (IDEA-MO1FBWB9)

S3/compatible bucket accessibility check.

```yaml
expect:
  s3_bucket: {bucket: "my-bucket", region: "us-east-1", endpoint: "s3.amazonaws.com"}
```

- `bucket` (required): bucket name
- `region` (optional, default "us-east-1"): AWS region
- `endpoint` (optional, default "s3.amazonaws.com"): custom endpoint for S3-compatible storage (MinIO, GCS, etc.)

Uses path-style URL: `https://s3.amazonaws.com/my-bucket?location`. Anonymous HEAD request. For authenticated access, users reference env vars via Go templates in the broader test config.

### 4. version_check (ROAD-013)

Tool version verification via shell command.

```yaml
expect:
  version_check: {command: "go version", pattern: "go1\\.[0-9]+"}
```

- `command` (required): shell command to run
- `pattern` (required): Go regex to match against stdout

Runs via `os/exec` ("sh", "-c", command). Regex match on combined stdout. Fails if command exits non-zero or pattern doesn't match.

### 5. Pre-commit hook (IDEA-MO1FBQHL)

`.pre-commit-hooks.yaml` file in repository root:

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

No code changes. Users add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/CosmoLabs-org/cosmo-smoke
    rev: v0.6.0
    hooks:
      - id: smoke
```

## Schema Changes

New structs in `internal/schema/schema.go`:

```go
type URLReachableCheck struct {
    URL        string   `yaml:"url"`
    Timeout    Duration `yaml:"timeout,omitempty"`
    StatusCode *int     `yaml:"status_code,omitempty"`
}

type ServiceReachableCheck struct {
    URL     string `yaml:"url"`
    Timeout Duration `yaml:"timeout,omitempty"`
}

type S3BucketCheck struct {
    Bucket   string `yaml:"bucket"`
    Region   string `yaml:"region,omitempty"`
    Endpoint string `yaml:"endpoint,omitempty"`
}

type VersionCheck struct {
    Command string `yaml:"command"`
    Pattern string `yaml:"pattern"`
}
```

New fields on `Expect`:

```go
URLReachable     *URLReachableCheck     `yaml:"url_reachable,omitempty"`
ServiceReachable *ServiceReachableCheck `yaml:"service_reachable,omitempty"`
S3Bucket         *S3BucketCheck         `yaml:"s3_bucket,omitempty"`
VersionCheck     *VersionCheck          `yaml:"version_check,omitempty"`
```

## Runner Changes

New check functions in `internal/runner/assertion.go`:

- `CheckURLReachable(check *schema.URLReachableCheck) AssertionResult`
- `CheckServiceReachable(check *schema.ServiceReachableCheck) AssertionResult`
- `CheckS3Bucket(check *schema.S3BucketCheck) AssertionResult`
- `CheckVersionCheck(check *schema.VersionCheck) AssertionResult`

Internal shared helper: `httpReachable(url string, timeout time.Duration, expectedStatus int) (int, time.Duration, error)`

Wire in `runner.go` `runTestOnce()` alongside existing assertion checks.

## Error Handling

| Condition | Behavior |
|-----------|----------|
| Network timeout | `Passed: false`, `Actual: "connection timed out after Xs"` |
| DNS failure | `Passed: false`, `Actual: "DNS lookup failed: ..."` |
| Non-2xx response | `Passed: false`, `Actual: "HTTP 403 Forbidden"` |
| S3 403 | `Passed: false`, hint: "bucket may require authentication" |
| version_check command failure | `Passed: false`, `Actual: "exit code N"` |
| version_check pattern miss | `Passed: false`, `Actual: "output 'go version go1.22.0' did not match pattern 'go2\\..*'"` |

## Testing Plan

| Assertion | Test approach | Count |
|-----------|--------------|-------|
| `url_reachable` | httptest server: 2xx pass, 5xx fail, timeout, invalid URL | 4 |
| `service_reachable` | httptest server: pass via delegation, fail | 2 |
| `s3_bucket` | httptest server mocking S3 HEAD: 200 pass, 404 fail, 403 auth hint | 3 |
| `version_check` | test helper running echo commands: pattern match, pattern miss, command failure | 3 |
| Pre-commit YAML | validate YAML structure, verify hook fields | 1 |

~13 new tests. Suite total: ~246.

## File Scope

```yaml
files_modified:
  - internal/schema/schema.go
  - internal/runner/runner.go
  - internal/runner/assertion.go
files_created:
  - internal/runner/assertion_reachable_test.go
  - internal/runner/assertion_version_test.go
  - .pre-commit-hooks.yaml
```

## v0.6 Release Scope

Headline: "cosmo-smoke v0.6 — Connect and Verify"

- service_reachable: verify external service dependencies (ROAD-018)
- version_check: assert installed tool versions (ROAD-013)
- s3_bucket: S3/compatible bucket accessibility (IDEA-MO1FBWB9)
- url_reachable: generic HTTP connectivity primitive
- Pre-commit hook: zero-config smoke test integration (IDEA-MO1FBQHL)
- 246+ tests passing
