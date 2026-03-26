package scenario

import "testing"

func validScenario() *Scenario {
	return &Scenario{
		Name: "test-scenario",
		Target: Target{
			BaseURL: "https://staging.example.com",
			Path:    "/test",
			Method:  "GET",
		},
		Execution: Execution{
			Mode:        "constant",
			Duration:    "10s",
			Timeout:     "5s",
			Concurrency: 3,
		},
		Rate: Rate{RPS: 10},
		Safety: Safety{
			Environment:      "staging",
			MaxTotalRequests: 1000,
			AllowHosts:       []string{"staging.example.com"},
		},
		Expect: Expect{
			BlockStatusCodes: []int{403, 429},
		},
	}
}

func TestValidate_ValidScenario(t *testing.T) {
	s := validScenario()
	errs := Validate(s)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidate_MissingName(t *testing.T) {
	s := validScenario()
	s.Name = ""
	errs := Validate(s)
	if !hasField(errs, "name") {
		t.Error("expected error for missing name")
	}
}

func TestValidate_InvalidMode(t *testing.T) {
	s := validScenario()
	s.Execution.Mode = "invalid"
	errs := Validate(s)
	if !hasField(errs, "execution.mode") {
		t.Error("expected error for invalid mode")
	}
}

func TestValidate_InvalidDuration(t *testing.T) {
	s := validScenario()
	s.Execution.Duration = "notaduration"
	errs := Validate(s)
	if !hasField(errs, "execution.duration") {
		t.Error("expected error for invalid duration")
	}
}

func TestValidate_HostNotInAllowList(t *testing.T) {
	s := validScenario()
	s.Safety.AllowHosts = []string{"other.example.com"}
	errs := Validate(s)
	if !hasField(errs, "safety.allow_hosts") {
		t.Error("expected error for host not in allow_hosts")
	}
}

func TestValidate_BurstModeRequiresBurstConfig(t *testing.T) {
	s := validScenario()
	s.Execution.Mode = "burst"
	s.Rate.Burst = nil
	errs := Validate(s)
	if !hasField(errs, "rate.burst") {
		t.Error("expected error for missing burst config")
	}
}

func TestValidate_MissingBlockStatusCodes(t *testing.T) {
	s := validScenario()
	s.Expect.BlockStatusCodes = nil
	errs := Validate(s)
	if !hasField(errs, "expect.block_status_codes") {
		t.Error("expected error for missing block_status_codes")
	}
}

func hasField(errs []ValidationError, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}
