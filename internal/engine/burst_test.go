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

func TestBurst_RunCompletes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		TargetURL:   server.URL + "/test",
		Method:      "GET",
		RPS:         10,
		Concurrency: 3,
		MaxRequests: 200,
	}

	eng := NewBurst(cfg, 3*time.Second, 50, 500*time.Millisecond, 1*time.Second)
	client := httpclient.NewClient(5*time.Second, &profile.Profile{Name: "test", Headers: map[string]string{}}, nil, "", false)
	defer client.Close()

	collector := observe.NewCollector()

	start := time.Now()
	err := eng.Run(context.Background(), client, collector)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if elapsed > 5*time.Second {
		t.Errorf("expected completion within ~3s, got %v", elapsed)
	}

	m := collector.Snapshot()
	if m.TotalRequests < 20 {
		t.Errorf("expected at least 20 requests, got %d", m.TotalRequests)
	}
}
