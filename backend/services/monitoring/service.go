package monitoring

import (
	"context"
	"encoding/hex"
	"errors"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"

	apiv1 "github.com/syjn99/leanView/backend/gen/proto/api/v1"
	"github.com/syjn99/leanView/backend/indexer"
)

// MonitoringService handles monitoring API requests for all clients
type MonitoringService struct {
	indexer *indexer.Indexer
	logger  *logrus.Entry
}

// NewMonitoringService creates a new Monitoring service instance
func NewMonitoringService(indexer *indexer.Indexer, logger *logrus.Logger) *MonitoringService {
	return &MonitoringService{
		indexer: indexer,
		logger:  logger.WithField("component", "monitoring_service"),
	}
}

// GetAllClientsHeads returns the current head block from all connected clients
func (s *MonitoringService) GetAllClientsHeads(
	ctx context.Context,
	req *connect.Request[apiv1.GetAllClientsHeadsRequest],
) (*connect.Response[apiv1.GetAllClientsHeadsResponse], error) {
	// Get client pool from indexer
	clientPool := s.indexer.GetClientPool()
	if clientPool == nil {
		s.logger.Error("Client pool is not available")
		return nil, connect.NewError(
			connect.CodeInternal,
			errors.New("client pool not initialized"),
		)
	}

	// Get all clients
	clients := clientPool.GetAllClients()
	clientHeads := make([]*apiv1.ClientHead, 0, len(clients))
	healthyCount := 0

	// Fetch head from each client
	for _, client := range clients {
		config := client.GetConfig()
		isHealthy := client.IsHealthy()
		lastChecked := client.GetLastChecked()

		clientHead := &apiv1.ClientHead{
			ClientLabel:  config.Name,
			EndpointUrl:  config.Url,
			IsHealthy:    isHealthy,
			LastUpdateMs: lastChecked.UnixMilli(),
		}

		if isHealthy {
			healthyCount++

			// Try to fetch the latest block from this client
			block, err := client.GetLatestBlock(ctx)
			if err != nil {
				s.logger.WithError(err).WithField("client", config.Name).Warn("Failed to fetch head from client")
				// Mark as unhealthy if we can't fetch the block
				clientHead.IsHealthy = false
				healthyCount--
			} else {
				// Calculate block root
				blockRoot, err := block.HashTreeRoot()
				if err != nil {
					s.logger.WithError(err).WithField("client", config.Name).Warn("Failed to calculate block root")
				} else {
					clientHead.BlockRoot = "0x" + hex.EncodeToString(blockRoot[:])
				}

				// Convert to protobuf format with 0x prefix
				clientHead.BlockHeader = &apiv1.BlockHeader{
					Slot:          block.Slot,
					ProposerIndex: block.ProposerIndex,
					ParentRoot:    "0x" + hex.EncodeToString(block.ParentRoot),
					StateRoot:     "0x" + hex.EncodeToString(block.StateRoot),
					BodyRoot:      "0x" + hex.EncodeToString(block.BodyRoot),
				}
			}
		}

		clientHeads = append(clientHeads, clientHead)
	}

	s.logger.WithFields(logrus.Fields{
		"total_clients":   len(clients),
		"healthy_clients": healthyCount,
	}).Debug("Serving client heads")

	return connect.NewResponse(&apiv1.GetAllClientsHeadsResponse{
		ClientHeads:    clientHeads,
		TotalClients:   int32(len(clients)),
		HealthyClients: int32(healthyCount),
	}), nil
}
