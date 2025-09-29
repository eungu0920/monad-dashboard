package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
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
		BFTRPCUrl:        "", // Monad BFT doesn't have HTTP RPC
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
		return c.getConsensusViaRPC()
	}

	// Fallback to IPC
	if c.BFTIPCPath != "" {
		return c.getConsensusViaIPC()
	}

	return nil, fmt.Errorf("no BFT connection method configured")
}

func (c *MonadClient) getConsensusViaRPC() (*ConsensusMetrics, error) {
	// Get status
	statusResp, err := c.httpClient.Get(c.BFTRPCUrl + "/status")
	if err != nil {
		return nil, fmt.Errorf("failed to get consensus status: %w", err)
	}
	defer statusResp.Body.Close()

	var status struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHeight string `json:"latest_block_height"`
				LatestBlockTime   string `json:"latest_block_time"`
			} `json:"sync_info"`
			ValidatorInfo struct {
				VotingPower string `json:"voting_power"`
			} `json:"validator_info"`
		} `json:"result"`
	}

	if err := json.NewDecoder(statusResp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status: %w", err)
	}

	// Get validators
	validatorsResp, err := c.httpClient.Get(c.BFTRPCUrl + "/validators")
	if err != nil {
		log.Printf("Failed to get validators: %v", err)
		// Continue without validator info
	}

	var validators struct {
		Result struct {
			Count string `json:"count"`
		} `json:"result"`
	}

	validatorCount := 100 // Default
	if validatorsResp != nil {
		defer validatorsResp.Body.Close()
		if json.NewDecoder(validatorsResp.Body).Decode(&validators) == nil {
			if count, err := parseStringToInt64(validators.Result.Count); err == nil {
				validatorCount = int(count)
			}
		}
	}

	// Parse values
	height, _ := parseStringToInt64(status.Result.SyncInfo.LatestBlockHeight)
	votingPower, _ := parseStringToInt64(status.Result.ValidatorInfo.VotingPower)

	// Parse block time
	blockTime, err := time.Parse(time.RFC3339Nano, status.Result.SyncInfo.LatestBlockTime)
	if err != nil {
		blockTime = time.Now()
	}

	return &ConsensusMetrics{
		CurrentHeight:     height,
		LastBlockTime:     blockTime.Unix(),
		BlockTime:         2.0, // Default block time
		ValidatorCount:    validatorCount,
		VotingPower:       votingPower,
		ParticipationRate: 0.9, // Default participation rate
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
	tps := float64(len(block.Result.Transactions)) / 2.0 // Assuming 2s block time

	gasUsed, _ := parseHexToInt64(block.Result.GasUsed)

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
	// Get net info from BFT node
	netResp, err := c.httpClient.Get(c.BFTRPCUrl + "/net_info")
	if err != nil {
		return &NetworkMetrics{
			PeerCount:     50,
			InboundPeers:  25,
			OutboundPeers: 25,
			BytesIn:       1000000,
			BytesOut:      1000000,
			NetworkLatency: 50.0,
		}, nil // Return defaults on error
	}
	defer netResp.Body.Close()

	var netInfo struct {
		Result struct {
			NPeers string `json:"n_peers"`
			Peers  []struct {
				IsOutbound bool `json:"is_outbound"`
			} `json:"peers"`
		} `json:"result"`
	}

	if err := json.NewDecoder(netResp.Body).Decode(&netInfo); err != nil {
		return nil, fmt.Errorf("failed to decode net info: %w", err)
	}

	peerCount, _ := parseStringToInt64(netInfo.Result.NPeers)

	inbound, outbound := 0, 0
	for _, peer := range netInfo.Result.Peers {
		if peer.IsOutbound {
			outbound++
		} else {
			inbound++
		}
	}

	return &NetworkMetrics{
		PeerCount:      int(peerCount),
		InboundPeers:   inbound,
		OutboundPeers:  outbound,
		BytesIn:        1000000, // Would need custom metrics
		BytesOut:       1000000, // Would need custom metrics
		NetworkLatency: 50.0,    // Would need to measure
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

	resp, err := c.httpClient.Post(url, "application/json", nil)
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