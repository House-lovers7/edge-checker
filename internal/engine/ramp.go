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

// Ramp implements a traffic pattern that gradually increases RPS over time.
type Ramp struct {
	cfg          Config
	duration     time.Duration
	startRPS     int
	endRPS       int
	stepDuration time.Duration
}

// NewRamp creates a ramp-mode engine.
func NewRamp(cfg Config, duration time.Duration, startRPS, endRPS int, stepDuration time.Duration) *Ramp {
	return &Ramp{
		cfg:          cfg,
		duration:     duration,
		startRPS:     startRPS,
		endRPS:       endRPS,
		stepDuration: stepDuration,
	}
}

// Run sends requests with gradually increasing rate.
func (r *Ramp) Run(ctx context.Context, client *httpclient.Client, collector *observe.Collector) error {
	ctx, cancel := context.WithTimeout(ctx, r.duration)
	defer cancel()

	// Calculate steps
	totalSteps := int(r.duration / r.stepDuration)
	if totalSteps < 1 {
		totalSteps = 1
	}
	rpsIncrement := float64(r.endRPS-r.startRPS) / float64(totalSteps)

	currentRPS := r.startRPS
	limiter := rate.NewLimiter(rate.Limit(currentRPS), currentRPS)
	sem := make(chan struct{}, r.cfg.Concurrency)
	var wg sync.WaitGroup
	var totalSent atomic.Int64

	collector.Start()
	defer collector.Stop()

	// Goroutine to step up rate
	go func() {
		ticker := time.NewTicker(r.stepDuration)
		defer ticker.Stop()
		step := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				step++
				newRPS := r.startRPS + int(rpsIncrement*float64(step))
				if newRPS > r.endRPS {
					newRPS = r.endRPS
				}
				if newRPS < 1 {
					newRPS = 1
				}
				limiter.SetLimit(rate.Limit(newRPS))
				limiter.SetBurst(newRPS)
			}
		}
	}()

	for {
		if ctx.Err() != nil {
			break
		}

		if r.cfg.MaxRequests > 0 && totalSent.Load() >= int64(r.cfg.MaxRequests) {
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
			resp := client.Do(ctx, r.cfg.Method, r.cfg.BuildURL())
			collector.Record(resp)
		}()
	}

done:
	wg.Wait()
	return nil
}
