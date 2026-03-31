package engine

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/url"
	"strconv"
	"time"

	"github.com/House-lovers7/edge-checker/internal/httpclient"
	"github.com/House-lovers7/edge-checker/internal/observe"
	"github.com/House-lovers7/edge-checker/internal/scenario"
)

// Engine defines the interface for traffic generation modes.
type Engine interface {
	// Run executes the traffic pattern. It blocks until completion or context cancellation.
	Run(ctx context.Context, client *httpclient.Client, collector *observe.Collector) error
}

// Config holds the common configuration extracted from a Scenario for engine use.
type Config struct {
	TargetURL    string
	Method       string
	RPS          int
	Concurrency  int
	MaxRequests  int
	QueryStatic  map[string]string
	QueryRandom  bool
}

// BuildURL returns the target URL with query parameters applied.
// If QueryRandom is true, a unique _t parameter is appended for cache busting.
func (c Config) BuildURL() string {
	if len(c.QueryStatic) == 0 && !c.QueryRandom {
		return c.TargetURL
	}

	u, err := url.Parse(c.TargetURL)
	if err != nil {
		return c.TargetURL
	}

	q := u.Query()
	for k, v := range c.QueryStatic {
		q.Set(k, v)
	}
	if c.QueryRandom {
		q.Set("_t", strconv.FormatInt(rand.Int64(), 36))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// New creates an appropriate Engine based on the scenario's execution mode.
func New(s *scenario.Scenario) (Engine, error) {
	cfg := Config{
		TargetURL:   s.Target.FullURL(),
		Method:      s.Target.Method,
		RPS:         s.Rate.RPS,
		Concurrency: s.Execution.Concurrency,
		MaxRequests: s.Safety.MaxTotalRequests,
		QueryStatic: s.Query.Static,
		QueryRandom: s.Query.RandomSuffix,
	}

	duration, err := s.Execution.ParseDuration()
	if err != nil {
		return nil, fmt.Errorf("invalid duration: %w", err)
	}

	switch s.Execution.Mode {
	case scenario.ModeConstant:
		return NewConstant(cfg, duration), nil

	case scenario.ModeBurst:
		if s.Rate.Burst == nil {
			return nil, fmt.Errorf("burst config is required for burst mode")
		}
		spikeDur, err := time.ParseDuration(s.Rate.Burst.SpikeDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid burst spike_duration: %w", err)
		}
		interval, err := time.ParseDuration(s.Rate.Burst.Interval)
		if err != nil {
			return nil, fmt.Errorf("invalid burst interval: %w", err)
		}
		return NewBurst(cfg, duration, s.Rate.Burst.SpikeRPS, spikeDur, interval), nil

	case scenario.ModeRamp:
		if s.Rate.Ramp == nil {
			return nil, fmt.Errorf("ramp config is required for ramp mode")
		}
		stepDur, err := time.ParseDuration(s.Rate.Ramp.StepDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid ramp step_duration: %w", err)
		}
		return NewRamp(cfg, duration, s.Rate.Ramp.StartRPS, s.Rate.Ramp.EndRPS, stepDur), nil

	case scenario.ModeCooldownCheck:
		if s.Rate.Cooldown == nil {
			return nil, fmt.Errorf("cooldown config is required for cooldown-check mode")
		}
		activeDur, err := time.ParseDuration(s.Rate.Cooldown.ActiveDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid cooldown active_duration: %w", err)
		}
		waitDur, err := time.ParseDuration(s.Rate.Cooldown.WaitDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid cooldown wait_duration: %w", err)
		}
		return NewCooldown(cfg, activeDur, s.Rate.Cooldown.ActiveRPS, waitDur, s.Rate.Cooldown.ProbeRPS), nil

	default:
		return nil, fmt.Errorf("unknown engine mode %q", s.Execution.Mode)
	}
}
