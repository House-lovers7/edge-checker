package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/House-lovers7/edge-checker/internal/engine"
	"github.com/House-lovers7/edge-checker/internal/httpclient"
	"github.com/House-lovers7/edge-checker/internal/judge"
	"github.com/House-lovers7/edge-checker/internal/observe"
	"github.com/House-lovers7/edge-checker/internal/output"
	"github.com/House-lovers7/edge-checker/internal/profile"
	"github.com/House-lovers7/edge-checker/internal/safety"
	"github.com/House-lovers7/edge-checker/internal/scenario"
	"github.com/House-lovers7/edge-checker/internal/version"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a verification scenario",
	Long:  "Run a WAF/CDN defense verification scenario and save the results.",
	RunE:  runRun,
}

func init() {
	runCmd.Flags().StringP("file", "f", "", "Path to scenario YAML file (required)")
	runCmd.MarkFlagRequired("file")
	runCmd.Flags().Bool("allow-production", false, "Allow execution against production environment")
	runCmd.Flags().Bool("insecure", false, "Skip TLS certificate verification (use with caution)")
	runCmd.Flags().StringP("output", "o", "", "Override output JSON path")
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	allowProd, _ := cmd.Flags().GetBool("allow-production")
	insecure, _ := cmd.Flags().GetBool("insecure")
	outputPath, _ := cmd.Flags().GetString("output")

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)

	fmt.Printf("edge-checker %s\n\n", version.Version)

	// 1. Load scenario
	fmt.Printf("Loading scenario: %s ... ", filePath)
	s, err := scenario.Load(filePath)
	if err != nil {
		red.Println("FAILED")
		return fmt.Errorf("load error: %w", err)
	}
	green.Println("OK")

	// 2. Validate
	if errs := scenario.Validate(s); len(errs) > 0 {
		red.Printf("Validation failed with %d error(s):\n", len(errs))
		for _, e := range errs {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("validation failed")
	}

	// 3. Safety checks
	fmt.Println("Safety checks:")

	// Host allowlist
	fmt.Print("  Host allowlist .... ")
	if err := safety.CheckHost(s.Target.BaseURL, s.Safety.AllowHosts); err != nil {
		red.Println("BLOCKED")
		return err
	}
	green.Println("OK")

	// Parse duration for limits check
	duration, _ := s.Execution.ParseDuration()

	// Request limits
	fmt.Print("  Request limits .... ")
	estimated := estimateRequests(s, duration)
	if limitErrs := safety.CheckLimits(s.Safety.MaxTotalRequests, s.Execution.Concurrency, s.Rate.RPS, duration, estimated); len(limitErrs) > 0 {
		red.Println("BLOCKED")
		for _, e := range limitErrs {
			fmt.Printf("    %v\n", e)
		}
		return fmt.Errorf("safety limits exceeded")
	}
	maxLabel := "unlimited"
	if s.Safety.MaxTotalRequests > 0 {
		maxLabel = fmt.Sprintf("%d", s.Safety.MaxTotalRequests)
	}
	green.Printf("OK (estimated: %d / max: %s)\n", estimated, maxLabel)

	// Environment
	fmt.Print("  Environment ...... ")
	if err := safety.CheckProduction(s.Safety.Environment, allowProd); err != nil {
		red.Println("BLOCKED")
		return err
	}
	green.Printf("%s OK\n", s.Safety.Environment)

	// DNS resolution
	fmt.Print("  DNS resolution ... ")
	if err := safety.CheckDNS(s.Target.BaseURL); err != nil {
		yellow.Printf("WARN (%v)\n", err)
	} else {
		green.Println("OK")
	}

	if insecure {
		yellow.Println("  TLS verify ....... DISABLED (--insecure)")
	}

	fmt.Println()

	// 4. Resolve profile
	prof := &profile.Profile{Name: "none", Headers: map[string]string{}}
	if s.Profile != "" {
		p, err := profile.GetProfile(s.Profile)
		if err != nil {
			return err
		}
		prof = p
	}

	// 5. Create HTTP client
	timeout, _ := s.Execution.ParseTimeout()
	client := httpclient.NewClient(timeout, prof, s.Headers, s.Target.Host, insecure)
	defer client.Close()

	// 6. Create engine
	eng, err := engine.New(s)
	if err != nil {
		return fmt.Errorf("engine creation failed: %w", err)
	}

	// 7. Setup context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupted := false
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		yellow.Println("\nGracefully shutting down... (press Ctrl+C again to force)")
		interrupted = true
		cancel()
		// Second signal = force exit
		<-sigCh
		os.Exit(1)
	}()

	// 8. Create collector + progress
	collector := observe.NewCollector()
	totalSeconds := int(duration.Seconds())
	progress := observe.NewProgress(collector, totalSeconds, os.Stdout, verbose)

	// Print run header
	fmt.Printf("Starting %s mode: %d rps x %s\n", s.Execution.Mode, s.Rate.RPS, s.Execution.Duration)
	fmt.Printf("  Target:  %s %s\n", s.Target.Method, s.Target.FullURL())
	fmt.Printf("  Profile: %s\n", prof.Name)
	fmt.Println()

	// 9. Start probe runner (unaffected paths + bypass scenarios)
	probeRunner := engine.NewProbeRunner(s.Target.BaseURL, s.Target.Host, 5*time.Second)
	probeDone := make(chan struct{})
	if len(s.Expect.UnaffectedPaths) > 0 || len(s.Expect.BypassScenarios) > 0 {
		fmt.Printf("  Probes:  %d unaffected path(s), %d bypass scenario(s)\n",
			len(s.Expect.UnaffectedPaths), len(s.Expect.BypassScenarios))
		go func() {
			defer close(probeDone)
			probeRunner.Run(ctx, client, s.Expect.UnaffectedPaths, s.Expect.BypassScenarios)
		}()
	} else {
		close(probeDone)
	}

	// 10. Progress ticker
	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				progress.Update()
			case <-ctx.Done():
				progress.Update()
				return
			}
		}
	}()

	// 10. Run engine
	if err := eng.Run(ctx, client, collector); err != nil {
		red.Printf("Engine error: %v\n", err)
	}

	// Wait for progress and probes to finish
	cancel() // ensure progress + probe goroutines stop
	<-progressDone
	<-probeDone
	progress.Finish()
	fmt.Println()

	// 11. Build result
	metrics := collector.Snapshot()
	result := output.BuildResult(
		s.Name,
		s.Description,
		output.TargetInfo{
			BaseURL: s.Target.BaseURL,
			Path:    s.Target.Path,
			Method:  s.Target.Method,
			Profile: prof.Name,
		},
		output.ExecutionInfo{
			Mode:        s.Execution.Mode,
			Duration:    s.Execution.Duration,
			Concurrency: s.Execution.Concurrency,
			RPS:         s.Rate.RPS,
			Environment: s.Safety.Environment,
		},
		metrics,
		interrupted,
	)

	// 12. Judge (skip full verdict if interrupted)
	if interrupted {
		result.Verdict = &judge.Verdict{
			Overall: judge.StatusSkip,
			Rules: []judge.RuleResult{{
				Name:    "Interrupted",
				Status:  judge.StatusSkip,
				Message: "Test was interrupted before completion — judgment is not reliable",
			}},
		}
	} else {
		result.Verdict = judge.JudgeWithProbes(metrics, &s.Expect, probeRunner.Results())
	}

	// 13. Print summary + verdict
	printSummary(result, interrupted)
	printVerdict(result.Verdict)

	// 14. Save JSON
	jsonPath := outputPath
	if jsonPath == "" {
		if s.Output.JSON != "" {
			jsonPath = s.Output.JSON
		} else {
			jsonPath = output.DefaultJSONPath(s.Name)
		}
	}

	savedPath, err := output.WriteJSON(result, jsonPath)
	if err != nil {
		red.Printf("Failed to save results: %v\n", err)
		return err
	}
	fmt.Printf("Results saved: %s\n", savedPath)

	// 15. Save Markdown if configured
	if s.Output.Markdown != "" {
		mdPath, err := output.WriteMarkdown(result, s.Output.Markdown)
		if err != nil {
			red.Printf("Failed to save Markdown report: %v\n", err)
		} else {
			fmt.Printf("Report saved: %s\n", mdPath)
		}
	}

	return nil
}

