package engine

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/House-lovers7/edge-checker/internal/httpclient"
	"github.com/House-lovers7/edge-checker/internal/observe"
	"golang.org/x/time/rate"
)

// Constant implements a fixed-rate traffic pattern.
type Constant struct {
	cfg      Config
	duration time.Duration
}

// NewConstant creates a constant-rate engine.
func NewConstant(cfg Config, duration time.Duration) *Constant {
	return &Constant{
		cfg:      cfg,
		duration: duration,
	}
}

// Run sends requests at a constant rate for the configured duration.
func (c *Constant) Run(ctx context.Context, client *httpclient.Client, collector *observe.Collector) error {
	// Create a context that expires after duration
	ctx, cancel := context.WithTimeout(ctx, c.duration)
	defer cancel()

	limiter := rate.NewLimiter(rate.Limit(c.cfg.RPS), c.cfg.RPS)
	sem := make(chan struct{}, c.cfg.Concurrency)
	var wg sync.WaitGroup
	var totalSent atomic.Int64

	collector.Start()
	defer collector.Stop()

	for {
		// Check context before waiting for rate limiter
		if ctx.Err() != nil {
			break
		}

		// Check max requests limit
		if c.cfg.MaxRequests > 0 && totalSent.Load() >= int64(c.cfg.MaxRequests) {
			break
		}

		// Wait for rate limiter
		if err := limiter.Wait(ctx); err != nil {
			break // context cancelled or deadline exceeded
		}

		// Acquire semaphore
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			goto done
		}

		totalSent.Add(1)
		wg.Add(1)
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()
			resp := client.Do(ctx, c.cfg.Method, c.cfg.BuildURL())
			collector.Record(resp)
		}()
	}

done:
	// Wait for in-flight requests to complete
	wg.Wait()
	return nil
}
