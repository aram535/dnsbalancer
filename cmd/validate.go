package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/dnsbalancer/config"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long: `Validate the syntax and content of the configuration file.

This command checks:
  - YAML syntax is correct
  - All required fields are present
  - Values are within acceptable ranges
  - Backend addresses are properly formatted

Example:
  dnsbalancer validate
  dnsbalancer validate --config /etc/dnsbalancer/config.yaml`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	configFile := findConfigFile()
	
	if configFile == "" {
		return fmt.Errorf("no config file found (searched: ./config.yaml, /etc/dnsbalancer/config.yaml)")
	}

	fmt.Printf("Validating config file: %s\n", configFile)

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("❌ Configuration is INVALID\n")
		return err
	}

	// Print summary
	fmt.Printf("✅ Configuration is VALID\n\n")
	fmt.Printf("Summary:\n")
	fmt.Printf("  Listen Address:    %s\n", cfg.Listen)
	fmt.Printf("  Timeout:           %s\n", cfg.Timeout)
	fmt.Printf("  Log Level:         %s\n", cfg.LogLevel)
	fmt.Printf("  Log Directory:     %s\n", cfg.LogDir)
	fmt.Printf("  Fail Behavior:     %s\n", cfg.FailBehavior)
	fmt.Printf("  Backends:          %d\n", len(cfg.Backends))
	
	for i, backend := range cfg.Backends {
		fmt.Printf("    %d. %s\n", i+1, backend.Address)
	}

	fmt.Printf("\n  Health Check:\n")
	if cfg.HealthCheck.Enabled {
		fmt.Printf("    Enabled:         yes\n")
		fmt.Printf("    Interval:        %s\n", cfg.HealthCheck.Interval)
		fmt.Printf("    Timeout:         %s\n", cfg.HealthCheck.Timeout)
		fmt.Printf("    Fail Threshold:  %d\n", cfg.HealthCheck.FailureThreshold)
		fmt.Printf("    Success Threshold: %d\n", cfg.HealthCheck.SuccessThreshold)
		fmt.Printf("    Query:           %s (%s)\n", cfg.HealthCheck.QueryName, cfg.HealthCheck.QueryType)
	} else {
		fmt.Printf("    Enabled:         no\n")
	}

	if cfg.GELF != nil && cfg.GELF.Enabled {
		fmt.Printf("\n  GELF Logging:\n")
		fmt.Printf("    Enabled:         yes\n")
		fmt.Printf("    Address:         %s\n", cfg.GELF.Address)
		fmt.Printf("    Protocol:        %s\n", cfg.GELF.Protocol)
	}

	return nil
}
