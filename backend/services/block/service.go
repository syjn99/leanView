package block

import (
	"context"
	"encoding/hex"
	"errors"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"

	apiv1 "github.com/syjn99/leanView/backend/gen/proto/api/v1"
	"github.com/syjn99/leanView/backend/indexer"
)

// BlockService handles API requests for block data
type BlockService struct {
	indexer *indexer.Indexer
	logger  *logrus.Entry
}

// NewBlockService creates a new Block service instance
func NewBlockService(indexer *indexer.Indexer, logger *logrus.Logger) *BlockService {
	return &BlockService{
		indexer: indexer,
		logger:  logger.WithField("component", "block_service"),
	}
}

// GetLatestBlockHeader returns the current head block header from the head cache
func (s *BlockService) GetLatestBlockHeader(
	ctx context.Context,
	req *connect.Request[apiv1.GetLatestBlockHeaderRequest],
) (*connect.Response[apiv1.GetLatestBlockHeaderResponse], error) {
	// Get head cache from indexer
	headCache := s.indexer.GetHeadCache()
	if headCache == nil {
		s.logger.Error("Head cache is not available")
		return nil, connect.NewError(
			connect.CodeInternal,
			errors.New("head cache not initialized"),
		)
	}

	// Get current head block
	currentHead := headCache.GetCurrentHead()
	if currentHead == nil {
		s.logger.Warn("No head block available in cache")
		return nil, connect.NewError(
			connect.CodeNotFound,
			errors.New("no blocks available yet"),
		)
	}

	// Calculate block root
	blockRoot, err := currentHead.HashTreeRoot()
	if err != nil {
		s.logger.WithError(err).Error("Failed to calculate block root")
		return nil, connect.NewError(
			connect.CodeInternal,
			errors.New("failed to calculate block root"),
		)
	}

	// Convert to protobuf format with 0x prefix
	protoHeader := &apiv1.BlockHeader{
		Slot:          currentHead.Slot,
		ProposerIndex: currentHead.ProposerIndex,
		ParentRoot:    "0x" + hex.EncodeToString(currentHead.ParentRoot),
		StateRoot:     "0x" + hex.EncodeToString(currentHead.StateRoot),
		BodyRoot:      "0x" + hex.EncodeToString(currentHead.BodyRoot),
	}

	blockRootHex := "0x" + hex.EncodeToString(blockRoot[:])

	s.logger.WithFields(logrus.Fields{
		"slot":       currentHead.Slot,
		"block_root": blockRootHex[:8] + "...",
	}).Debug("Serving latest block header")

	return connect.NewResponse(&apiv1.GetLatestBlockHeaderResponse{
		BlockHeader: protoHeader,
		BlockRoot:   blockRootHex,
	}), nil
}
