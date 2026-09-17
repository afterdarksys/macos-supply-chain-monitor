package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/afterdark/supply-chain-monitor/internal/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the monitoring daemon",
	Long:  `Start, stop, or check the status of the supply-chain monitoring daemon.`,
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the monitoring daemon",
	Long:  `Start the background daemon that monitors package managers and system changes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🚀 Starting Supply-Chain Monitor daemon...")

		dbPath := viper.GetString("database")
		fmt.Printf("📊 Database: %s\n", dbPath)

		d, err := daemon.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize daemon: %w", err)
		}

		if err := d.Start(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		fmt.Println("✅ Daemon started successfully")
		fmt.Println("📡 Monitoring package managers:", viper.GetStringSlice("package_managers"))
		fmt.Println("\n💡 Use 'scm status' to check activity")
		fmt.Println("💡 Use 'scm history' to view recent events")
		fmt.Println("\n⏸  Press Ctrl+C to stop")

		// Wait for interrupt
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\n\n🛑 Shutting down daemon...")
		if err := d.Stop(); err != nil {
			return fmt.Errorf("error during shutdown: %w", err)
		}

		fmt.Println("✅ Daemon stopped")
		return nil
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := daemon.ReadStatus(viper.GetString("database"))
		if err != nil {
			return fmt.Errorf("daemon unavailable: %w", err)
		}
		fmt.Printf("Daemon PID %d, %d watchers, heartbeat %s\n", status.PID, status.Watchers, status.Updated.Format("2006-01-02T15:04:05Z07:00"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
}
