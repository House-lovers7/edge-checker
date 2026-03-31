package cli

import (
	"fmt"
	"strconv"

	"github.com/House-lovers7/edge-checker/internal/store"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View past test run history",
	Long:  "List past test runs stored in the SQLite database, or view details of a specific run.",
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().IntP("limit", "n", 20, "Number of recent runs to show")
	historyCmd.Flags().String("db", "", "Path to SQLite database (default: results/history.db)")
	historyCmd.Flags().Int64("id", 0, "Show details for a specific run ID")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	dbPath, _ := cmd.Flags().GetString("db")
	limit, _ := cmd.Flags().GetInt("limit")
	showID, _ := cmd.Flags().GetInt64("id")

	db, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// If --id is specified, show details for that run
	if showID > 0 {
		return showRunDetail(db, showID)
	}

	// List recent runs
	runs, err := db.List(limit)
	if err != nil {
		return err
	}

	if len(runs) == 0 {
		fmt.Println("No test runs found in history.")
		fmt.Println("Run a scenario with 'edge-checker run' to start recording history.")
		return nil
	}

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	fmt.Printf("%-4s  %-28s  %-20s  %-8s  %-10s  %-8s  %s\n",
		"ID", "Scenario", "Started", "Duration", "Mode", "Verdict", "Requests")
	fmt.Println("----  ----------------------------  --------------------  --------  ----------  --------  --------")

	for _, r := range runs {
		startShort := r.StartedAt
		if len(startShort) > 19 {
			startShort = startShort[:19]
		}

		verdictStr := r.Verdict
		switch r.Verdict {
		case "PASS":
			verdictStr = green.Sprint("PASS")
		case "FAIL":
			verdictStr = red.Sprint("FAIL")
		}

		scenarioName := r.ScenarioName
		if len(scenarioName) > 28 {
			scenarioName = scenarioName[:25] + "..."
		}

		fmt.Printf("%-4d  %-28s  %-20s  %-8s  %-10s  %-8s  %d\n",
			r.ID, scenarioName, startShort, r.Duration, r.Mode, verdictStr, r.TotalRequests)
	}

	fmt.Printf("\nShowing %d most recent runs. Use 'edge-checker history --id <ID>' for details.\n", len(runs))
	return nil
}

func showRunDetail(db *store.Store, id int64) error {
	// If first arg is a number, use it as ID
	result, err := db.Get(id)
	if err != nil {
		return err
	}

	printSummary(result, result.Interrupted)
	if result.Verdict != nil {
		printVerdict(result.Verdict)
	}

	return nil
}

// ParseRunID tries to parse a run ID from command args.
func ParseRunID(args []string) (int64, bool) {
	if len(args) > 0 {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err == nil {
			return id, true
		}
	}
	return 0, false
}
