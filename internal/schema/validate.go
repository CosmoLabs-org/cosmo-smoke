package schema

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError collects multiple validation failures.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed:\n  - %s", strings.Join(e.Errors, "\n  - "))
}

// Validate checks a SmokeConfig for required fields and consistency.
// Returns all errors at once rather than stopping at the first.
func Validate(cfg *SmokeConfig) error {
	var errs []string

	if cfg.Version != 1 {
		errs = append(errs, fmt.Sprintf("unsupported version %d (expected 1)", cfg.Version))
	}

	if cfg.Project == "" {
		errs = append(errs, "project name is required")
	}

	if len(cfg.Tests) == 0 {
		errs = append(errs, "at least one test is required")
	}

	for i, t := range cfg.Tests {
		prefix := fmt.Sprintf("tests[%d]", i)
		if t.Name == "" {
			errs = append(errs, fmt.Sprintf("%s: name is required", prefix))
		}
		if t.Run == "" && !hasStandaloneAssertions(t.Expect) {
			errs = append(errs, fmt.Sprintf("%s: run command is required (or add a network/storage assertion)", prefix))
		}
		if t.Retry != nil {
			if t.Retry.Count < 1 {
				errs = append(errs, fmt.Sprintf("test[%d] retry.count must be >= 1", i))
			}
			if t.Retry.Backoff.Duration <= 0 {
				errs = append(errs, fmt.Sprintf("test[%d] retry.backoff must be > 0", i))
			}
		}
		if t.Expect.DockerContainer != nil && t.Expect.DockerContainer.Name == "" {
			errs = append(errs, fmt.Sprintf("%s: docker_container_running.name is required", prefix))
		}
		if t.Expect.DockerImage != nil && t.Expect.DockerImage.Image == "" {
			errs = append(errs, fmt.Sprintf("%s: docker_image_exists.image is required", prefix))
		}
		if e := t.Expect.URLReachable; e != nil {
			if e.URL == "" {
				errs = append(errs, fmt.Sprintf("%s: url_reachable.url is required", prefix))
			} else if !strings.HasPrefix(e.URL, "http://") && !strings.HasPrefix(e.URL, "https://") {
				errs = append(errs, fmt.Sprintf("%s: url_reachable.url must start with http:// or https://", prefix))
			}
		}
		if e := t.Expect.ServiceReachable; e != nil {
			if e.URL == "" {
				errs = append(errs, fmt.Sprintf("%s: service_reachable.url is required", prefix))
			} else if !strings.HasPrefix(e.URL, "http://") && !strings.HasPrefix(e.URL, "https://") {
				errs = append(errs, fmt.Sprintf("%s: service_reachable.url must start with http:// or https://", prefix))
			}
		}
		if e := t.Expect.S3Bucket; e != nil {
			if e.Bucket == "" {
				errs = append(errs, fmt.Sprintf("%s: s3_bucket.bucket is required", prefix))
			}
		}
		if e := t.Expect.VersionCheck; e != nil {
			if e.Command == "" {
				errs = append(errs, fmt.Sprintf("%s: version_check.command is required", prefix))
			}
			if e.Pattern == "" {
				errs = append(errs, fmt.Sprintf("%s: version_check.pattern is required", prefix))
			} else if _, err := regexp.Compile(e.Pattern); err != nil {
				errs = append(errs, fmt.Sprintf("%s: version_check.pattern is invalid regex: %v", prefix, err))
			}
		}
		if e := t.Expect.WebSocket; e != nil {
			if e.URL == "" {
				errs = append(errs, fmt.Sprintf("%s: websocket.url is required", prefix))
			} else if !strings.HasPrefix(e.URL, "ws://") && !strings.HasPrefix(e.URL, "wss://") {
				errs = append(errs, fmt.Sprintf("%s: websocket.url must start with ws:// or wss://", prefix))
			}
			if e.ExpectMatches != "" {
				if _, err := regexp.Compile(e.ExpectMatches); err != nil {
					errs = append(errs, fmt.Sprintf("%s: websocket.expect_matches is invalid regex: %v", prefix, err))
				}
			}
		}
		if e := t.Expect.OTelTrace; e != nil {
			if e.JaegerURL == "" && cfg.OTel.JaegerURL == "" {
				errs = append(errs, fmt.Sprintf("%s: otel_trace.jaeger_url is required (or set otel.jaeger_url globally)", prefix))
			} else if e.JaegerURL != "" && !strings.HasPrefix(e.JaegerURL, "http://") && !strings.HasPrefix(e.JaegerURL, "https://") {
				errs = append(errs, fmt.Sprintf("%s: otel_trace.jaeger_url must start with http:// or https://", prefix))
			}
			if e.MinSpans < 0 {
				errs = append(errs, fmt.Sprintf("%s: otel_trace.min_spans must be >= 0", prefix))
			}
		}
	}

	if cfg.OTel.Enabled && cfg.OTel.JaegerURL == "" {
		errs = append(errs, "otel.jaeger_url is required when otel is enabled")
	}
	if cfg.OTel.Enabled && cfg.OTel.JaegerURL != "" && !strings.HasPrefix(cfg.OTel.JaegerURL, "http://") && !strings.HasPrefix(cfg.OTel.JaegerURL, "https://") {
		errs = append(errs, "otel.jaeger_url must start with http:// or https://")
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// hasStandaloneAssertions returns true if the test has assertions that don't
// require command output (stdout/stderr). These can run without a run command.
func hasStandaloneAssertions(e Expect) bool {
	return e.PortListening != nil ||
		e.ProcessRunning != "" ||
		e.HTTP != nil ||
		e.SSLCert != nil ||
		e.Redis != nil ||
		e.Memcached != nil ||
		e.Postgres != nil ||
		e.MySQL != nil ||
		e.GRPCHealth != nil ||
		e.DockerContainer != nil ||
		e.DockerImage != nil ||
		e.URLReachable != nil ||
		e.ServiceReachable != nil ||
		e.S3Bucket != nil ||
		e.VersionCheck != nil ||
		e.WebSocket != nil ||
		e.OTelTrace != nil
}
