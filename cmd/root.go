package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "scm",
	Short: "macOS Supply-Chain Integrity Monitor",
	Long: `A developer-focused security tool that monitors package managers
and tracks system changes to detect supply-chain attacks.

Supports: Homebrew, npm, pip, cargo, and go install.`,
	Version: "0.1.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/scm/config.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
	rootCmd.PersistentFlags().String("db", "", "database path (default is $HOME/.config/scm/scm.db)")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("database", rootCmd.PersistentFlags().Lookup("db"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configDir := fmt.Sprintf("%s/.config/scm", home)
		os.MkdirAll(configDir, 0755)

		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}

	// Set defaults
	setDefaults()
}

func setDefaults() {
	home, _ := os.UserHomeDir()

	viper.SetDefault("database", fmt.Sprintf("%s/.config/scm/scm.db", home))
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_file", fmt.Sprintf("%s/.config/scm/scm.log", home))

	viper.SetDefault("package_managers", []string{"homebrew", "npm", "pip", "cargo", "go"})

	viper.SetDefault("watch_paths", []string{
		"/usr/local/bin",
		"/opt/homebrew/bin",
		fmt.Sprintf("%s/.cargo/bin", home),
		fmt.Sprintf("%s/go/bin", home),
	})

	viper.SetDefault("alerts.risk_threshold", 70)
}
