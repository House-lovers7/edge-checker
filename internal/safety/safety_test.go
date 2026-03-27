package safety

import (
	"testing"
	"time"
)

func TestCheckHost_Allowed(t *testing.T) {
	err := CheckHost("https://staging.example.com/path", []string{"staging.example.com"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCheckHost_NotAllowed(t *testing.T) {
	err := CheckHost("https://evil.com/path", []string{"staging.example.com"})
	if err == nil {
		t.Error("expected error for disallowed host")
	}
}

func TestCheckHost_EmptyURL(t *testing.T) {
	err := CheckHost("", []string{"staging.example.com"})
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestCheckLimits_WithinBounds(t *testing.T) {
	errs := CheckLimits(5000, 10, 50, 30*time.Second, 1500)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestCheckLimits_ExceedsConcurrency(t *testing.T) {
	errs := CheckLimits(1000, 200, 50, 30*time.Second, 1500)
	if len(errs) == 0 {
		t.Error("expected error for exceeding concurrency")
	}
}

func TestCheckLimits_ExceedsRPS(t *testing.T) {
	errs := CheckLimits(1000, 10, 2000, 30*time.Second, 60000)
	if len(errs) == 0 {
		t.Error("expected error for exceeding RPS")
	}
}

func TestCheckLimits_ExceedsDuration(t *testing.T) {
	errs := CheckLimits(1000, 10, 50, 60*time.Minute, 180000)
	if len(errs) == 0 {
		t.Error("expected error for exceeding duration")
	}
}

func TestCheckLimits_EstimatedExceedsMax(t *testing.T) {
	errs := CheckLimits(100, 5, 10, 30*time.Second, 300)
	found := false
	for _, e := range errs {
		if e.Error() == "estimated total requests 300 exceeds max_total_requests 100" {
			found = true
		}
	}
	if !found {
		t.Error("expected error for estimated exceeding max_total_requests")
	}
}

func TestCheckProduction_Staging(t *testing.T) {
	err := CheckProduction("staging", false)
	if err != nil {
		t.Errorf("expected no error for staging, got %v", err)
	}
}

func TestCheckProduction_ProdWithoutFlag(t *testing.T) {
	err := CheckProduction("production", false)
	if err == nil {
		t.Error("expected error for production without flag")
	}
}

func TestCheckProduction_ProdWithFlag(t *testing.T) {
	err := CheckProduction("production", true)
	if err != nil {
		t.Errorf("expected no error for production with flag, got %v", err)
	}
}

func TestCheckDNS_Resolvable(t *testing.T) {
	err := CheckDNS("https://google.com")
	if err != nil {
		t.Skipf("skipping DNS test (network unavailable): %v", err)
	}
}

func TestCheckDNS_IPAddress(t *testing.T) {
	err := CheckDNS("https://127.0.0.1/test")
	if err != nil {
		t.Errorf("expected no error for IP address, got %v", err)
	}
}

func TestCheckDNS_Unresolvable(t *testing.T) {
	err := CheckDNS("https://this-host-definitely-does-not-exist-xyz123.invalid")
	if err == nil {
		t.Error("expected error for unresolvable host")
	}
}
