package lb

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/dnsbalancer/backend"
	"github.com/yourusername/dnsbalancer/config"
)

// LoadBalancer manages DNS query distribution across backends
type LoadBalancer struct {
	backends      []*backend.Backend
	currentIndex  uint32
	timeout       time.Duration
	failBehavior  string // "closed" or "open"
	logger        *logrus.Logger
	healthChecker *HealthChecker
	listener      *net.UDPConn
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// New creates a new LoadBalancer instance
func New(cfg *config.Config, logger *logrus.Logger) (*LoadBalancer, error) {
	// Create backends
	backends := make([]*backend.Backend, len(cfg.Backends))
	for i, bcfg := range cfg.Backends {
		backends[i] = backend.NewBackend(bcfg.Address)
		logger.WithField("backend", bcfg.Address).Info("Registered backend")
	}

	ctx, cancel := context.WithCancel(context.Background())

	lb := &LoadBalancer{
		backends:     backends,
		timeout:      cfg.Timeout,
		failBehavior: cfg.FailBehavior,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Initialize health checker if enabled
	if cfg.HealthCheck.Enabled {
		lb.healthChecker = NewHealthChecker(backends, &cfg.HealthCheck, logger)
		logger.Info("Health checking enabled")
	}

	return lb, nil
}

// Start begins listening for DNS queries
func (lb *LoadBalancer) Start(listenAddr string) error {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve listen address: %w", err)
	}

	lb.listener, err = net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	lb.logger.WithField("address", listenAddr).Info("DNS load balancer started")

	// Start health checker if configured
	if lb.healthChecker != nil {
		lb.healthChecker.Start(lb.ctx)
	}

	// Start accepting queries
	lb.wg.Add(1)
	go lb.acceptQueries()

	return nil
}

// Stop gracefully shuts down the load balancer
func (lb *LoadBalancer) Stop() error {
	lb.logger.Info("Shutting down DNS load balancer")

	// Cancel context to stop health checker and query handlers
	lb.cancel()

	// Close listener
	if lb.listener != nil {
		if err := lb.listener.Close(); err != nil {
			lb.logger.WithError(err).Error("Error closing listener")
		}
	}

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		lb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		lb.logger.Info("Graceful shutdown complete")
	case <-time.After(5 * time.Second):
		lb.logger.Warn("Shutdown timeout reached, forcing exit")
	}

	return nil
}

// acceptQueries listens for incoming DNS queries
func (lb *LoadBalancer) acceptQueries() {
	defer lb.wg.Done()

	buffer := make([]byte, 4096)

	for {
		select {
		case <-lb.ctx.Done():
			return
		default:
		}

		// Set read deadline to allow periodic context checking
		lb.listener.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, clientAddr, err := lb.listener.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Read timeout, check context and try again
			}
			
			// Check if we're shutting down
			select {
			case <-lb.ctx.Done():
				return
			default:
				lb.logger.WithError(err).Error("Error reading from UDP socket")
				continue
			}
		}

		// Copy query data for the goroutine
		query := make([]byte, n)
		copy(query, buffer[:n])

		// Handle query in separate goroutine
		lb.wg.Add(1)
		go lb.handleQuery(query, clientAddr)
	}
}

// handleQuery processes a single DNS query
func (lb *LoadBalancer) handleQuery(query []byte, clientAddr *net.UDPAddr) {
	defer lb.wg.Done()

	logger := lb.logger.WithFields(logrus.Fields{
		"client": clientAddr.String(),
	})

	// Select backend
	backend := lb.selectBackend()
	if backend == nil {
		logger.Error("No healthy backends available")
		
		if lb.failBehavior == "closed" {
			// TODO: Send SERVFAIL response
			logger.Debug("Fail-closed: dropping query")
			return
		}
		// Fail-open: try anyway with first backend
		if len(lb.backends) > 0 {
			backend = lb.backends[0]
			logger.Debug("Fail-open: attempting query with unhealthy backend")
		} else {
			return
		}
	}

	logger = logger.WithField("backend", backend.Address)
	logger.Debug("Forwarding query to backend")

	// Forward query to backend
	response, err := backend.ForwardQuery(query, lb.timeout)
	if err != nil {
		logger.WithError(err).Error("Backend query failed")
		return
	}

	// Send response back to client
	if _, err := lb.listener.WriteToUDP(response, clientAddr); err != nil {
		logger.WithError(err).Error("Failed to send response to client")
		return
	}

	logger.Debug("Query handled successfully")
}

// selectBackend chooses the next healthy backend using round-robin
func (lb *LoadBalancer) selectBackend() *backend.Backend {
	if len(lb.backends) == 0 {
		return nil
	}

	maxAttempts := len(lb.backends)

	for i := 0; i < maxAttempts; i++ {
		idx := atomic.AddUint32(&lb.currentIndex, 1) % uint32(len(lb.backends))
		backend := lb.backends[idx]

		if backend.IsHealthy() {
			return backend
		}
	}

	// All backends unhealthy
	return nil
}

// GetBackends returns the list of backends (for status reporting)
func (lb *LoadBalancer) GetBackends() []*backend.Backend {
	return lb.backends
}
