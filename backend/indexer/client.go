package indexer

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/syjn99/leanView/backend/types"
)

// Client represents a connection to a PQ Devnet endpoint
type Client struct {
	config     *types.EndpointConfig
	httpClient *HTTPClient

	// Connection state
	isHealthy   bool
	lastError   error
	lastChecked time.Time

	// Synchronization
	mutex sync.RWMutex

	logger logrus.FieldLogger
}

// NewClient creates a new client for a PQ Devnet endpoint
func NewClient(config *types.EndpointConfig, logger logrus.FieldLogger) *Client {
	return &Client{
		config:      config,
		httpClient:  NewHTTPClient(config.Url, defaultHTTPTimeout),
		isHealthy:   true, // Start optimistically
		lastChecked: time.Now(),
		logger:      logger.WithField("endpoint", config.Name),
	}
}

// HealthCheck performs a health check using the /lean/v0/headers/head endpoint
func (c *Client) HealthCheck(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create a context with health check timeout
	healthCtx, cancel := context.WithTimeout(ctx, defaultHealthTimeout)
	defer cancel()

	// Try to fetch the head block to verify connectivity
	_, err := c.httpClient.GetHeadBlock(healthCtx)
	c.lastChecked = time.Now()

	if err != nil {
		c.isHealthy = false
		c.lastError = err
		c.logger.WithError(err).Warn("Health check failed")
		return err
	}

	c.isHealthy = true
	c.lastError = nil
	c.logger.Debug("Health check passed")
	return nil
}

// GetLatestBlock fetches the current head block
func (c *Client) GetLatestBlock(ctx context.Context) (*types.BlockHeader, error) {
	return c.httpClient.GetHeadBlock(ctx)
}

// GetBlockBySlot fetches a block by slot number
func (c *Client) GetBlockBySlot(ctx context.Context, slot uint64) (*types.BlockHeader, error) {
	return c.httpClient.GetBlockBySlot(ctx, slot)
}

// GetBlockByRoot fetches a block by its root hash
func (c *Client) GetBlockByRoot(ctx context.Context, root []byte) (*types.BlockHeader, error) {
	return c.httpClient.GetBlockByRoot(ctx, root)
}

// GetFinalizedBlock fetches the finalized block
func (c *Client) GetFinalizedBlock(ctx context.Context) (*types.BlockHeader, error) {
	return c.httpClient.GetFinalizedBlock(ctx)
}

// GetJustifiedBlock fetches the justified block
func (c *Client) GetJustifiedBlock(ctx context.Context) (*types.BlockHeader, error) {
	return c.httpClient.GetJustifiedBlock(ctx)
}

// GetGenesisBlock fetches the genesis block
func (c *Client) GetGenesisBlock(ctx context.Context) (*types.BlockHeader, error) {
	return c.httpClient.GetGenesisBlock(ctx)
}

// GetBlockRange fetches a range of blocks by slot numbers
func (c *Client) GetBlockRange(ctx context.Context, start, end uint64) ([]*types.BlockHeader, error) {
	return c.httpClient.GetBlockRange(ctx, start, end)
}

// IsHealthy returns the current health status of the client
func (c *Client) IsHealthy() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isHealthy
}

// GetLastError returns the last error encountered by this client
func (c *Client) GetLastError() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.lastError
}

// GetConfig returns the client's endpoint configuration
func (c *Client) GetConfig() *types.EndpointConfig {
	// Config is immutable, no lock needed
	return c.config
}

// GetLastChecked returns the last health check timestamp
func (c *Client) GetLastChecked() time.Time {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.lastChecked
}
