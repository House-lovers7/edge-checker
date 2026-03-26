package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/House-lovers7/edge-checker/internal/httpclient"
	"github.com/House-lovers7/edge-checker/internal/observe"
	"golang.org/x/time/rate"
)

// Cooldown implements a three-phase traffic pattern:
//  1. Active: send at high rate to trigger defense
//  2. Wait: pause and send low-frequency probes to observe block status
//  3. Verify: confirm block has been lifted
type Cooldown struct {
	cfg            Config
	activeDuration time.Duration
	activeRPS      int
	waitDuration   time.Duration
	probeRPS       int
}

// NewCooldown creates a cooldown-check engine.
func NewCooldown(cfg Config, activeDuration time.Duration, activeRPS int, waitDuration time.Duration, probeRPS int) *Cooldown {
	return &Cooldown{
		cfg:            cfg,
		activeDuration: activeDuration,
		activeRPS:      activeRPS,
		waitDuration:   waitDuration,
		probeRPS:       probeRPS,
	}
}

// Run executes the three-phase cooldown check.
func (c *Cooldown) Run(ctx context.Context, client *httpclient.Client, collector *observe.Collector) error {
	collector.Start()
	defer collector.Stop()

	// Phase 1: Active — trigger defense
	fmt.Printf("  [Phase 1/3] Active: %d rps for %s\n", c.activeRPS, c.activeDuration)
	if err := c.runPhase(ctx, client, collector, c.activeRPS, c.activeDuration); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return nil
	}

	// Phase 2: Wait — probe during cooldown period
	fmt.Printf("  [Phase 2/3] Cooldown: probing at %d rps for %s\n", c.probeRPS, c.waitDuration)
	if err := c.runPhase(ctx, client, collector, c.probeRPS, c.waitDuration); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return nil
	}

	// Phase 3: Verify — short burst to confirm unblock
	verifyDuration := 5 * time.Second
	verifyRPS := c.activeRPS / 2
	if verifyRPS < 1 {
		verifyRPS = 1
	}
	fmt.Printf("  [Phase 3/3] Verify: %d rps for %s\n", verifyRPS, verifyDuration)
	return c.runPhase(ctx, client, collector, verifyRPS, verifyDuration)
}

func (c *Cooldown) runPhase(ctx context.Context, client *httpclient.Client, collector *observe.Collector, rps int, duration time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	limiter := rate.NewLimiter(rate.Limit(rps), rps)
	sem := make(chan struct{}, c.cfg.Concurrency)
	var wg sync.WaitGroup
	var totalSent atomic.Int64

	for {
		if ctx.Err() != nil {
			break
		}

		if c.cfg.MaxRequests > 0 && totalSent.Load() >= int64(c.cfg.MaxRequests) {
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
			resp := client.Do(ctx, c.cfg.Method, c.cfg.TargetURL)
			collector.Record(resp)
		}()
	}

done:
	wg.Wait()
	return nil
}
