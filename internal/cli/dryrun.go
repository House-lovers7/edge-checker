package cli

import (
	"fmt"
	"time"

	"github.com/House-lovers7/edge-checker/internal/safety"
	"github.com/House-lovers7/edge-checker/internal/scenario"
	"github.com/House-lovers7/edge-checker/internal/version"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Show execution plan without sending requests",
	Long:  "Load and validate a scenario, display the full execution plan, but do not send any requests.",
	RunE:  runDryRun,
}

func init() {
	dryRunCmd.Flags().StringP("file", "f", "", "Path to scenario YAML file (required)")
	dryRunCmd.MarkFlagRequired("file")
	dryRunCmd.Flags().Bool("allow-production", false, "Allow production environment (for plan display)")
	rootCmd.AddCommand(dryRunCmd)
}

func runDryRun(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	allowProd, _ := cmd.Flags().GetBool("allow-production")

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	fmt.Printf("edge-checker %s (dry-run)\n\n", version.Version)

	// Load
	fmt.Printf("Loading scenario: %s ... ", filePath)
	s, err := scenario.Load(filePath)
	if err != nil {
		red.Println("FAILED")
		return fmt.Errorf("load error: %w", err)
	}
	green.Println("OK")

	// Validate
	if errs := scenario.Validate(s); len(errs) > 0 {
		red.Printf("Validation failed with %d error(s):\n", len(errs))
		for _, e := range errs {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("validation failed")
	}

	// Safety checks
	fmt.Println("Safety checks:")

	fmt.Print("  Host allowlist .... ")
	if err := safety.CheckHost(s.Target.BaseURL, s.Safety.AllowHosts); err != nil {
		red.Println("BLOCKED")
		return err
	}
	green.Println("OK")

	duration, _ := s.Execution.ParseDuration()

	fmt.Print("  Request limits .... ")
	estimated := estimateRequests(s, duration)
	if limitErrs := safety.CheckLimits(s.Safety.MaxTotalRequests, s.Execution.Concurrency, s.Rate.RPS, duration, estimated); len(limitErrs) > 0 {
		red.Println("BLOCKED")
		for _, e := range limitErrs {
			fmt.Printf("    %v\n", e)
		}
		return fmt.Errorf("safety limits exceeded")
	}
	green.Println("OK")

	fmt.Print("  Environment ...... ")
	if err := safety.CheckProduction(s.Safety.Environment, allowProd); err != nil {
		red.Println("BLOCKED")
		return err
	}
	green.Printf("%s OK\n", s.Safety.Environment)

	fmt.Println()

	// Execution plan
	cyan.Println("=== Execution Plan ===")
	fmt.Println()
	fmt.Printf("  Scenario:    %s\n", s.Name)
	if s.Description != "" {
		fmt.Printf("  Description: %s\n", s.Description)
	}
	fmt.Printf("  Target:      %s %s\n", s.Target.Method, s.Target.FullURL())
	if s.Target.Host != "" {
		fmt.Printf("  Host header: %s\n", s.Target.Host)
	}
	fmt.Printf("  Mode:        %s\n", s.Execution.Mode)
	fmt.Printf("  Duration:    %s\n", s.Execution.Duration)
	fmt.Printf("  Concurrency: %d\n", s.Execution.Concurrency)
	fmt.Printf("  Profile:     %s\n", profileOrDefault(s.Profile))
	fmt.Println()

	// Mode-specific details
	printModeDetails(s)

	// Estimated totals (already calculated above for limits check)
	fmt.Printf("  Estimated total requests: %d\n", estimated)
	if s.Safety.MaxTotalRequests > 0 {
		fmt.Printf("  Max total requests:       %d\n", s.Safety.MaxTotalRequests)
	}
	fmt.Printf("  Allow hosts:              %v\n", s.Safety.AllowHosts)
	fmt.Println()

	// Judgment criteria
	cyan.Println("  Judgment criteria:")
	if len(s.Expect.BlockStatusCodes) > 0 {
		fmt.Printf("    - Expect status %v", s.Expect.BlockStatusCodes)
		if s.Expect.BlockWithinSeconds > 0 {
			fmt.Printf(" within %ds", s.Expect.BlockWithinSeconds)
		}
		fmt.Println()
	}
	if s.Expect.MinBlockRatio > 0 {
		fmt.Printf("    - Expect block ratio >= %.0f%%\n", s.Expect.MinBlockRatio*100)
	}
	if len(s.Expect.UnaffectedPaths) > 0 {
		for _, p := range s.Expect.UnaffectedPaths {
			fmt.Printf("    - Expect %s %s unaffected (status %d)\n", p.Method, p.Path, p.ExpectStatus)
		}
	}
	if len(s.Expect.BypassScenarios) > 0 {
		fmt.Printf("    - %d bypass scenario(s) should remain unblocked\n", len(s.Expect.BypassScenarios))
	}
	fmt.Println()

	yellow.Println("  [DRY-RUN] No requests will be sent.")

	return nil
}

func profileOrDefault(p string) string {
	if p == "" {
		return "(none)"
	}
	return p
}

func printModeDetails(s *scenario.Scenario) {
	switch s.Execution.Mode {
	case scenario.ModeConstant:
		fmt.Printf("  Rate:        %d rps (constant)\n", s.Rate.RPS)
	case scenario.ModeBurst:
		fmt.Printf("  Base rate:   %d rps\n", s.Rate.RPS)
		if s.Rate.Burst != nil {
			fmt.Printf("  Spike:       %d rps for %s every %s\n",
				s.Rate.Burst.SpikeRPS, s.Rate.Burst.SpikeDuration, s.Rate.Burst.Interval)
		}
	case scenario.ModeRamp:
		if s.Rate.Ramp != nil {
			fmt.Printf("  Ramp:        %d → %d rps (step every %s)\n",
				s.Rate.Ramp.StartRPS, s.Rate.Ramp.EndRPS, s.Rate.Ramp.StepDuration)
		}
	case scenario.ModeCooldownCheck:
		if s.Rate.Cooldown != nil {
			fmt.Printf("  Active:      %d rps for %s\n", s.Rate.Cooldown.ActiveRPS, s.Rate.Cooldown.ActiveDuration)
			fmt.Printf("  Cooldown:    wait %s, probe at %d rps\n", s.Rate.Cooldown.WaitDuration, s.Rate.Cooldown.ProbeRPS)
		}
	}
}

func estimateRequests(s *scenario.Scenario, duration time.Duration) int {
	secs := int(duration.Seconds())

	switch s.Execution.Mode {
	case scenario.ModeBurst:
		if s.Rate.Burst != nil {
			interval, _ := time.ParseDuration(s.Rate.Burst.Interval)
			spikeDur, _ := time.ParseDuration(s.Rate.Burst.SpikeDuration)
			if interval > 0 {
				cycles := secs / int(interval.Seconds())
				baseReqs := s.Rate.RPS * secs
				spikeReqs := cycles * s.Rate.Burst.SpikeRPS * int(spikeDur.Seconds())
				return baseReqs + spikeReqs
			}
		}
	case scenario.ModeRamp:
		if s.Rate.Ramp != nil {
			avgRPS := (s.Rate.Ramp.StartRPS + s.Rate.Ramp.EndRPS) / 2
			return avgRPS * secs
		}
	case scenario.ModeCooldownCheck:
		if s.Rate.Cooldown != nil {
			activeDur, _ := time.ParseDuration(s.Rate.Cooldown.ActiveDuration)
			waitDur, _ := time.ParseDuration(s.Rate.Cooldown.WaitDuration)
			active := s.Rate.Cooldown.ActiveRPS * int(activeDur.Seconds())
			probe := s.Rate.Cooldown.ProbeRPS * int(waitDur.Seconds())
			return active + probe
		}
	}

	return s.Rate.RPS * secs
}
