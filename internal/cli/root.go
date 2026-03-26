package cli

import (
	"fmt"
	"os"

	"github.com/House-lovers7/edge-checker/internal/version"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	noColor bool
)

var rootCmd = &cobra.Command{
	Use:     "edge-checker",
	Short:   "WAF/CDN defense configuration verification harness",
	Long:    "edge-checker is a CLI tool that verifies WAF/CDN defense settings work as expected by reproducing controlled traffic patterns and judging whether defenses triggered correctly.",
	Version: version.Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			color.NoColor = true
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	rootCmd.SetVersionTemplate(fmt.Sprintf("edge-checker %s (commit: %s, built: %s)\n", version.Version, version.Commit, version.Date))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func IsVerbose() bool {
	return verbose
}
