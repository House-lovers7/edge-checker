package judge

import (
	"fmt"

	"github.com/House-lovers7/edge-checker/internal/observe"
	"github.com/House-lovers7/edge-checker/internal/scenario"
)

// VerdictStatus represents the outcome of a judgment.
type VerdictStatus string

const (
	StatusPass VerdictStatus = "PASS"
	StatusFail VerdictStatus = "FAIL"
	StatusSkip VerdictStatus = "SKIP"
)

// RuleResult is the outcome of a single judgment rule.
type RuleResult struct {
	Name     string        `json:"name"`
	Status   VerdictStatus `json:"status"`
	Expected string        `json:"expected"`
	Actual   string        `json:"actual"`
	Message  string        `json:"message"`
}

// Verdict is the overall judgment of a test run.
type Verdict struct {
	Overall VerdictStatus `json:"overall"`
	Rules   []RuleResult  `json:"rules"`
}

// Judge evaluates collected metrics against expected outcomes.
func Judge(metrics *observe.Metrics, expect *scenario.Expect) *Verdict {
	v := &Verdict{
		Overall: StatusPass,
	}

	// Rule 1: Block detection — did block-like statuses appear within the expected time?
	v.Rules = append(v.Rules, checkBlockDetection(metrics, expect))

	// Rule 2: Block ratio — was the overall block ratio sufficient?
	v.Rules = append(v.Rules, checkBlockRatio(metrics, expect))

	// Determine overall
	for _, r := range v.Rules {
		if r.Status == StatusFail {
			v.Overall = StatusFail
			break
		}
	}

	return v
}

func checkBlockDetection(metrics *observe.Metrics, expect *scenario.Expect) RuleResult {
	if len(expect.BlockStatusCodes) == 0 || expect.BlockWithinSeconds <= 0 {
		return RuleResult{
			Name:    "Block detection",
			Status:  StatusSkip,
			Message: "No block detection criteria specified",
		}
	}

	blockCodes := make(map[int]bool)
	for _, c := range expect.BlockStatusCodes {
		blockCodes[c] = true
	}

	firstBlockSecond := -1
	for _, bucket := range metrics.Timeline {
		for code := range bucket.StatusCounts {
			if blockCodes[code] {
				firstBlockSecond = bucket.Second
				break
			}
		}
		if firstBlockSecond >= 0 {
			break
		}
	}

	expected := fmt.Sprintf("Block status %v within %ds", expect.BlockStatusCodes, expect.BlockWithinSeconds)

	if firstBlockSecond < 0 {
		return RuleResult{
			Name:     "Block detection",
			Status:   StatusFail,
			Expected: expected,
			Actual:   "No block statuses observed",
			Message:  "Defense did not trigger — no block-like status codes were returned",
		}
	}

	if firstBlockSecond > expect.BlockWithinSeconds {
		return RuleResult{
			Name:     "Block detection",
			Status:   StatusFail,
			Expected: expected,
			Actual:   fmt.Sprintf("First block at %ds", firstBlockSecond),
			Message:  fmt.Sprintf("Defense triggered too late (at %ds, expected within %ds)", firstBlockSecond, expect.BlockWithinSeconds),
		}
	}

	return RuleResult{
		Name:     "Block detection",
		Status:   StatusPass,
		Expected: expected,
		Actual:   fmt.Sprintf("First block at %ds", firstBlockSecond),
		Message:  fmt.Sprintf("Defense triggered within expected window (at %ds)", firstBlockSecond),
	}
}

func checkBlockRatio(metrics *observe.Metrics, expect *scenario.Expect) RuleResult {
	if expect.MinBlockRatio <= 0 {
		return RuleResult{
			Name:    "Block ratio",
			Status:  StatusSkip,
			Message: "No minimum block ratio specified",
		}
	}

	if metrics.TotalRequests == 0 {
		return RuleResult{
			Name:     "Block ratio",
			Status:   StatusFail,
			Expected: fmt.Sprintf(">= %.0f%%", expect.MinBlockRatio*100),
			Actual:   "No requests sent",
			Message:  "Cannot evaluate block ratio — no requests were sent",
		}
	}

	blockCount := metrics.BlockCount(expect.BlockStatusCodes)
	actualRatio := float64(blockCount) / float64(metrics.TotalRequests)

	expected := fmt.Sprintf(">= %.0f%%", expect.MinBlockRatio*100)
	actual := fmt.Sprintf("%.1f%% (%d/%d)", actualRatio*100, blockCount, metrics.TotalRequests)

	if actualRatio < expect.MinBlockRatio {
		return RuleResult{
			Name:     "Block ratio",
			Status:   StatusFail,
			Expected: expected,
			Actual:   actual,
			Message:  "Block ratio is below the expected minimum",
		}
	}

	return RuleResult{
		Name:     "Block ratio",
		Status:   StatusPass,
		Expected: expected,
		Actual:   actual,
		Message:  "Block ratio meets the expected minimum",
	}
}
