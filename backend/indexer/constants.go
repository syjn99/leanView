package indexer

import "time"

const (
	// TODO: Make it configurable
	SECONDS_PER_SLOT = 4 * time.Second

	// HTTP client configuration
	defaultHTTPTimeout = 30 * time.Second

	// Health checking configuration
	defaultHealthTimeout = 10 * time.Second

	// Health checking configuration
	defaultHealthCheckInterval = 30 * time.Second

	// Block polling configuration
	defaultPollingInterval = SECONDS_PER_SLOT // Poll every slot (4 seconds)
	defaultRetryDelay      = 2 * time.Second
	defaultMaxRetries      = 3
)
