package judge

import (
	"fmt"

	"github.com/House-lovers7/edge-checker/internal/engine"
)

// CheckFalsePositive evaluates whether unaffected paths were incorrectly blocked.
func CheckFalsePositive(probeResults []engine.ProbeResult) RuleResult {
	var unaffected []engine.ProbeResult
	for _, r := range probeResults {
		if r.Type == "unaffected" {
			unaffected = append(unaffected, r)
		}
	}

	if len(unaffected) == 0 {
		return RuleResult{
			Name:    "False positive",
			Status:  StatusSkip,
			Message: "No unaffected paths configured",
		}
	}

	var failures []string
	for _, r := range unaffected {
		if r.Error != "" {
			failures = append(failures, fmt.Sprintf("%s %s: error (%s)", r.Method, r.Path, r.Error))
			continue
		}
		if r.StatusCode != r.Expected {
			failures = append(failures, fmt.Sprintf("%s %s: got %d, expected %d", r.Method, r.Path, r.StatusCode, r.Expected))
		}
	}

	total := len(unaffected)
	passed := total - len(failures)

	if len(failures) > 0 {
		return RuleResult{
			Name:     "False positive",
			Status:   StatusFail,
			Expected: "All unaffected paths return expected status",
			Actual:   fmt.Sprintf("%d/%d passed", passed, total),
			Message:  fmt.Sprintf("Unaffected paths were impacted: %s", failures[0]),
		}
	}

	return RuleResult{
		Name:     "False positive",
		Status:   StatusPass,
		Expected: "All unaffected paths return expected status",
		Actual:   fmt.Sprintf("%d/%d passed", passed, total),
		Message:  "No false positives detected — unaffected paths remain accessible",
	}
}
