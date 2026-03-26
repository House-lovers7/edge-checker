package safety

import (
	"fmt"
	"time"
)

// Hard limits that cannot be exceeded regardless of scenario configuration.
const (
	HardMaxTotalRequests = 100_000
	HardMaxConcurrency   = 100
	HardMaxDuration      = 30 * time.Minute
	HardMaxRPS           = 1_000
)

// CheckLimits verifies that the scenario's parameters are within safe bounds.
// estimatedRequests should be calculated by the caller based on the execution mode.
func CheckLimits(maxTotalRequests, concurrency, rps int, duration time.Duration, estimatedRequests int) []error {
	var errs []error

	if maxTotalRequests > HardMaxTotalRequests {
		errs = append(errs, fmt.Errorf("max_total_requests %d exceeds hard limit %d", maxTotalRequests, HardMaxTotalRequests))
	}

	if concurrency > HardMaxConcurrency {
		errs = append(errs, fmt.Errorf("concurrency %d exceeds hard limit %d", concurrency, HardMaxConcurrency))
	}

	if rps > HardMaxRPS {
		errs = append(errs, fmt.Errorf("rps %d exceeds hard limit %d", rps, HardMaxRPS))
	}

	if duration > HardMaxDuration {
		errs = append(errs, fmt.Errorf("duration %s exceeds hard limit %s", duration, HardMaxDuration))
	}

	if maxTotalRequests > 0 && estimatedRequests > maxTotalRequests {
		errs = append(errs, fmt.Errorf("estimated total requests %d exceeds max_total_requests %d", estimatedRequests, maxTotalRequests))
	}

	return errs
}
