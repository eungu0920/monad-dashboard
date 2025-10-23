package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// MonadIPCCollector collects real-time metrics from Monad via IPC
type MonadIPCCollector struct {
	ipcPath string
	conn    net.Conn
	mu      sync.RWMutex

	// Real-time counters from Monad
	metrics *MonadRealMetrics
}

// MonadRealMetrics represents actual metrics from Monad node
type MonadRealMetrics struct {
	// TxPool metrics (from monad-eth-txpool/src/metrics.rs)
	InsertOwnedTxs       int64 // RPC 트랜잭션
	InsertForwardedTxs   int64 // P2P 트랜잭션

	DropNotWellFormed    int64
	DropInvalidSignature int64 // verify_failed
	DropNonceTooLow      int64 // nonce_failed
	DropFeeTooLow        int64 // fee_too_low
	DropInsufficientBalance int64 // balance_failed
	DropPoolFull         int64 // pool_full

	CreateProposal       int64
	CreateProposalTxs    int64

	// Pending pool
	PendingAddresses     int64
	PendingTxs           int64
	PendingPromoteTxs    int64

	// Tracked pool
	TrackedAddresses     int64
	TrackedTxs           int64

	// Execution metrics (would come from monad execution layer)
	ParallelSuccess      int64
	SequentialFallback   int64
	StateReads           int64
	StateWrites          int64

	LastUpdated time.Time
}

// NewMonadIPCCollector creates a new IPC-based metrics collector
func NewMonadIPCCollector(ipcPath string) *MonadIPCCollector {
	return &MonadIPCCollector{
		ipcPath: ipcPath,
		metrics: &MonadRealMetrics{
			LastUpdated: time.Now(),
		},
	}
}

// Connect establishes connection to Monad IPC socket
func (c *MonadIPCCollector) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := net.Dial("unix", c.ipcPath)
	if err != nil {
		return fmt.Errorf("failed to connect to Monad IPC %s: %w", c.ipcPath, err)
	}

	c.conn = conn
	log.Printf("Connected to Monad IPC: %s", c.ipcPath)

	// Start metrics collection goroutine
	go c.collectMetrics()

	return nil
}

// collectMetrics continuously collects metrics from Monad
func (c *MonadIPCCollector) collectMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.requestMetrics(); err != nil {
			log.Printf("Error collecting metrics: %v", err)
			continue
		}
	}
}

