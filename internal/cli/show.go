package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/House-lovers7/edge-checker/internal/judge"
	"github.com/House-lovers7/edge-checker/internal/output"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display results from a previous run",
	Long:  "Load a JSON result file and display a formatted summary in the terminal.",
	RunE:  runShow,
}

func init() {
	showCmd.Flags().StringP("result", "r", "", "Path to result JSON file (required)")
	showCmd.MarkFlagRequired("result")
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	resultPath, _ := cmd.Flags().GetString("result")

	result, err := loadResult(resultPath)
	if err != nil {
		return err
	}

	printSummary(result, result.Interrupted)

	if result.Verdict != nil {
		printVerdict(result.Verdict)
	}

	return nil
}

func loadResult(path string) (*output.Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read result file %q: %w", path, err)
	}

	var result output.Result
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse result JSON %q: %w", path, err)
	}

	return &result, nil
}

// Suppress unused import warnings
var _ = color.New
var _ = judge.StatusPass
