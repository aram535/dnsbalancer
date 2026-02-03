package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/aram535/dnsbalancer/config"
	"github.com/aram535/dnsbalancer/lb"
	"github.com/aram535/dnsbalancer/logging"
)

var (
	listenAddr string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the DNS load balancer server",
	Long: `Start the DNS load balancer server and begin accepting DNS queries.

The server will distribute queries across configured backends using
round-robin load balancing with optional health checking.

Example:
  dnsbalancer serve
  dnsbalancer serve --config /etc/dnsbalancer/config.yaml
  dnsbalancer serve --debug
  dnsbalancer serve --listen 0.0.0.0:5353`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&listenAddr, "listen", "", "listen address override (e.g., 0.0.0.0:53)")
}

func runServe(cmd *cobra.Command, args []string) error {
	// Find and load config
	configFile := findConfigFile()
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with command-line flags
	if listenAddr != "" {
		cfg.Listen = listenAddr
	}
	if logLevel != "" {
		cfg.LogLevel = logLevel
	}

	// Setup logger
	logger, err := logging.SetupLogger(cfg, debug)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"version":       "1.0.0",
		"config_file":   configFile,
		"listen":        cfg.Listen,
		"backends":      len(cfg.Backends),
		"health_check":  cfg.HealthCheck.Enabled,
		"fail_behavior": cfg.FailBehavior,
	}).Info("Starting dnsbalancer")

	// Create load balancer
	loadBalancer, err := lb.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create load balancer: %w", err)
	}

	// Start the server
	if err := loadBalancer.Start(cfg.Listen); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.WithField("signal", sig.String()).Info("Received shutdown signal")

	// Graceful shutdown
	if err := loadBalancer.Stop(); err != nil {
		logger.WithError(err).Error("Error during shutdown")
		return err
	}

	logger.Info("Shutdown complete")
	return nil
}
