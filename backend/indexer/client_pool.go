package indexer

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/syjn99/leanView/backend/types"
)

// ClientPool manages multiple endpoint connections
type ClientPool struct {
	clients []*Client
	primary *Client
	logger  logrus.FieldLogger

	// Health check management
	healthCheckTicker *time.Ticker
	stopHealthCheck   chan bool
	mutex             sync.RWMutex
}

// NewClientPool creates a new client pool with multiple endpoints
func NewClientPool(endpoints []types.EndpointConfig, logger logrus.FieldLogger) *ClientPool {
	clients := make([]*Client, len(endpoints))
	for i, endpoint := range endpoints {
		clients[i] = NewClient(&endpoint, logger)
	}

	var primary *Client
	if len(clients) > 0 {
		primary = clients[0]
	}

	return &ClientPool{
		clients:         clients,
		primary:         primary,
		logger:          logger.WithField("component", "client_pool"),
		stopHealthCheck: make(chan bool, 1),
	}
}

// GetHealthyClient returns a healthy client from the pool, or nil if none available
func (cp *ClientPool) GetHealthyClient() *Client {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	// First try the primary client
	if cp.primary != nil && cp.primary.IsHealthy() {
		return cp.primary
	}

	// Fall back to any healthy client
	for _, client := range cp.clients {
		if client.IsHealthy() {
			return client
		}
	}

	cp.logger.Warn("No healthy clients available")
	return nil
}

// GetPrimaryClient returns the primary client regardless of health status
func (cp *ClientPool) GetPrimaryClient() *Client {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()
	return cp.primary
}

// RunHealthChecks starts background health checking for all clients
func (cp *ClientPool) RunHealthChecks(ctx context.Context) {
	cp.healthCheckTicker = time.NewTicker(defaultHealthCheckInterval)

	go func() {
		for {
			select {
			case <-cp.healthCheckTicker.C:
				cp.performHealthChecks(ctx)
			case <-cp.stopHealthCheck:
				cp.healthCheckTicker.Stop()
				return
			case <-ctx.Done():
				cp.healthCheckTicker.Stop()
				return
			}
		}
	}()

	cp.logger.Info("Health checking started")
}

// StopHealthChecks stops the background health checking
func (cp *ClientPool) StopHealthChecks() {
	if cp.healthCheckTicker != nil {
		cp.stopHealthCheck <- true
	}
	cp.logger.Info("Health checking stopped")
}

// performHealthChecks runs health checks on all clients
func (cp *ClientPool) performHealthChecks(ctx context.Context) {
	for _, client := range cp.clients {
		go func(c *Client) {
			if err := c.HealthCheck(ctx); err != nil {
				cp.logger.WithError(err).WithField("endpoint", c.config.Name).Warn("Client health check failed")
			}
		}(client)
	}
}

// GetClientCount returns the total number of clients in the pool
func (cp *ClientPool) GetClientCount() int {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()
	return len(cp.clients)
}

// GetHealthyClientCount returns the number of healthy clients
func (cp *ClientPool) GetHealthyClientCount() int {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	count := 0
	for _, client := range cp.clients {
		if client.IsHealthy() {
			count++
		}
	}
	return count
}
