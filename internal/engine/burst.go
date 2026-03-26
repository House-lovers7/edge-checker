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

// Burst implements a traffic pattern with periodic spikes above the base rate.
type Burst struct {
	cfg           Config
	duration      time.Duration
	spikeRPS      int
	spikeDuration time.Duration
	interval      time.Duration
}

// NewBurst creates a burst-mode engine.
func NewBurst(cfg Config, duration time.Duration, spikeRPS int, spikeDuration, interval time.Duration) *Burst {
	return &Burst{
		cfg:           cfg,
		duration:      duration,
		spikeRPS:      spikeRPS,
		spikeDuration: spikeDuration,
		interval:      interval,
	}
}

// Run sends requests at a base rate with periodic spikes.
func (b *Burst) Run(ctx context.Context, client *httpclient.Client, collector *observe.Collector) error {
	ctx, cancel := context.WithTimeout(ctx, b.duration)
	defer cancel()

	limiter := rate.NewLimiter(rate.Limit(b.cfg.RPS), b.cfg.RPS)
	sem := make(chan struct{}, b.cfg.Concurrency)
	var wg sync.WaitGroup
	var totalSent atomic.Int64

	collector.Start()
	defer collector.Stop()

	start := time.Now()

	// Goroutine to toggle spike/base rate
	go func() {
		ticker := time.NewTicker(b.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Enter spike
				limiter.SetLimit(rate.Limit(b.spikeRPS))
				limiter.SetBurst(b.spikeRPS)

				// Schedule return to base
				spikeTimer := time.NewTimer(b.spikeDuration)
				select {
				case <-ctx.Done():
					spikeTimer.Stop()
					return
				case <-spikeTimer.C:
					limiter.SetLimit(rate.Limit(b.cfg.RPS))
					limiter.SetBurst(b.cfg.RPS)
				}
			}
		}
	}()

	_ = start

	for {
		if ctx.Err() != nil {
			break
		}

		if b.cfg.MaxRequests > 0 && totalSent.Load() >= int64(b.cfg.MaxRequests) {
			break
		}

		if err := limiter.Wait(ctx); err != nil {
			break
		}

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
			resp := client.Do(ctx, b.cfg.Method, b.cfg.TargetURL)
			collector.Record(resp)
		}()
	}

done:
	wg.Wait()
	return nil
}
