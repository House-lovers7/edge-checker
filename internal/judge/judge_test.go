package judge

import (
	"testing"

	"github.com/House-lovers7/edge-checker/internal/observe"
	"github.com/House-lovers7/edge-checker/internal/scenario"
)

func TestJudge_BlockDetectionPass(t *testing.T) {
	metrics := &observe.Metrics{
		TotalRequests: 100,
		StatusCounts:  map[int]int{200: 70, 403: 30},
		Timeline: []observe.SecondBucket{
			{Second: 0, StatusCounts: map[int]int{200: 10}},
			{Second: 1, StatusCounts: map[int]int{200: 10}},
			{Second: 2, StatusCounts: map[int]int{200: 5, 403: 5}},
			{Second: 3, StatusCounts: map[int]int{403: 10}},
		},
	}
	expect := &scenario.Expect{
		BlockStatusCodes:   []int{403, 429},
		BlockWithinSeconds: 10,
		MinBlockRatio:      0.20,
	}

	v := Judge(metrics, expect)

	if v.Overall != StatusPass {
		t.Errorf("expected PASS, got %s", v.Overall)
	}
	if len(v.Rules) != 4 {
		t.Errorf("expected 4 rules, got %d", len(v.Rules))
	}
}

func TestJudge_BlockDetectionFail_NoBlocks(t *testing.T) {
	metrics := &observe.Metrics{
		TotalRequests: 100,
		StatusCounts:  map[int]int{200: 100},
		Timeline: []observe.SecondBucket{
			{Second: 0, StatusCounts: map[int]int{200: 50}},
			{Second: 1, StatusCounts: map[int]int{200: 50}},
		},
	}
	expect := &scenario.Expect{
		BlockStatusCodes:   []int{403, 429},
		BlockWithinSeconds: 5,
		MinBlockRatio:      0.20,
	}

	v := Judge(metrics, expect)

	if v.Overall != StatusFail {
		t.Errorf("expected FAIL, got %s", v.Overall)
	}
}

func TestJudge_BlockDetectionFail_TooLate(t *testing.T) {
	metrics := &observe.Metrics{
		TotalRequests: 100,
		StatusCounts:  map[int]int{200: 80, 403: 20},
		Timeline: []observe.SecondBucket{
			{Second: 0, StatusCounts: map[int]int{200: 20}},
			{Second: 5, StatusCounts: map[int]int{200: 20}},
			{Second: 10, StatusCounts: map[int]int{200: 20}},
			{Second: 15, StatusCounts: map[int]int{403: 20}},
			{Second: 20, StatusCounts: map[int]int{200: 20}},
		},
	}
	expect := &scenario.Expect{
		BlockStatusCodes:   []int{403},
		BlockWithinSeconds: 10,
		MinBlockRatio:      0.10,
	}

	v := Judge(metrics, expect)

	// Block detection should fail (first block at 15s > 10s)
	blockRule := v.Rules[0]
	if blockRule.Status != StatusFail {
		t.Errorf("expected block detection FAIL, got %s: %s", blockRule.Status, blockRule.Message)
	}
}

func TestJudge_BlockRatioFail(t *testing.T) {
	metrics := &observe.Metrics{
		TotalRequests: 100,
		StatusCounts:  map[int]int{200: 95, 403: 5},
		Timeline: []observe.SecondBucket{
			{Second: 0, StatusCounts: map[int]int{200: 95, 403: 5}},
		},
	}
	expect := &scenario.Expect{
		BlockStatusCodes:   []int{403},
		BlockWithinSeconds: 10,
		MinBlockRatio:      0.50,
	}

	v := Judge(metrics, expect)

	ratioRule := v.Rules[1]
	if ratioRule.Status != StatusFail {
		t.Errorf("expected ratio FAIL (5%% < 50%%), got %s", ratioRule.Status)
	}
}

func TestJudge_SkipWhenNoCriteria(t *testing.T) {
	metrics := &observe.Metrics{
		TotalRequests: 100,
		StatusCounts:  map[int]int{200: 100},
	}
	expect := &scenario.Expect{}

	v := Judge(metrics, expect)

	for _, r := range v.Rules {
		if r.Status != StatusSkip {
			t.Errorf("expected SKIP for rule %q with no criteria, got %s", r.Name, r.Status)
		}
	}
}
