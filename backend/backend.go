package backend

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

// Backend represents a DNS backend server
type Backend struct {
	Address            string
	Healthy            bool
	ConsecutiveFails   int
	ConsecutiveSuccess int
	LastCheck          time.Time
	LastFail           time.Time
	TotalQueries       uint64
	TotalFailures      uint64
	mu                 sync.RWMutex
}

// NewBackend creates a new backend instance
func NewBackend(address string) *Backend {
	return &Backend{
		Address: address,
		Healthy: true, // Start optimistic
	}
}

// IsHealthy returns the current health status
func (b *Backend) IsHealthy() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Healthy
}

// MarkQueryAttempt increments query counter
func (b *Backend) MarkQueryAttempt() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.TotalQueries++
}

// MarkFailure records a query failure
func (b *Backend) MarkFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.TotalFailures++
	b.LastFail = time.Now()
}

// UpdateHealth updates the health status and logs changes
func (b *Backend) UpdateHealth(healthy bool, logger *logrus.Logger) {
	b.mu.Lock()
	defer b.mu.Unlock()

	oldHealth := b.Healthy
	b.Healthy = healthy

	if oldHealth != healthy {
		if healthy {
			logger.WithFields(logrus.Fields{
				"backend":             b.Address,
				"consecutive_success": b.ConsecutiveSuccess,
			}).Info("Backend recovered and marked healthy")
		} else {
			logger.WithFields(logrus.Fields{
				"backend":            b.Address,
				"consecutive_fails":  b.ConsecutiveFails,
				"last_fail":          b.LastFail,
			}).Warn("Backend marked unhealthy")
		}
	}
}

// RecordHealthCheck records the result of a health check
func (b *Backend) RecordHealthCheck(success bool, failThreshold, successThreshold int) (healthChanged bool, newHealth bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.LastCheck = time.Now()

	if success {
		b.ConsecutiveSuccess++
		b.ConsecutiveFails = 0

		if !b.Healthy && b.ConsecutiveSuccess >= successThreshold {
			b.Healthy = true
			healthChanged = true
			newHealth = true
		}
	} else {
		b.ConsecutiveFails++
		b.ConsecutiveSuccess = 0
		b.LastFail = time.Now()

		if b.Healthy && b.ConsecutiveFails >= failThreshold {
			b.Healthy = false
			healthChanged = true
			newHealth = false
		}
	}

	return healthChanged, b.Healthy
}

// Stats returns current backend statistics
func (b *Backend) Stats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return map[string]interface{}{
		"address":             b.Address,
		"healthy":             b.Healthy,
		"total_queries":       b.TotalQueries,
		"total_failures":      b.TotalFailures,
		"consecutive_fails":   b.ConsecutiveFails,
		"consecutive_success": b.ConsecutiveSuccess,
		"last_check":          b.LastCheck,
		"last_fail":           b.LastFail,
	}
}

// ForwardQuery forwards a DNS query to this backend
func (b *Backend) ForwardQuery(query []byte, timeout time.Duration) ([]byte, error) {
	b.MarkQueryAttempt()

	conn, err := net.DialTimeout("udp", b.Address, timeout)
	if err != nil {
		b.MarkFailure()
		return nil, fmt.Errorf("failed to connect to backend: %w", err)
	}
	defer conn.Close()

	// Set deadline for the entire operation
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		b.MarkFailure()
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Send query
	if _, err := conn.Write(query); err != nil {
		b.MarkFailure()
		return nil, fmt.Errorf("failed to send query: %w", err)
	}

	// Read response (DNS messages are typically < 512 bytes for UDP)
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		b.MarkFailure()
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return buffer[:n], nil
}

// HealthCheck performs a DNS health check query
func (b *Backend) HealthCheck(queryName, queryType string, timeout time.Duration) error {
	// Create DNS query message
	m := new(dns.Msg)
	
	var qtype uint16
	switch queryType {
	case "A":
		qtype = dns.TypeA
	case "AAAA":
		qtype = dns.TypeAAAA
	case "NS":
		qtype = dns.TypeNS
	case "ANY":
		qtype = dns.TypeANY
	default:
		qtype = dns.TypeNS
	}

	m.SetQuestion(dns.Fqdn(queryName), qtype)
	m.RecursionDesired = true

	// Pack the message
	query, err := m.Pack()
	if err != nil {
		return fmt.Errorf("failed to pack DNS query: %w", err)
	}

	// Send to backend
	conn, err := net.DialTimeout("udp", b.Address, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	if _, err := conn.Write(query); err != nil {
		return fmt.Errorf("failed to send query: %w", err)
	}

	// Read response
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Verify it's a valid DNS response
	response := new(dns.Msg)
	if err := response.Unpack(buffer[:n]); err != nil {
		return fmt.Errorf("invalid DNS response: %w", err)
	}

	// Check if response has error
	if response.Rcode != dns.RcodeSuccess && response.Rcode != dns.RcodeNameError {
		return fmt.Errorf("DNS error response: %s", dns.RcodeToString[response.Rcode])
	}

	return nil
}
