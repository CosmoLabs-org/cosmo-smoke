package schema

import (
	"fmt"
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
		if t.Run == "" {
			errs = append(errs, fmt.Sprintf("%s: run command is required", prefix))
		}
		if t.Retry != nil {
			if t.Retry.Count < 1 {
				errs = append(errs, fmt.Sprintf("test[%d] retry.count must be >= 1", i))
			}
			if t.Retry.Backoff.Duration <= 0 {
				errs = append(errs, fmt.Sprintf("test[%d] retry.backoff must be > 0", i))
			}
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}