func printSummary(result *output.Result, interrupted bool) {
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	if interrupted {
		yellow.Println("--- Results (interrupted) ---")
	} else {
		fmt.Println("--- Results ---")
	}

	fmt.Printf("Total: %d requests in %s\n", result.Summary.TotalRequests, result.Duration)

	// Status codes
	statusParts := formatStatusLine(result.Summary.StatusCounts, result.Summary.TotalRequests)
	fmt.Printf("Status: %s\n", statusParts)

	if result.Summary.ErrorCount > 0 {
		fmt.Printf("Errors: %d\n", result.Summary.ErrorCount)
	}

	// Latency
	fmt.Printf("Latency: p50=%.0fms p95=%.0fms p99=%.0fms\n",
		result.Summary.P50LatencyMs,
		result.Summary.P95LatencyMs,
		result.Summary.P99LatencyMs,
	)
	fmt.Println()

	_ = green
}

func printVerdict(v *judge.Verdict) {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow)

	if v.Overall == judge.StatusPass {
		green.Println("--- Verdict: PASS ---")
	} else {
		red.Println("--- Verdict: FAIL ---")
	}

	for _, r := range v.Rules {
		var icon string
		switch r.Status {
		case judge.StatusPass:
			icon = color.GreenString("  ✓")
		case judge.StatusFail:
			icon = color.RedString("  ✗")
		case judge.StatusSkip:
			icon = color.YellowString("  -")
		}

		line := fmt.Sprintf("%s %-18s", icon, r.Name)
		if r.Actual != "" {
			line += fmt.Sprintf(": %s (expected: %s)", r.Actual, r.Expected)
		} else if r.Message != "" {
			line += fmt.Sprintf(": %s", r.Message)
		}
		fmt.Println(line)
	}
	fmt.Println()
	_ = yellow
}

func formatStatusLine(counts map[int]int, total int) string {
	codes := make([]int, 0, len(counts))
	for code := range counts {
		codes = append(codes, code)
	}
	sort.Ints(codes)

	parts := make([]string, 0, len(codes))
	for _, code := range codes {
		count := counts[code]
		pct := 0
		if total > 0 {
			pct = count * 100 / total
		}
		parts = append(parts, fmt.Sprintf("%d=%d (%d%%)", code, count, pct))
	}
	return strings.Join(parts, " ")
}
