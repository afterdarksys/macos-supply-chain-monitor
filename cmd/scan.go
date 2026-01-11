package cmd

import (
	"fmt"

	"github.com/afterdark/supply-chain-monitor/internal/scanner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan current system state",
	Long:  `Perform a one-time scan of installed packages and binaries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔍 Scanning system...")

		dbPath := viper.GetString("database")
		s := scanner.New(dbPath)

		brewOnly, _ := cmd.Flags().GetBool("brew")
		npmOnly, _ := cmd.Flags().GetBool("npm")

		var managers []string
		if brewOnly {
			managers = []string{"homebrew"}
		} else if npmOnly {
			managers = []string{"npm"}
		} else {
			managers = viper.GetStringSlice("package_managers")
		}

		results, err := s.ScanPackageManagers(managers)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		printScanResults(results)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().Bool("brew", false, "Scan only Homebrew")
	scanCmd.Flags().Bool("npm", false, "Scan only npm")
	scanCmd.Flags().Bool("pip", false, "Scan only pip")
	scanCmd.Flags().Bool("verify", false, "Verify all binary hashes")
}

func printScanResults(results map[string]int) {
	fmt.Println("\n📊 Scan Results:")
	fmt.Println("─────────────────────────────────────────")

	total := 0
	for manager, count := range results {
		fmt.Printf("  %s: %d packages\n", manager, count)
		total += count
	}

	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("  Total: %d packages\n\n", total)
}
