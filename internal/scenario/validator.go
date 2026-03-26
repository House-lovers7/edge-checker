package scenario

import (
	"fmt"
	"net/url"
	"slices"
	"time"
)

// ValidationError represents a single validation issue.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) String() string {
	return fmt.Sprintf("[%s] %s", e.Field, e.Message)
}

// Validate checks a scenario for correctness and returns all errors found.
func Validate(s *Scenario) []ValidationError {
	var errs []ValidationError
	add := func(field, msg string) {
		errs = append(errs, ValidationError{Field: field, Message: msg})
	}

	// Required fields
	if s.Name == "" {
		add("name", "is required")
	}
	if s.Target.BaseURL == "" {
		add("target.base_url", "is required")
	}
	if s.Target.Path == "" {
		add("target.path", "is required")
	}
	if s.Execution.Mode == "" {
		add("execution.mode", "is required")
	}
	if s.Execution.Duration == "" {
		add("execution.duration", "is required")
	}

	// Validate base_url is a valid URL
	if s.Target.BaseURL != "" {
		u, err := url.Parse(s.Target.BaseURL)
		if err != nil {
			add("target.base_url", fmt.Sprintf("is not a valid URL: %v", err))
		} else if u.Scheme != "http" && u.Scheme != "https" {
			add("target.base_url", "must use http or https scheme")
		}
	}

	// Validate method
	validMethods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "PATCH"}
	if !slices.Contains(validMethods, s.Target.Method) {
		add("target.method", fmt.Sprintf("must be one of %v, got %q", validMethods, s.Target.Method))
	}

	// Validate mode
	if s.Execution.Mode != "" && !slices.Contains(ValidModes(), s.Execution.Mode) {
		add("execution.mode", fmt.Sprintf("must be one of %v, got %q", ValidModes(), s.Execution.Mode))
	}

	// Validate duration is parseable
	if s.Execution.Duration != "" {
		if _, err := time.ParseDuration(s.Execution.Duration); err != nil {
			add("execution.duration", fmt.Sprintf("invalid duration format: %v", err))
		}
	}

	// Validate timeout is parseable
	if s.Execution.Timeout != "" {
		if _, err := time.ParseDuration(s.Execution.Timeout); err != nil {
			add("execution.timeout", fmt.Sprintf("invalid duration format: %v", err))
		}
	}

	// Validate concurrency
	if s.Execution.Concurrency < 1 {
		add("execution.concurrency", "must be at least 1")
	}

	// Validate RPS
	if s.Rate.RPS < 1 {
		add("rate.rps", "must be at least 1")
	}

	// Mode-specific validation
	switch s.Execution.Mode {
	case ModeBurst:
		if s.Rate.Burst == nil {
			add("rate.burst", "is required when mode is burst")
		} else {
			if s.Rate.Burst.SpikeRPS < 1 {
				add("rate.burst.spike_rps", "must be at least 1")
			}
			if s.Rate.Burst.SpikeDuration != "" {
				if _, err := time.ParseDuration(s.Rate.Burst.SpikeDuration); err != nil {
					add("rate.burst.spike_duration", fmt.Sprintf("invalid duration: %v", err))
				}
			}
			if s.Rate.Burst.Interval != "" {
				if _, err := time.ParseDuration(s.Rate.Burst.Interval); err != nil {
					add("rate.burst.interval", fmt.Sprintf("invalid duration: %v", err))
				}
			}
		}
	case ModeRamp:
		if s.Rate.Ramp == nil {
			add("rate.ramp", "is required when mode is ramp")
		} else {
			if s.Rate.Ramp.StartRPS < 1 {
				add("rate.ramp.start_rps", "must be at least 1")
			}
			if s.Rate.Ramp.EndRPS < s.Rate.Ramp.StartRPS {
				add("rate.ramp.end_rps", "must be greater than or equal to start_rps")
			}
			if s.Rate.Ramp.StepDuration != "" {
				if _, err := time.ParseDuration(s.Rate.Ramp.StepDuration); err != nil {
					add("rate.ramp.step_duration", fmt.Sprintf("invalid duration: %v", err))
				}
			}
		}
	case ModeCooldownCheck:
		if s.Rate.Cooldown == nil {
			add("rate.cooldown", "is required when mode is cooldown-check")
		} else {
			if s.Rate.Cooldown.ActiveRPS < 1 {
				add("rate.cooldown.active_rps", "must be at least 1")
			}
			if s.Rate.Cooldown.ActiveDuration != "" {
				if _, err := time.ParseDuration(s.Rate.Cooldown.ActiveDuration); err != nil {
					add("rate.cooldown.active_duration", fmt.Sprintf("invalid duration: %v", err))
				}
			}
			if s.Rate.Cooldown.WaitDuration != "" {
				if _, err := time.ParseDuration(s.Rate.Cooldown.WaitDuration); err != nil {
					add("rate.cooldown.wait_duration", fmt.Sprintf("invalid duration: %v", err))
				}
			}
		}
	}

	// Validate safety
	if len(s.Safety.AllowHosts) == 0 {
		add("safety.allow_hosts", "must contain at least one allowed host")
	}

	// Check that target host is in allow_hosts
	if s.Target.BaseURL != "" && len(s.Safety.AllowHosts) > 0 {
		u, err := url.Parse(s.Target.BaseURL)
		if err == nil {
			if !slices.Contains(s.Safety.AllowHosts, u.Host) {
				add("safety.allow_hosts", fmt.Sprintf("target host %q is not in allow_hosts", u.Host))
			}
		}
	}

	// Validate expect
	if len(s.Expect.BlockStatusCodes) == 0 {
		add("expect.block_status_codes", "must contain at least one expected block status code")
	}

	return errs
}
