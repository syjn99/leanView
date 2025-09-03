package indexer

import "time"

const (
	// HTTP client configuration
	defaultHTTPTimeout = 30 * time.Second

	// Health checking configuration
	defaultHealthTimeout = 10 * time.Second

	// Health checking configuration
	defaultHealthCheckInterval = 30 * time.Second
)
