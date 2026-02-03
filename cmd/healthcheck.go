package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/aram535/dnsbalancer/backend"
	"github.com/aram535/dnsbalancer/config"
)

var (
	testTimeout time.Duration
	testQuery   string
	testType    string
)

// healthcheckCmd represents the healthcheck command
var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Test backend DNS server connectivity",
	Long: `Perform a one-off health check against all configured backends.

This is useful for:
  - Testing backend connectivity before starting the server
  - Debugging backend issues
  - Validating DNS server responses

Example:
  dnsbalancer healthcheck
  dnsbalancer healthcheck --config /etc/dnsbalancer/config.yaml
  dnsbalancer healthcheck --timeout 5s --query example.com --type A`,
	RunE: runHealthcheck,
}

func init() {
	rootCmd.AddCommand(healthcheckCmd)

	healthcheckCmd.Flags().DurationVar(&testTimeout, "timeout", 3*time.Second, "timeout for health check query")
	healthcheckCmd.Flags().StringVar(&testQuery, "query", ".", "DNS query name to test")
	healthcheckCmd.Flags().StringVar(&testType, "type", "NS", "DNS query type (A, AAAA, NS, ANY)")
}

func runHealthcheck(cmd *cobra.Command, args []string) error {
	configFile := findConfigFile()
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if configFile != "" {
		fmt.Printf("Using config: %s\n", configFile)
	} else {
		fmt.Printf("Using default configuration\n")
	}

	fmt.Printf("Testing %d backends with query: %s (%s)\n", len(cfg.Backends), testQuery, testType)
	fmt.Printf("Timeout: %s\n\n", testTimeout)

	allHealthy := true

	for i, backendCfg := range cfg.Backends {
		b := backend.NewBackend(backendCfg.Address)
		
		fmt.Printf("[%d/%d] Testing %s ... ", i+1, len(cfg.Backends), b.Address)
		
		start := time.Now()
		err := b.HealthCheck(testQuery, testType, testTimeout)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("❌ FAILED (%.0fms)\n", elapsed.Seconds()*1000)
			fmt.Printf("      Error: %v\n", err)
			allHealthy = false
		} else {
			fmt.Printf("✅ OK (%.0fms)\n", elapsed.Seconds()*1000)
		}
	}

	fmt.Println()

	if allHealthy {
		fmt.Println("✅ All backends are healthy")
		return nil
	} else {
		fmt.Println("❌ Some backends failed health check")
		return fmt.Errorf("health check failed")
	}
}
