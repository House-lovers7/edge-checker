package cli

import (
	"fmt"

	"github.com/House-lovers7/edge-checker/internal/scenario"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a scenario YAML file",
	Long:  "Check a scenario YAML file for structural and semantic errors without executing it.",
	RunE:  runValidate,
}

func init() {
	validateCmd.Flags().StringP("file", "f", "", "Path to scenario YAML file (required)")
	validateCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	fmt.Printf("Validating: %s\n", filePath)

	s, err := scenario.Load(filePath)
	if err != nil {
		red.Printf("  LOAD ERROR: %v\n", err)
		return fmt.Errorf("validation failed")
	}

	errs := scenario.Validate(s)
	if len(errs) == 0 {
		green.Println("  OK — no errors found")
		fmt.Printf("  Scenario: %s\n", s.Name)
		fmt.Printf("  Target:   %s %s\n", s.Target.Method, s.Target.FullURL())
		fmt.Printf("  Mode:     %s\n", s.Execution.Mode)
		fmt.Printf("  Duration: %s\n", s.Execution.Duration)
		fmt.Printf("  RPS:      %d\n", s.Rate.RPS)
		return nil
	}

	red.Printf("  FAILED — %d error(s) found:\n", len(errs))
	for _, e := range errs {
		fmt.Printf("    - %s\n", e)
	}
	return fmt.Errorf("validation failed with %d error(s)", len(errs))
}
