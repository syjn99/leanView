package block

import (
	"context"
	"encoding/hex"
	"errors"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"

	"github.com/syjn99/leanView/backend/db"
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

// GetBlockHeaders returns paginated block headers from the database
func (s *BlockService) GetBlockHeaders(
	ctx context.Context,
	req *connect.Request[apiv1.GetBlockHeadersRequest],
) (*connect.Response[apiv1.GetBlockHeadersResponse], error) {
	// Validate and set default values for request parameters
	limit := req.Msg.Limit
	if limit == 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}
	
	offset := req.Msg.Offset
	ascending := req.Msg.SortOrder == apiv1.GetBlockHeadersRequest_SLOT_ASC
	
	// Query database for paginated headers
	headers, err := db.GetBlockHeadersPaginated(int(limit), offset, ascending)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch paginated block headers")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	
	// Get total count
	totalCount, err := db.GetTotalBlockCount()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get total block count")
		totalCount = uint32(len(headers))
	}
	
	// Convert to protobuf format
	var protoHeaders []*apiv1.BlockHeaderWithRoot
	for _, header := range headers {
		// Calculate block root
		blockRoot, err := header.HashTreeRoot()
		if err != nil {
			s.logger.WithError(err).WithField("slot", header.Slot).Warn("Failed to calculate block root")
			continue
		}
		
		protoHeader := &apiv1.BlockHeaderWithRoot{
			Header: &apiv1.BlockHeader{
				Slot:          header.Slot,
				ProposerIndex: header.ProposerIndex,
				ParentRoot:    "0x" + hex.EncodeToString(header.ParentRoot),
				StateRoot:     "0x" + hex.EncodeToString(header.StateRoot),
				BodyRoot:      "0x" + hex.EncodeToString(header.BodyRoot),
			},
			BlockRoot: "0x" + hex.EncodeToString(blockRoot[:]),
		}
		protoHeaders = append(protoHeaders, protoHeader)
	}
	
	// Determine if there are more results
	hasMore := len(headers) == int(limit)
	
	// Calculate next offset
	var nextOffset uint64
	if len(headers) > 0 {
		if ascending {
			nextOffset = headers[len(headers)-1].Slot + 1
		} else {
			nextOffset = headers[len(headers)-1].Slot - 1
		}
	} else {
		nextOffset = offset
	}
	
	s.logger.WithFields(logrus.Fields{
		"limit":      limit,
		"offset":     offset,
		"count":      len(protoHeaders),
		"total":      totalCount,
		"ascending":  ascending,
	}).Debug("Serving paginated block headers")
	
	return connect.NewResponse(&apiv1.GetBlockHeadersResponse{
		Headers:     protoHeaders,
		TotalCount:  totalCount,
		HasMore:     hasMore,
		NextOffset:  nextOffset,
	}), nil
}
