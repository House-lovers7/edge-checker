package judge

import (
	"fmt"

	"github.com/House-lovers7/edge-checker/internal/engine"
)

// CheckBypass evaluates whether bypass scenarios were correctly excluded from blocking.
func CheckBypass(probeResults []engine.ProbeResult) RuleResult {
	var bypass []engine.ProbeResult
	for _, r := range probeResults {
		if r.Type == "bypass" {
			bypass = append(bypass, r)
		}
	}

	if len(bypass) == 0 {
		return RuleResult{
			Name:    "Bypass check",
			Status:  StatusSkip,
			Message: "No bypass scenarios configured",
		}
	}

	var failures []string
	for _, r := range bypass {
		if r.Error != "" {
			failures = append(failures, fmt.Sprintf("%s: error (%s)", r.Path, r.Error))
			continue
		}
		if r.StatusCode != r.Expected {
			failures = append(failures, fmt.Sprintf("%s: got %d, expected %d", r.Path, r.StatusCode, r.Expected))
		}
	}

	total := len(bypass)
	passed := total - len(failures)

	if len(failures) > 0 {
		return RuleResult{
			Name:     "Bypass check",
			Status:   StatusFail,
			Expected: "All bypass scenarios return expected status",
			Actual:   fmt.Sprintf("%d/%d passed", passed, total),
			Message:  fmt.Sprintf("Bypass not working correctly: %s", failures[0]),
		}
	}

	return RuleResult{
		Name:     "Bypass check",
		Status:   StatusPass,
		Expected: "All bypass scenarios return expected status",
		Actual:   fmt.Sprintf("%d/%d passed", passed, total),
		Message:  "Bypass is working correctly — bypass scenarios were not blocked",
	}
}
