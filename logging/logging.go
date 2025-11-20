package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/dnsbalancer/config"
)

// SetupLogger initializes and configures the application logger
func SetupLogger(cfg *config.Config, debug bool) (*logrus.Logger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	
	if debug {
		level = logrus.DebugLevel
	}
	
	logger.SetLevel(level)

	// Set formatter
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Configure output
	if debug {
		// Debug mode: log to console
		logger.SetOutput(os.Stdout)
		logger.Info("Logging to console (debug mode)")
	} else {
		// Normal mode: log to file
		if err := setupFileLogging(logger, cfg.LogDir); err != nil {
			return nil, err
		}
	}

	// Setup GELF logging if enabled
	if cfg.GELF != nil && cfg.GELF.Enabled {
		if err := setupGELFLogging(logger, cfg.GELF); err != nil {
			logger.WithError(err).Warn("Failed to setup GELF logging, continuing without it")
		} else {
			logger.WithField("address", cfg.GELF.Address).Info("GELF logging enabled")
		}
	}

	return logger, nil
}

// setupFileLogging configures file-based logging
func setupFileLogging(logger *logrus.Logger, logDir string) error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logFile := filepath.Join(logDir, "dnsbalancer.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logger.SetOutput(file)
	logger.WithField("file", logFile).Info("Logging to file")

	return nil
}

// setupGELFLogging configures GELF output (placeholder for v1.0)
// TODO: Implement actual GELF support with graylog/gelf-go or similar
func setupGELFLogging(logger *logrus.Logger, cfg *config.GELFConfig) error {
	// Placeholder for GELF implementation
	// This would use a library like:
	// - github.com/gemnasium/logrus-graylog-hook
	// - Or custom TCP/UDP GELF writer
	
	logger.WithFields(logrus.Fields{
		"address":  cfg.Address,
		"protocol": cfg.Protocol,
	}).Warn("GELF logging requested but not yet implemented in v1.0")
	
	return fmt.Errorf("GELF support is planned for future release")
}

// RotateLog provides a simple log rotation mechanism
// This is a placeholder - in production you'd use logrotate or similar
func RotateLog(logDir string) error {
	// TODO: Implement log rotation
	// For now, rely on external tools like logrotate
	return nil
}
