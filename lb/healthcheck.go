package lb

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/aram535/dnsbalancer/backend"
	"github.com/aram535/dnsbalancer/config"
)

// HealthChecker performs periodic health checks on backends
type HealthChecker struct {
	backends         []*backend.Backend
	config           *config.HealthCheckConfig
	logger           *logrus.Logger
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(backends []*backend.Backend, cfg *config.HealthCheckConfig, logger *logrus.Logger) *HealthChecker {
	return &HealthChecker{
		backends: backends,
		config:   cfg,
		logger:   logger,
	}
}

// Start begins periodic health checking
func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(hc.config.Interval)

	go func() {
		// Perform initial health check immediately
		hc.checkAllBackends()

		for {
			select {
			case <-ticker.C:
				hc.checkAllBackends()
			case <-ctx.Done():
				ticker.Stop()
				hc.logger.Info("Health checker stopped")
				return
			}
		}
	}()

	hc.logger.WithFields(logrus.Fields{
		"interval":           hc.config.Interval,
		"timeout":            hc.config.Timeout,
		"failure_threshold":  hc.config.FailureThreshold,
		"success_threshold":  hc.config.SuccessThreshold,
		"query":              hc.config.QueryName,
	}).Info("Health checker started")
}

// checkAllBackends performs health checks on all backends
func (hc *HealthChecker) checkAllBackends() {
	for _, backend := range hc.backends {
		go hc.checkBackend(backend)
	}
}

// checkBackend performs a health check on a single backend
func (hc *HealthChecker) checkBackend(b *backend.Backend) {
	logger := hc.logger.WithField("backend", b.Address)

	err := b.HealthCheck(hc.config.QueryName, hc.config.QueryType, hc.config.Timeout)
	success := err == nil

	if !success {
		logger.WithError(err).Debug("Health check failed")
	}

	// Record the result and check if health status changed
	healthChanged, newHealth := b.RecordHealthCheck(
		success,
		hc.config.FailureThreshold,
		hc.config.SuccessThreshold,
	)

	if healthChanged {
		if newHealth {
			logger.Info("Backend recovered and marked healthy")
		} else {
			logger.Warn("Backend marked unhealthy")
		}
	} else if !success {
		// Log failures even if health hasn't changed yet
		logger.Debug("Health check failed but threshold not reached")
	}
}
