package observe

import (
	"sync"
	"testing"
	"time"

	"github.com/House-lovers7/edge-checker/internal/httpclient"
)

func TestCollector_Record(t *testing.T) {
	c := NewCollector()
	c.Start()

	c.Record(&httpclient.Response{
		StatusCode: 200,
		Duration:   10 * time.Millisecond,
		Timestamp:  time.Now(),
	})
	c.Record(&httpclient.Response{
		StatusCode: 403,
		Duration:   20 * time.Millisecond,
		Timestamp:  time.Now(),
	})

	c.Stop()
	m := c.Snapshot()

	if m.TotalRequests != 2 {
		t.Errorf("expected 2 total requests, got %d", m.TotalRequests)
	}
	if m.StatusCounts[200] != 1 {
		t.Errorf("expected 1 x 200, got %d", m.StatusCounts[200])
	}
	if m.StatusCounts[403] != 1 {
		t.Errorf("expected 1 x 403, got %d", m.StatusCounts[403])
	}
}

func TestCollector_ErrorCounting(t *testing.T) {
	c := NewCollector()
	c.Start()

	c.Record(&httpclient.Response{
		Error:     "connection refused",
		Duration:  5 * time.Millisecond,
		Timestamp: time.Now(),
	})

	c.Stop()
	m := c.Snapshot()

	if m.ErrorCount != 1 {
		t.Errorf("expected 1 error, got %d", m.ErrorCount)
	}
}

func TestCollector_ConcurrentAccess(t *testing.T) {
	c := NewCollector()
	c.Start()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Record(&httpclient.Response{
				StatusCode: 200,
				Duration:   1 * time.Millisecond,
				Timestamp:  time.Now(),
			})
		}()
	}
	wg.Wait()

	c.Stop()
	m := c.Snapshot()

	if m.TotalRequests != 100 {
		t.Errorf("expected 100 requests, got %d", m.TotalRequests)
	}
}
