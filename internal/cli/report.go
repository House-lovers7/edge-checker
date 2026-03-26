package cli

import (
	"fmt"

	"github.com/House-lovers7/edge-checker/internal/output"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a Markdown report from results",
	Long:  "Load a JSON result file and generate a Markdown report suitable for documentation or stakeholder review.",
	RunE:  runReport,
}

func init() {
	reportCmd.Flags().StringP("result", "r", "", "Path to result JSON file (required)")
	reportCmd.MarkFlagRequired("result")
	reportCmd.Flags().StringP("output", "o", "", "Output path for Markdown report (default: results/report-<timestamp>.md)")
	rootCmd.AddCommand(reportCmd)
}

func runReport(cmd *cobra.Command, args []string) error {
	resultPath, _ := cmd.Flags().GetString("result")
	outputPath, _ := cmd.Flags().GetString("output")

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	result, err := loadResult(resultPath)
	if err != nil {
		return err
	}

	if outputPath == "" {
		outputPath = fmt.Sprintf("results/report-%s.md", result.StartedAt.Format("20060102-150405"))
	}

	savedPath, err := output.WriteMarkdown(result, outputPath)
	if err != nil {
		red.Printf("Failed to generate report: %v\n", err)
		return err
	}

	green.Printf("Report saved: %s\n", savedPath)
	return nil
}
