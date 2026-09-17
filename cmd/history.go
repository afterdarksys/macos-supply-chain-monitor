package cmd

import (
	"fmt"
	"time"

	"github.com/afterdark/supply-chain-monitor/internal/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View installation history",
	Long:  `Display a timeline of package installations and system changes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := viper.GetString("database")

		store, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer store.Close()

		duration, _ := cmd.Flags().GetString("last")
		limit, _ := cmd.Flags().GetInt("limit")
		riskFilter, _ := cmd.Flags().GetString("risk")

		events, err := store.GetRecentEvents(parseDuration(duration), limit, riskFilter)
		if err != nil {
			return fmt.Errorf("failed to retrieve events: %w", err)
		}

		if len(events) == 0 {
			fmt.Println("📭 No events found")
			fmt.Println("💡 Make sure the daemon is running: scm daemon start")
			return nil
		}

		printHistoryTable(events)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().String("last", "24h", "Time range (e.g., 1h, 24h, 7d)")
	historyCmd.Flags().Int("limit", 50, "Maximum number of events to show")
	historyCmd.Flags().String("risk", "", "Filter by risk level (low, medium, high)")
	historyCmd.Flags().Bool("today", false, "Show only today's events")
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func printHistoryTable(events []db.Event) {
	fmt.Println("\n┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Supply-Chain Activity                                            │")
	fmt.Print("└─────────────────────────────────────────────────────────────────┘\n\n")

	var highCount, mediumCount, lowCount int

	for _, event := range events {
		var riskIcon, riskColor string

		switch {
		case event.RiskScore >= 70:
			riskIcon = "🔴"
			riskColor = "HIGH RISK"
			highCount++
		case event.RiskScore >= 40:
			riskIcon = "🟡"
			riskColor = "MEDIUM RISK"
			mediumCount++
		default:
			riskIcon = "🟢"
			riskColor = "LOW RISK"
			lowCount++
		}

		fmt.Printf("%s %s\n", riskIcon, riskColor)
		fmt.Printf("  [%s] %s %s\n",
			event.Timestamp.Format("15:04:05"),
			event.PackageManager,
			event.PackageName)

		if event.PackageVersion != "" {
			fmt.Printf("  └─ Version: %s\n", event.PackageVersion)
		}

		if event.RiskScore > 0 {
			fmt.Printf("  └─ Risk Score: %d/100\n", event.RiskScore)
		}

		fmt.Println()
	}

	fmt.Println("─────────────────────────────────────────────────────────────────")
	fmt.Printf("Risk Summary: %d High, %d Medium, %d Low\n", highCount, mediumCount, lowCount)
	fmt.Println()
}
