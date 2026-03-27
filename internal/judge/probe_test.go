package judge

import (
	"testing"

	"github.com/House-lovers7/edge-checker/internal/engine"
)

func TestCheckFalsePositive_AllPass(t *testing.T) {
	probes := []engine.ProbeResult{
		{Type: "unaffected", Path: "/api/health", Method: "GET", StatusCode: 200, Expected: 200},
		{Type: "unaffected", Path: "/api/status", Method: "GET", StatusCode: 200, Expected: 200},
	}

	r := CheckFalsePositive(probes)
	if r.Status != StatusPass {
		t.Errorf("expected PASS, got %s: %s", r.Status, r.Message)
	}
}

func TestCheckFalsePositive_SomeFail(t *testing.T) {
	probes := []engine.ProbeResult{
		{Type: "unaffected", Path: "/api/health", Method: "GET", StatusCode: 200, Expected: 200},
		{Type: "unaffected", Path: "/api/status", Method: "GET", StatusCode: 403, Expected: 200},
	}

	r := CheckFalsePositive(probes)
	if r.Status != StatusFail {
		t.Errorf("expected FAIL, got %s: %s", r.Status, r.Message)
	}
}

func TestCheckFalsePositive_NoProbes(t *testing.T) {
	r := CheckFalsePositive(nil)
	if r.Status != StatusSkip {
		t.Errorf("expected SKIP, got %s", r.Status)
	}
}

func TestCheckFalsePositive_IgnoresBypassProbes(t *testing.T) {
	probes := []engine.ProbeResult{
		{Type: "bypass", Path: "/api/login", Method: "GET", StatusCode: 403, Expected: 200},
	}

	r := CheckFalsePositive(probes)
	if r.Status != StatusSkip {
		t.Errorf("expected SKIP (no unaffected probes), got %s", r.Status)
	}
}

func TestCheckBypass_AllPass(t *testing.T) {
	probes := []engine.ProbeResult{
		{Type: "bypass", Path: "/assets/sample.jpg", Method: "GET", StatusCode: 200, Expected: 200},
	}

	r := CheckBypass(probes)
	if r.Status != StatusPass {
		t.Errorf("expected PASS, got %s: %s", r.Status, r.Message)
	}
}

func TestCheckBypass_Blocked(t *testing.T) {
	probes := []engine.ProbeResult{
		{Type: "bypass", Path: "/assets/sample.jpg", Method: "GET", StatusCode: 403, Expected: 200},
	}

	r := CheckBypass(probes)
	if r.Status != StatusFail {
		t.Errorf("expected FAIL (bypass was blocked), got %s: %s", r.Status, r.Message)
	}
}

func TestCheckBypass_NoProbes(t *testing.T) {
	r := CheckBypass(nil)
	if r.Status != StatusSkip {
		t.Errorf("expected SKIP, got %s", r.Status)
	}
}

func TestCheckBypass_WithError(t *testing.T) {
	probes := []engine.ProbeResult{
		{Type: "bypass", Path: "/assets/sample.jpg", Method: "GET", Error: "connection refused", Expected: 200},
	}

	r := CheckBypass(probes)
	if r.Status != StatusFail {
		t.Errorf("expected FAIL (error = not passing), got %s", r.Status)
	}
}
