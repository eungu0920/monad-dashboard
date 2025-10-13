package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

type MonadClient struct {
	BFTRPCUrl      string
	ExecutionRPCUrl string
	BFTIPCPath     string
	ExecutionIPCPath string
	httpClient     *http.Client
}

func NewMonadClient(monadRPC, bftIPC, execIPC string) *MonadClient {
	return &MonadClient{
		BFTRPCUrl:        monadRPC, // Use same RPC server for BFT metrics
		ExecutionRPCUrl:  monadRPC, // This is actually monad-rpc server
		BFTIPCPath:       bftIPC,
		ExecutionIPCPath: execIPC,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// BFT Consensus metrics via RPC
func (c *MonadClient) GetConsensusMetrics() (*ConsensusMetrics, error) {
	// Try RPC first
	if c.BFTRPCUrl != "" {
		metrics, err := c.getConsensusViaRPC()
		if err != nil {
			log.Printf("RPC connection failed: %v", err)
			// Don't try IPC if RPC fails, return error to trigger fallback to mock
			return nil, err
		}
		return metrics, nil
	}

	// Fallback to IPC (only if RPC URL not configured)
	if c.BFTIPCPath != "" {
		return c.getConsensusViaIPC()
	}

	return nil, fmt.Errorf("no BFT connection method configured")
}

func (c *MonadClient) getConsensusViaRPC() (*ConsensusMetrics, error) {
	// Get latest block number
	blockNumResp, err := c.rpcCall(c.BFTRPCUrl, "eth_blockNumber", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get block number: %w", err)
	}

	var blockNumResult struct {
		Result string `json:"result"`
	}

	if err := json.Unmarshal(blockNumResp, &blockNumResult); err != nil {
		return nil, fmt.Errorf("failed to decode block number: %w", err)
	}

	// Get latest block
	blockResp, err := c.rpcCall(c.BFTRPCUrl, "eth_getBlockByNumber", []interface{}{"latest", false})
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	var block struct {
		Result struct {
			Number    string `json:"number"`
			Timestamp string `json:"timestamp"`
			Hash      string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(blockResp, &block); err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}

	// Parse block height and timestamp
	height, _ := parseHexToInt64(block.Result.Number)
	timestamp, _ := parseHexToInt64(block.Result.Timestamp)

	return &ConsensusMetrics{
		CurrentHeight:     height,
		LastBlockTime:     timestamp,
		BlockTime:         0.4,  // Monad block time
		ValidatorCount:    100,  // Default - would need custom endpoint
		VotingPower:       1000000, // Default
		ParticipationRate: 0.9,  // Default
	}, nil
}

func (c *MonadClient) getConsensusViaIPC() (*ConsensusMetrics, error) {
	// Connect to Unix socket
	conn, err := net.Dial("unix", c.BFTIPCPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to BFT IPC: %w", err)
	}
	defer conn.Close()

	// Send request for consensus metrics
	request := map[string]interface{}{
		"method": "consensus_metrics",
	}

	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return nil, fmt.Errorf("failed to send IPC request: %w", err)
	}

	var response ConsensusMetrics
	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode IPC response: %w", err)
	}

	return &response, nil
}

// Execution metrics via RPC
func (c *MonadClient) GetExecutionMetrics() (*ExecutionMetrics, error) {
	if c.ExecutionRPCUrl == "" && c.ExecutionIPCPath == "" {
		return nil, fmt.Errorf("no execution connection method configured")
	}

	// Try RPC first
	if c.ExecutionRPCUrl != "" {
		return c.getExecutionViaRPC()
	}

	// Fallback to IPC
	return c.getExecutionViaIPC()
}

func (c *MonadClient) getExecutionViaRPC() (*ExecutionMetrics, error) {
	// Get latest block
	blockResp, err := c.rpcCall(c.ExecutionRPCUrl, "eth_getBlockByNumber", []interface{}{"latest", false})
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	var block struct {
		Result struct {
			Number       string   `json:"number"`
			Transactions []string `json:"transactions"`
			GasUsed      string   `json:"gasUsed"`
		} `json:"result"`
	}

	if err := json.Unmarshal(blockResp, &block); err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}

	// Get pending transactions
	pendingResp, err := c.rpcCall(c.ExecutionRPCUrl, "eth_pendingTransactions", []interface{}{})
	if err != nil {
		log.Printf("Failed to get pending transactions: %v", err)
	}

	var pending struct {
		Result []interface{} `json:"result"`
	}

	pendingCount := int64(0)
	if pendingResp != nil {
		if json.Unmarshal(pendingResp, &pending) == nil {
			pendingCount = int64(len(pending.Result))
		}
	}

	// Calculate TPS (rough estimation)
	tps := float64(len(block.Result.Transactions)) / 0.4 // Monad 0.4s block time

	gasUsed, _ := parseHexToInt64(block.Result.GasUsed)
	_ = gasUsed // Use the variable to avoid unused error

	return &ExecutionMetrics{
		TPS:                 tps,
		PendingTxCount:      pendingCount,
		ParallelSuccessRate: 0.85, // Default - would need custom metrics endpoint
		AvgGasPrice:         21,   // Default gwei
		AvgExecutionTime:    5.0,  // Default ms
		StateSize:           1000000000, // Default bytes
	}, nil
}

func (c *MonadClient) getExecutionViaIPC() (*ExecutionMetrics, error) {
	conn, err := net.Dial("unix", c.ExecutionIPCPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to execution IPC: %w", err)
	}
	defer conn.Close()

	request := map[string]interface{}{
		"method": "execution_metrics",
	}

	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return nil, fmt.Errorf("failed to send IPC request: %w", err)
	}

	var response ExecutionMetrics
	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode IPC response: %w", err)
	}

	return &response, nil
}

// Network metrics (can be gathered from both BFT and Execution)
func (c *MonadClient) GetNetworkMetrics() (*NetworkMetrics, error) {
	// For now, return default network metrics as Monad doesn't expose standard network endpoints
	// In a real implementation, these would come from custom Monad metrics endpoints
	return &NetworkMetrics{
		PeerCount:      50 + rand.Intn(20),
		InboundPeers:   25 + rand.Intn(10),
		OutboundPeers:  25 + rand.Intn(10),
		BytesIn:        int64(rand.Intn(1000000)),
		BytesOut:       int64(rand.Intn(1000000)),
		NetworkLatency: 50.0 + rand.Float64()*50.0,
	}, nil
}

// Helper functions
func (c *MonadClient) rpcCall(url, method string, params []interface{}) ([]byte, error) {
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(url, "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func parseStringToInt64(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func parseHexToInt64(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "0x%x", &result)
	return result, err
}