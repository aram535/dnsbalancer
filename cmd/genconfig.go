package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/aram535/dnsbalancer/config"
)

var (
	outputFile string
)

// genconfigCmd represents the genconfig command
var genconfigCmd = &cobra.Command{
	Use:   "genconfig",
	Short: "Generate an example configuration file",
	Long: `Generate an example configuration file with all available options.

This creates a fully documented YAML configuration file that you can
customize for your environment.

Example:
  dnsbalancer genconfig
  dnsbalancer genconfig --output /etc/dnsbalancer/config.yaml
  dnsbalancer genconfig --output ./my-config.yaml`,
	RunE: runGenconfig,
}

func init() {
	rootCmd.AddCommand(genconfigCmd)

	genconfigCmd.Flags().StringVarP(&outputFile, "output", "o", "config.yaml", "output file path")
}

func runGenconfig(cmd *cobra.Command, args []string) error {
	// Check if file exists
	if _, err := os.Stat(outputFile); err == nil {
		fmt.Printf("File already exists: %s\n", outputFile)
		fmt.Print("Overwrite? (y/N): ")
		
		var response string
		fmt.Scanln(&response)
		
		if response != "y" && response != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Generate example config
	if err := config.SaveExample(outputFile); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	fmt.Printf("✅ Example configuration written to: %s\n", outputFile)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit the configuration file to match your environment")
	fmt.Println("  2. Validate it: dnsbalancer validate --config " + outputFile)
	fmt.Println("  3. Start the server: dnsbalancer serve --config " + outputFile)

	return nil
}
