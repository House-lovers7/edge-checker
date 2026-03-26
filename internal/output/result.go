package output

import (
	"time"

	"github.com/House-lovers7/edge-checker/internal/judge"
	"github.com/House-lovers7/edge-checker/internal/observe"
)

// Result represents the complete output of a test run.
type Result struct {
	ScenarioName string            `json:"scenario_name"`
	Description  string            `json:"description"`
	StartedAt    time.Time         `json:"started_at"`
	EndedAt      time.Time         `json:"ended_at"`
	Duration     string            `json:"duration"`
	Interrupted  bool              `json:"interrupted,omitempty"`
	Target       TargetInfo        `json:"target"`
	Execution    ExecutionInfo     `json:"execution"`
	Summary      Summary                `json:"summary"`
	Timeline     []observe.SecondBucket `json:"timeline"`
	Verdict      *judge.Verdict         `json:"verdict,omitempty"`
}

// TargetInfo describes what was tested.
type TargetInfo struct {
	BaseURL string `json:"base_url"`
	Path    string `json:"path"`
	Method  string `json:"method"`
	Profile string `json:"profile"`
}

// ExecutionInfo describes how the test was run.
type ExecutionInfo struct {
	Mode        string `json:"mode"`
	Duration    string `json:"duration"`
	Concurrency int    `json:"concurrency"`
	RPS         int    `json:"rps"`
	Environment string `json:"environment"`
}

// Summary contains aggregate metrics.
type Summary struct {
	TotalRequests int         `json:"total_requests"`
	SuccessCount  int         `json:"success_count"`
	ErrorCount    int         `json:"error_count"`
	StatusCounts  map[int]int `json:"status_counts"`
	AvgLatencyMs  float64     `json:"avg_latency_ms"`
	P50LatencyMs  float64     `json:"p50_latency_ms"`
	P95LatencyMs  float64     `json:"p95_latency_ms"`
	P99LatencyMs  float64     `json:"p99_latency_ms"`
}

// BuildResult creates a Result from collected metrics and scenario info.
func BuildResult(
	scenarioName, description string,
	targetInfo TargetInfo,
	execInfo ExecutionInfo,
	metrics *observe.Metrics,
	interrupted bool,
) *Result {
	return &Result{
		ScenarioName: scenarioName,
		Description:  description,
		StartedAt:    metrics.StartTime,
		EndedAt:      metrics.EndTime,
		Duration:     metrics.EndTime.Sub(metrics.StartTime).Round(time.Millisecond).String(),
		Interrupted:  interrupted,
		Target:       targetInfo,
		Execution:    execInfo,
		Summary: Summary{
			TotalRequests: metrics.TotalRequests,
			SuccessCount:  metrics.SuccessCount(),
			ErrorCount:    metrics.ErrorCount,
			StatusCounts:  metrics.StatusCounts,
			AvgLatencyMs:  float64(metrics.AvgLatency.Microseconds()) / 1000.0,
			P50LatencyMs:  float64(metrics.P50Latency.Microseconds()) / 1000.0,
			P95LatencyMs:  float64(metrics.P95Latency.Microseconds()) / 1000.0,
			P99LatencyMs:  float64(metrics.P99Latency.Microseconds()) / 1000.0,
		},
		Timeline: metrics.Timeline,
	}
}
