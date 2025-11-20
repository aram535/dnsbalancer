package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	Listen      string              `yaml:"listen"`
	Timeout     time.Duration       `yaml:"timeout"`
	LogLevel    string              `yaml:"log_level"`
	LogDir      string              `yaml:"log_dir"`
	FailBehavior string             `yaml:"fail_behavior"` // "closed" or "open"
	HealthCheck HealthCheckConfig   `yaml:"health_check"`
	GELF        *GELFConfig         `yaml:"gelf,omitempty"`
	Backends    []BackendConfig     `yaml:"backends"`
}

// BackendConfig represents a single DNS backend server
type BackendConfig struct {
	Address string `yaml:"address"`
	Weight  int    `yaml:"weight,omitempty"` // For future weighted load balancing
}

// HealthCheckConfig represents health check settings
type HealthCheckConfig struct {
	Enabled           bool          `yaml:"enabled"`
	Interval          time.Duration `yaml:"interval"`
	Timeout           time.Duration `yaml:"timeout"`
	FailureThreshold  int           `yaml:"failure_threshold"`
	SuccessThreshold  int           `yaml:"success_threshold"`
	QueryName         string        `yaml:"query_name"`
	QueryType         string        `yaml:"query_type"`
}

// GELFConfig represents GELF logging configuration
type GELFConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Address  string `yaml:"address"`
	Protocol string `yaml:"protocol"` // "tcp" or "udp"
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Listen:       "0.0.0.0:53",
		Timeout:      3 * time.Second,
		LogLevel:     "info",
		LogDir:       "/var/log/dnsbalancer",
		FailBehavior: "closed",
		HealthCheck: HealthCheckConfig{
			Enabled:          false,
			Interval:         10 * time.Second,
			Timeout:          2 * time.Second,
			FailureThreshold: 3,
			SuccessThreshold: 2,
			QueryName:        ".",
			QueryType:        "NS",
		},
		Backends: []BackendConfig{
			{Address: "192.168.1.2:53"},
			{Address: "192.168.1.3:53"},
		},
	}
}

// LoadConfig attempts to load configuration from file
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	// If no file exists, return defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Listen == "" {
		return fmt.Errorf("listen address cannot be empty")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if len(c.Backends) == 0 {
		return fmt.Errorf("at least one backend must be configured")
	}

	for i, backend := range c.Backends {
		if backend.Address == "" {
			return fmt.Errorf("backend %d: address cannot be empty", i)
		}
	}

	if c.FailBehavior != "closed" && c.FailBehavior != "open" {
		return fmt.Errorf("fail_behavior must be either 'closed' or 'open'")
	}

	if c.HealthCheck.Enabled {
		if c.HealthCheck.Interval <= 0 {
			return fmt.Errorf("health check interval must be positive")
		}
		if c.HealthCheck.Timeout <= 0 {
			return fmt.Errorf("health check timeout must be positive")
		}
		if c.HealthCheck.FailureThreshold <= 0 {
			return fmt.Errorf("health check failure threshold must be positive")
		}
		if c.HealthCheck.SuccessThreshold <= 0 {
			return fmt.Errorf("health check success threshold must be positive")
		}
	}

	return nil
}

// SaveExample saves an example configuration file
func SaveExample(path string) error {
	cfg := DefaultConfig()
	cfg.HealthCheck.Enabled = true
	cfg.GELF = &GELFConfig{
		Enabled:  false,
		Address:  "graylog.example.com:12201",
		Protocol: "tcp",
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
