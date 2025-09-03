package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/syjn99/leanView/backend/types"
)

// HTTPClient handles communication with PQ Devnet API
type HTTPClient struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
}

// NewHTTPClient creates a new HTTP client for API communication
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		timeout: timeout,
	}
}

// GetHeadBlock fetches the current head block
func (hc *HTTPClient) GetHeadBlock(ctx context.Context) (*types.BlockHeader, error) {
	return hc.fetchBlockHeader(ctx, "head")
}

// GetFinalizedBlock fetches the finalized block
func (hc *HTTPClient) GetFinalizedBlock(ctx context.Context) (*types.BlockHeader, error) {
	return hc.fetchBlockHeader(ctx, "finalized")
}

// GetJustifiedBlock fetches the justified block
func (hc *HTTPClient) GetJustifiedBlock(ctx context.Context) (*types.BlockHeader, error) {
	return hc.fetchBlockHeader(ctx, "justified")
}

// GetGenesisBlock fetches the genesis block
func (hc *HTTPClient) GetGenesisBlock(ctx context.Context) (*types.BlockHeader, error) {
	return hc.fetchBlockHeader(ctx, "genesis")
}

// GetBlockBySlot fetches a block by slot number
func (hc *HTTPClient) GetBlockBySlot(ctx context.Context, slot uint64) (*types.BlockHeader, error) {
	return hc.fetchBlockHeader(ctx, fmt.Sprintf("%d", slot))
}

// GetBlockByRoot fetches a block by its root hash
func (hc *HTTPClient) GetBlockByRoot(ctx context.Context, root []byte) (*types.BlockHeader, error) {
	rootHex := fmt.Sprintf("0x%x", root)
	return hc.fetchBlockHeader(ctx, rootHex)
}

// GetBlockRange fetches a range of blocks by slot numbers
func (hc *HTTPClient) GetBlockRange(ctx context.Context, start, end uint64) ([]*types.BlockHeader, error) {
	var blocks []*types.BlockHeader

	for slot := start; slot <= end; slot++ {
		block, err := hc.GetBlockBySlot(ctx, slot)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch block at slot %d: %w", slot, err)
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

// fetchBlockHeader is the internal method that handles the actual HTTP request
func (hc *HTTPClient) fetchBlockHeader(ctx context.Context, blockId string) (*types.BlockHeader, error) {
	url := hc.buildEndpointURL(blockId)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block header: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d for block_id %s", resp.StatusCode, blockId)
	}

	return hc.parseBlockHeaderResponse(resp)
}

// buildEndpointURL constructs the full URL for the API request
// TODO: Make baseURL configurable (not only `headers`)
func (hc *HTTPClient) buildEndpointURL(blockId string) string {
	return fmt.Sprintf("%s/lean/v0/headers/%s", hc.baseURL, blockId)
}

// parseBlockHeaderResponse parses the JSON response into a BlockHeader
func (hc *HTTPClient) parseBlockHeaderResponse(resp *http.Response) (*types.BlockHeader, error) {
	var blockHeader types.BlockHeader
	if err := json.NewDecoder(resp.Body).Decode(&blockHeader); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &blockHeader, nil
}
