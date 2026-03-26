package observe

import (
	"sort"
	"sync"
	"time"

	"github.com/House-lovers7/edge-checker/internal/httpclient"
)

// Collector gathers metrics from HTTP responses in a thread-safe manner.
type Collector struct {
	mu        sync.Mutex
	responses []httpclient.Response
	timeline  *Timeline
	startTime time.Time
	endTime   time.Time
}

// NewCollector creates a new metrics collector.
func NewCollector() *Collector {
	return &Collector{
		timeline: NewTimeline(),
	}
}

// Start marks the beginning of a test run.
func (c *Collector) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startTime = time.Now()
}

// Stop marks the end of a test run.
func (c *Collector) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.endTime = time.Now()
}

// Record adds a response to the collector.
func (c *Collector) Record(resp *httpclient.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.responses = append(c.responses, *resp)

	if !c.startTime.IsZero() {
		second := int(resp.Timestamp.Sub(c.startTime).Seconds())
		c.timeline.Record(second, resp)
	}
}

// Snapshot returns a point-in-time summary of collected metrics.
func (c *Collector) Snapshot() *Metrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := &Metrics{
		TotalRequests: len(c.responses),
		StatusCounts:  make(map[int]int),
		StartTime:     c.startTime,
		EndTime:       c.endTime,
	}

	var durations []time.Duration

	for _, r := range c.responses {
		if r.Error != "" {
			m.ErrorCount++
		} else {
			m.StatusCounts[r.StatusCode]++
		}
		durations = append(durations, r.Duration)
	}

	if len(durations) > 0 {
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		m.P50Latency = percentile(durations, 0.50)
		m.P95Latency = percentile(durations, 0.95)
		m.P99Latency = percentile(durations, 0.99)

		var total time.Duration
		for _, d := range durations {
			total += d
		}
		m.AvgLatency = total / time.Duration(len(durations))
	}

	m.Timeline = c.timeline.Buckets()

	return m
}

// Metrics is a point-in-time summary of test run metrics.
type Metrics struct {
	TotalRequests int            `json:"total_requests"`
	StatusCounts  map[int]int    `json:"status_counts"`
	ErrorCount    int            `json:"error_count"`
	AvgLatency    time.Duration  `json:"avg_latency_ms"`
	P50Latency    time.Duration  `json:"p50_latency_ms"`
	P95Latency    time.Duration  `json:"p95_latency_ms"`
	P99Latency    time.Duration  `json:"p99_latency_ms"`
	StartTime     time.Time      `json:"started_at"`
	EndTime       time.Time      `json:"ended_at"`
	Timeline      []SecondBucket `json:"timeline"`
}

// SuccessCount returns the count of 2xx responses.
func (m *Metrics) SuccessCount() int {
	count := 0
	for code, n := range m.StatusCounts {
		if code >= 200 && code < 300 {
			count += n
		}
	}
	return count
}

// BlockCount returns the count of responses matching typical block status codes.
func (m *Metrics) BlockCount(blockCodes []int) int {
	count := 0
	for _, code := range blockCodes {
		count += m.StatusCounts[code]
	}
	return count
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}
