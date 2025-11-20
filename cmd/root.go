package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	debug    bool
	logLevel string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "dnsbalancer",
	Short: "A lightweight UDP DNS load balancer",
	Long: `dnsbalancer is a simple, high-performance DNS load balancer that
distributes queries across multiple DNS backend servers using round-robin
with health checking capabilities.

Perfect for homelab and production environments where you need reliable
DNS service with automatic failover.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml, then /etc/dnsbalancer/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging to console")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (debug, info, warn, error)")
}

// findConfigFile searches for config file in priority order
func findConfigFile() string {
	// 1. Command line flag takes precedence
	if cfgFile != "" {
		return cfgFile
	}

	// 2. Current directory
	if _, err := os.Stat("./config.yaml"); err == nil {
		return "./config.yaml"
	}

	// 3. System config directory
	if _, err := os.Stat("/etc/dnsbalancer/config.yaml"); err == nil {
		return "/etc/dnsbalancer/config.yaml"
	}

	// No config file found, will use defaults
	return ""
}
