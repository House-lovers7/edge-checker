package observe

import (
	"sync"
	"time"

	"github.com/House-lovers7/edge-checker/internal/httpclient"
)

// SecondBucket holds metrics for a single second of the test run.
type SecondBucket struct {
	Second       int         `json:"second"`
	RequestCount int         `json:"request_count"`
	StatusCounts map[int]int `json:"status_counts"`
	ErrorCount   int         `json:"error_count"`
	AvgLatency   float64     `json:"avg_latency_ms"`
}

// Timeline collects per-second metrics buckets.
type Timeline struct {
	mu      sync.Mutex
	buckets map[int]*bucketAccumulator
}

type bucketAccumulator struct {
	count        int
	statusCounts map[int]int
	errorCount   int
	totalLatency time.Duration
}

// NewTimeline creates a new Timeline.
func NewTimeline() *Timeline {
	return &Timeline{
		buckets: make(map[int]*bucketAccumulator),
	}
}

// Record adds a response to the appropriate second bucket.
func (t *Timeline) Record(second int, resp *httpclient.Response) {
	t.mu.Lock()
	defer t.mu.Unlock()

	b, ok := t.buckets[second]
	if !ok {
		b = &bucketAccumulator{
			statusCounts: make(map[int]int),
		}
		t.buckets[second] = b
	}

	b.count++
	b.totalLatency += resp.Duration

	if resp.Error != "" {
		b.errorCount++
	} else {
		b.statusCounts[resp.StatusCode]++
	}
}

// Buckets returns all second buckets in chronological order.
func (t *Timeline) Buckets() []SecondBucket {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.buckets) == 0 {
		return nil
	}

	// Find max second
	maxSecond := 0
	for s := range t.buckets {
		if s > maxSecond {
			maxSecond = s
		}
	}

	result := make([]SecondBucket, 0, maxSecond+1)
	for s := 0; s <= maxSecond; s++ {
		b, ok := t.buckets[s]
		if !ok {
			result = append(result, SecondBucket{
				Second:       s,
				StatusCounts: map[int]int{},
			})
			continue
		}

		avgLatency := float64(0)
		if b.count > 0 {
			avgLatency = float64(b.totalLatency.Milliseconds()) / float64(b.count)
		}

		// Copy status counts
		sc := make(map[int]int, len(b.statusCounts))
		for k, v := range b.statusCounts {
			sc[k] = v
		}

		result = append(result, SecondBucket{
			Second:       s,
			RequestCount: b.count,
			StatusCounts: sc,
			ErrorCount:   b.errorCount,
			AvgLatency:   avgLatency,
		})
	}

	return result
}
