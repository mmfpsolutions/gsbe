package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mmfpsolutions/gsbe/internal/config"
	"github.com/mmfpsolutions/gsbe/internal/logger"
	v1types "github.com/mmfpsolutions/gsbe/internal/types/v1"
)

// NodeService handles communication with blockchain node REST APIs
type NodeService struct {
	cfgManager *config.Manager
	httpClient *http.Client
	log        *logger.Logger
}

// NewNodeService creates a new NodeService
func NewNodeService(cfgManager *config.Manager) *NodeService {
	return &NodeService{
		cfgManager: cfgManager,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		log:        logger.New(logger.ModuleService),
	}
}

func (s *NodeService) getNode(nodeID string) (*config.NodeConnection, error) {
	node := s.cfgManager.GetNodeByID(nodeID)
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}
	return node, nil
}

func (s *NodeService) restURL(node *config.NodeConnection, path string) string {
	return fmt.Sprintf("http://%s:%d/rest/%s", node.Host, node.Port, path)
}

func (s *NodeService) restGet(node *config.NodeConnection, path string) ([]byte, error) {
	url := s.restURL(node, path)
	s.log.Debug("GET %s", url)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("node returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetChainInfo fetches blockchain info from the node
func (s *NodeService) GetChainInfo(nodeID string) (*v1types.ChainInfo, error) {
	node, err := s.getNode(nodeID)
	if err != nil {
		return nil, err
	}

	body, err := s.restGet(node, "chaininfo.json")
	if err != nil {
		return nil, err
	}

	var info v1types.ChainInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse chaininfo: %w", err)
	}

	return &info, nil
}

// GetBlock fetches a block by hash
func (s *NodeService) GetBlock(nodeID, hash string) (*v1types.Block, error) {
	node, err := s.getNode(nodeID)
	if err != nil {
		return nil, err
	}

	body, err := s.restGet(node, fmt.Sprintf("block/%s.json", hash))
	if err != nil {
		return nil, err
	}

	var block v1types.Block
	if err := json.Unmarshal(body, &block); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}

	return &block, nil
}

// BlockHashResponse is the response from the blockhashbyheight endpoint
type BlockHashResponse struct {
	BlockHash string `json:"blockhash"`
}

// GetBlockHashByHeight fetches the block hash for a given height
func (s *NodeService) GetBlockHashByHeight(nodeID string, height int64) (string, error) {
	node, err := s.getNode(nodeID)
	if err != nil {
		return "", err
	}

	body, err := s.restGet(node, fmt.Sprintf("blockhashbyheight/%d.json", height))
	if err != nil {
		return "", err
	}

	var resp BlockHashResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to parse blockhash response: %w", err)
	}

	return resp.BlockHash, nil
}

// GetMempoolInfo fetches mempool info from the node
func (s *NodeService) GetMempoolInfo(nodeID string) (v1types.MempoolInfo, error) {
	node, err := s.getNode(nodeID)
	if err != nil {
		return nil, err
	}

	body, err := s.restGet(node, "mempool/info.json")
	if err != nil {
		return nil, err
	}

	var info v1types.MempoolInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse mempool info: %w", err)
	}

	return info, nil
}

// BlockSummary is a simplified block for list views
type BlockSummary struct {
	Hash       string      `json:"hash"`
	Height     int64       `json:"height"`
	Time       int64       `json:"time"`
	NTx        int         `json:"nTx"`
	Size       int         `json:"size"`
	Weight     int         `json:"weight"`
	Difficulty interface{} `json:"difficulty"`
}

// GetRecentBlocks fetches the most recent N blocks
func (s *NodeService) GetRecentBlocks(nodeID string, count int) ([]BlockSummary, error) {
	chainInfo, err := s.GetChainInfo(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain info: %w", err)
	}

	currentHeight := chainInfo.Blocks
	blocks := make([]BlockSummary, 0, count)

	// Walk backwards from the tip
	hash := chainInfo.BestBlockHash
	for i := 0; i < count && hash != ""; i++ {
		block, err := s.GetBlock(nodeID, hash)
		if err != nil {
			s.log.Warn("Failed to fetch block %s: %v", hash, err)
			break
		}

		blocks = append(blocks, BlockSummary{
			Hash:       block.Hash,
			Height:     block.Height,
			Time:       block.Time,
			NTx:        block.NTx,
			Size:       block.Size,
			Weight:     block.Weight,
			Difficulty: block.Difficulty,
		})

		hash = block.PreviousBlockHash
		_ = currentHeight // used implicitly via chainInfo
	}

	return blocks, nil
}

// TestConnection tests whether a node's REST API is reachable
func (s *NodeService) TestConnection(node *config.NodeConnection) error {
	url := fmt.Sprintf("http://%s:%d/rest/chaininfo.json", node.Host, node.Port)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("node returned status %d", resp.StatusCode)
	}

	return nil
}

// GetNodeStatuses returns the online status of all configured nodes
func (s *NodeService) GetNodeStatuses() []v1types.NodeStatus {
	cfg := s.cfgManager.GetConfig()
	statuses := make([]v1types.NodeStatus, 0, len(cfg.Nodes))

	for _, node := range cfg.Nodes {
		status := v1types.NodeStatus{
			ID:      node.ID,
			Name:    node.Name,
			Symbol:  node.Symbol,
			Network: node.Network,
		}

		chainInfo, err := s.GetChainInfo(node.ID)
		if err != nil {
			status.Online = false
			status.Message = err.Error()
		} else {
			status.Online = true
			status.ChainHeight = chainInfo.Blocks
		}

		statuses = append(statuses, status)
	}

	return statuses
}
