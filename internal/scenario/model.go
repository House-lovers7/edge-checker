package scenario

import "time"

// Scenario represents a complete WAF/CDN defense verification scenario.
type Scenario struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Target      Target            `yaml:"target"`
	Execution   Execution         `yaml:"execution"`
	Rate        Rate              `yaml:"rate"`
	Profile     string            `yaml:"profile"`
	Headers     map[string]string `yaml:"headers"`
	Query       Query             `yaml:"query"`
	Safety      Safety            `yaml:"safety"`
	Expect      Expect            `yaml:"expect"`
	Output      Output            `yaml:"output"`
}

type Target struct {
	BaseURL string `yaml:"base_url"`
	Host    string `yaml:"host"`
	Path    string `yaml:"path"`
	Method  string `yaml:"method"`
}

// FullURL returns the complete URL for the target.
func (t Target) FullURL() string {
	return t.BaseURL + t.Path
}

type Execution struct {
	Mode        string `yaml:"mode"`
	Duration    string `yaml:"duration"`
	Timeout     string `yaml:"timeout"`
	Concurrency int    `yaml:"concurrency"`
}

// ParseDuration parses the duration string into time.Duration.
func (e Execution) ParseDuration() (time.Duration, error) {
	return time.ParseDuration(e.Duration)
}

// ParseTimeout parses the timeout string into time.Duration.
func (e Execution) ParseTimeout() (time.Duration, error) {
	return time.ParseDuration(e.Timeout)
}

type Rate struct {
	RPS      int            `yaml:"rps"`
	Burst    *BurstConfig   `yaml:"burst,omitempty"`
	Ramp     *RampConfig    `yaml:"ramp,omitempty"`
	Cooldown *CooldownConfig `yaml:"cooldown,omitempty"`
}

type BurstConfig struct {
	SpikeRPS      int    `yaml:"spike_rps"`
	SpikeDuration string `yaml:"spike_duration"`
	Interval      string `yaml:"interval"`
}

type RampConfig struct {
	StartRPS     int    `yaml:"start_rps"`
	EndRPS       int    `yaml:"end_rps"`
	StepDuration string `yaml:"step_duration"`
}

type CooldownConfig struct {
	ActiveDuration string `yaml:"active_duration"`
	ActiveRPS      int    `yaml:"active_rps"`
	WaitDuration   string `yaml:"wait_duration"`
	ProbeRPS       int    `yaml:"probe_rps"`
}

type Query struct {
	Static       map[string]string `yaml:"static"`
	RandomSuffix bool              `yaml:"random_suffix"`
}

type Safety struct {
	Environment      string   `yaml:"environment"`
	MaxTotalRequests int      `yaml:"max_total_requests"`
	AllowHosts       []string `yaml:"allow_hosts"`
}

type Expect struct {
	BlockStatusCodes     []int              `yaml:"block_status_codes"`
	BlockWithinSeconds   int                `yaml:"block_within_seconds"`
	MinBlockRatio        float64            `yaml:"min_block_ratio"`
	UnaffectedPaths      []UnaffectedPath   `yaml:"unaffected_paths"`
	BypassScenarios      []BypassScenario   `yaml:"bypass_scenarios"`
}

type UnaffectedPath struct {
	Path         string `yaml:"path"`
	Method       string `yaml:"method"`
	ExpectStatus int    `yaml:"expect_status"`
}

type BypassScenario struct {
	Path         string            `yaml:"path"`
	Headers      map[string]string `yaml:"headers"`
	ExpectStatus int               `yaml:"expect_status"`
}

type Output struct {
	JSON     string `yaml:"json"`
	Markdown string `yaml:"markdown"`
}

// Valid execution modes.
const (
	ModeConstant     = "constant"
	ModeBurst        = "burst"
	ModeRamp         = "ramp"
	ModeCooldownCheck = "cooldown-check"
)

// ValidModes returns all valid execution mode names.
func ValidModes() []string {
	return []string{ModeConstant, ModeBurst, ModeRamp, ModeCooldownCheck}
}
