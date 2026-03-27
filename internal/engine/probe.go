package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/House-lovers7/edge-checker/internal/httpclient"
	"github.com/House-lovers7/edge-checker/internal/scenario"
)

// ProbeResult records the outcome of a single probe request.
type ProbeResult struct {
	Type       string // "unaffected" or "bypass"
	Path       string
	Method     string
	StatusCode int
	Expected   int
	Error      string
	Duration   time.Duration
}

// ProbeRunner runs low-frequency probes against unaffected paths and bypass scenarios
// concurrently with the main test engine.
type ProbeRunner struct {
	baseURL  string
	host     string
	interval time.Duration
	results  []ProbeResult
	mu       sync.Mutex
}

// NewProbeRunner creates a probe runner.
func NewProbeRunner(baseURL, host string, interval time.Duration) *ProbeRunner {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &ProbeRunner{
		baseURL:  baseURL,
		host:     host,
		interval: interval,
	}
}

// Run executes probes periodically until context is cancelled.
func (p *ProbeRunner) Run(
	ctx context.Context,
	client *httpclient.Client,
	unaffected []scenario.UnaffectedPath,
	bypass []scenario.BypassScenario,
) {
	if len(unaffected) == 0 && len(bypass) == 0 {
		return
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Run initial probe immediately
	p.probeOnce(ctx, client, unaffected, bypass)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.probeOnce(ctx, client, unaffected, bypass)
		}
	}
}

func (p *ProbeRunner) probeOnce(
	ctx context.Context,
	client *httpclient.Client,
	unaffected []scenario.UnaffectedPath,
	bypass []scenario.BypassScenario,
) {
	// Probe unaffected paths
	for _, up := range unaffected {
		if ctx.Err() != nil {
			return
		}
		method := up.Method
		if method == "" {
			method = "GET"
		}
		url := p.baseURL + up.Path
		resp := client.Do(ctx, method, url)

		result := ProbeResult{
			Type:     "unaffected",
			Path:     up.Path,
			Method:   method,
			Expected: up.ExpectStatus,
			Duration: resp.Duration,
		}
		if resp.Error != "" {
			result.Error = resp.Error
		} else {
			result.StatusCode = resp.StatusCode
		}

		p.mu.Lock()
		p.results = append(p.results, result)
		p.mu.Unlock()
	}

	// Probe bypass scenarios
	for _, bs := range bypass {
		if ctx.Err() != nil {
			return
		}
		method := "GET"
		url := p.baseURL + bs.Path

		// Create a temporary client with bypass headers merged
		resp := doWithExtraHeaders(ctx, client, method, url, bs.Headers)

		result := ProbeResult{
			Type:     "bypass",
			Path:     bs.Path,
			Method:   method,
			Expected: bs.ExpectStatus,
			Duration: resp.Duration,
		}
		if resp.Error != "" {
			result.Error = resp.Error
		} else {
			result.StatusCode = resp.StatusCode
		}

		p.mu.Lock()
		p.results = append(p.results, result)
		p.mu.Unlock()
	}
}

// Results returns all probe results collected so far.
func (p *ProbeRunner) Results() []ProbeResult {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]ProbeResult, len(p.results))
	copy(out, p.results)
	return out
}

// Summary returns a human-readable summary of probe results.
func (p *ProbeRunner) Summary() string {
	results := p.Results()
	if len(results) == 0 {
		return "No probes executed"
	}

	unaffectedTotal, unaffectedPass := 0, 0
	bypassTotal, bypassPass := 0, 0

	for _, r := range results {
		switch r.Type {
		case "unaffected":
			unaffectedTotal++
			if r.Error == "" && r.StatusCode == r.Expected {
				unaffectedPass++
			}
		case "bypass":
			bypassTotal++
			if r.Error == "" && r.StatusCode == r.Expected {
				bypassPass++
			}
		}
	}

	parts := []string{}
	if unaffectedTotal > 0 {
		parts = append(parts, fmt.Sprintf("Unaffected: %d/%d passed", unaffectedPass, unaffectedTotal))
	}
	if bypassTotal > 0 {
		parts = append(parts, fmt.Sprintf("Bypass: %d/%d passed", bypassPass, bypassTotal))
	}

	out := ""
	for _, part := range parts {
		out += part + "\n"
	}
	return out
}

// doWithExtraHeaders sends a request with additional headers merged on top.
func doWithExtraHeaders(ctx context.Context, client *httpclient.Client, method, url string, headers map[string]string) *httpclient.Response {
	// We use the client's Do method which applies profile + extraHeaders,
	// but bypass headers need to be applied additionally.
	// For simplicity, we use DoWithHeaders.
	return client.DoWithHeaders(ctx, method, url, headers)
}
