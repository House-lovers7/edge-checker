package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/House-lovers7/edge-checker/internal/httpclient"
	"github.com/House-lovers7/edge-checker/internal/observe"
	"github.com/House-lovers7/edge-checker/internal/profile"
)

func TestConstant_RunCompletesWithinDuration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		TargetURL:   server.URL + "/test",
		Method:      "GET",
		RPS:         10,
		Concurrency: 2,
		MaxRequests: 100,
	}

	eng := NewConstant(cfg, 2*time.Second)
	client := httpclient.NewClient(5*time.Second, &profile.Profile{Name: "test", Headers: map[string]string{}}, nil, "", false)
	defer client.Close()

	collector := observe.NewCollector()

	start := time.Now()
	err := eng.Run(context.Background(), client, collector)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if elapsed < 1900*time.Millisecond || elapsed > 3*time.Second {
		t.Errorf("expected ~2s duration, got %v", elapsed)
	}

	m := collector.Snapshot()
	if m.TotalRequests < 15 {
		t.Errorf("expected at least 15 requests at 10 rps for 2s, got %d", m.TotalRequests)
	}
	if m.StatusCounts[200] != m.TotalRequests {
		t.Errorf("expected all 200s, got status counts: %v", m.StatusCounts)
	}
}

func TestConstant_RespectsMaxRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		TargetURL:   server.URL + "/test",
		Method:      "GET",
		RPS:         100,
		Concurrency: 5,
		MaxRequests: 20,
	}

	eng := NewConstant(cfg, 10*time.Second)
	client := httpclient.NewClient(5*time.Second, &profile.Profile{Name: "test", Headers: map[string]string{}}, nil, "", false)
	defer client.Close()

	collector := observe.NewCollector()
	err := eng.Run(context.Background(), client, collector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := collector.Snapshot()
	if m.TotalRequests > 25 {
		t.Errorf("expected ~20 requests (max), got %d", m.TotalRequests)
	}
}

func TestConstant_RespectsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		TargetURL:   server.URL + "/test",
		Method:      "GET",
		RPS:         10,
		Concurrency: 2,
		MaxRequests: 10000,
	}

	eng := NewConstant(cfg, 30*time.Second)
	client := httpclient.NewClient(5*time.Second, &profile.Profile{Name: "test", Headers: map[string]string{}}, nil, "", false)
	defer client.Close()

	collector := observe.NewCollector()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()
	_ = eng.Run(ctx, client, collector)
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Errorf("expected cancellation within ~1s, got %v", elapsed)
	}
}
