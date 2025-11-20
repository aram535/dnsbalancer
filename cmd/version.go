package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information (set by build flags)
	Version   = "1.0.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display the version, build date, and git commit of dnsbalancer.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("dnsbalancer version %s\n", Version)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Built: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