// requestMetrics requests current metrics snapshot from Monad
func (c *MonadIPCCollector) requestMetrics() error {
	// Create a new connection for each request to avoid broken pipe
	conn, err := net.Dial("unix", c.ipcPath)
	if err != nil {
		return fmt.Errorf("failed to dial IPC: %w", err)
	}
	defer conn.Close()

	// Send metrics request (JSON-RPC style)
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      time.Now().Unix(),
		"method":  "monad_getMetrics",
		"params":  []interface{}{},
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Write request
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Write(append(requestBytes, '\n')); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	// Read response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var response struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int64  `json:"id"`
		Result  struct {
			// TxPool metrics
			TxPool struct {
				InsertOwnedTxs       int64 `json:"insert_owned_txs"`
				InsertForwardedTxs   int64 `json:"insert_forwarded_txs"`
				DropNotWellFormed    int64 `json:"drop_not_well_formed"`
				DropInvalidSignature int64 `json:"drop_invalid_signature"`
				DropNonceTooLow      int64 `json:"drop_nonce_too_low"`
				DropFeeTooLow        int64 `json:"drop_fee_too_low"`
				DropInsufficientBalance int64 `json:"drop_insufficient_balance"`
				DropPoolFull         int64 `json:"drop_pool_full"`
				CreateProposal       int64 `json:"create_proposal"`
				CreateProposalTxs    int64 `json:"create_proposal_txs"`
				Pending struct {
					Addresses  int64 `json:"addresses"`
					Txs        int64 `json:"txs"`
					PromoteTxs int64 `json:"promote_txs"`
				} `json:"pending"`
				Tracked struct {
					Addresses int64 `json:"addresses"`
					Txs       int64 `json:"txs"`
				} `json:"tracked"`
			} `json:"txpool"`

			// Execution metrics
			Execution struct {
				ParallelSuccess    int64 `json:"parallel_success"`
				SequentialFallback int64 `json:"sequential_fallback"`
				StateReads         int64 `json:"state_reads"`
				StateWrites        int64 `json:"state_writes"`
			} `json:"execution"`
		} `json:"result"`
	}

	if err := json.Unmarshal(buffer[:n], &response); err != nil {
		// IPC might not support this method yet, fallback to estimation
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Update metrics
	c.mu.Lock()
	c.metrics.InsertOwnedTxs = response.Result.TxPool.InsertOwnedTxs
	c.metrics.InsertForwardedTxs = response.Result.TxPool.InsertForwardedTxs
	c.metrics.DropNotWellFormed = response.Result.TxPool.DropNotWellFormed
	c.metrics.DropInvalidSignature = response.Result.TxPool.DropInvalidSignature
	c.metrics.DropNonceTooLow = response.Result.TxPool.DropNonceTooLow
	c.metrics.DropFeeTooLow = response.Result.TxPool.DropFeeTooLow
	c.metrics.DropInsufficientBalance = response.Result.TxPool.DropInsufficientBalance
	c.metrics.DropPoolFull = response.Result.TxPool.DropPoolFull
	c.metrics.CreateProposal = response.Result.TxPool.CreateProposal
	c.metrics.CreateProposalTxs = response.Result.TxPool.CreateProposalTxs
	c.metrics.PendingAddresses = response.Result.TxPool.Pending.Addresses
	c.metrics.PendingTxs = response.Result.TxPool.Pending.Txs
	c.metrics.PendingPromoteTxs = response.Result.TxPool.Pending.PromoteTxs
	c.metrics.TrackedAddresses = response.Result.TxPool.Tracked.Addresses
	c.metrics.TrackedTxs = response.Result.TxPool.Tracked.Txs
	c.metrics.ParallelSuccess = response.Result.Execution.ParallelSuccess
	c.metrics.SequentialFallback = response.Result.Execution.SequentialFallback
	c.metrics.StateReads = response.Result.Execution.StateReads
	c.metrics.StateWrites = response.Result.Execution.StateWrites
	c.metrics.LastUpdated = time.Now()
	c.mu.Unlock()

	log.Printf("Updated real metrics: RPC=%d, P2P=%d, SigFailed=%d, Parallel=%d",
		c.metrics.InsertOwnedTxs, c.metrics.InsertForwardedTxs,
		c.metrics.DropInvalidSignature, c.metrics.ParallelSuccess)

	return nil
}

// GetMetrics returns current metrics snapshot
func (c *MonadIPCCollector) GetMetrics() *MonadRealMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid race conditions
	metricsCopy := *c.metrics
	return &metricsCopy
}

// IsHealthy checks if metrics are recent
func (c *MonadIPCCollector) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Metrics should be updated within last 5 seconds
	return time.Since(c.metrics.LastUpdated) < 5*time.Second
}

// Close closes the IPC connection
func (c *MonadIPCCollector) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Global IPC collector instance
var (
	ipcCollector   *MonadIPCCollector
	ipcCollectorMu sync.RWMutex
)

// InitializeIPCCollector initializes the IPC metrics collector
func InitializeIPCCollector(ipcPath string) error {
	ipcCollectorMu.Lock()
	defer ipcCollectorMu.Unlock()

	ipcCollector = NewMonadIPCCollector(ipcPath)

	// Try to connect (fallback to estimation if fails)
	if err := ipcCollector.Connect(); err != nil {
		log.Printf("Failed to connect to Monad IPC: %v", err)
		log.Printf("Will use estimation-based metrics")
		return err
	}

	log.Printf("IPC collector initialized successfully")
	return nil
}

// GetIPCCollector returns the global IPC collector
func GetIPCCollector() *MonadIPCCollector {
	ipcCollectorMu.RLock()
	defer ipcCollectorMu.RUnlock()
	return ipcCollector
}
