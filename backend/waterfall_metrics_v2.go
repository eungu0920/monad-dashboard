package main

import (
	"sync"
	"sync/atomic"
	"time"
)

// MonadWaterfallMetrics tracks metrics for Monad's transaction lifecycle
// Based on: https://docs.monad.xyz/monad-arch/transaction-lifecycle
type MonadWaterfallMetrics struct {
	// Stage 1: Submission (Network Ingress)
	SubmissionRPCReceived atomic.Int64
	SubmissionP2PReceived atomic.Int64
	SubmissionInvalidSig  atomic.Int64
	SubmissionInvalidFormat atomic.Int64

	// Stage 2: Mempool (Validation & Leader Propagation)
	MempoolReceived          atomic.Int64
	MempoolNonceInvalid      atomic.Int64
	MempoolGasTooHigh        atomic.Int64
	MempoolPropagationFailed atomic.Int64
	MempoolToBlockBuilding   atomic.Int64

	// Stage 3: Block Building (Inclusion Checks at Consensus Time)
	BlockBuildingReceived       atomic.Int64
	BlockBuildingInsufficientBalance atomic.Int64
	BlockBuildingNonceGap       atomic.Int64
	BlockBuildingBlockFull      atomic.Int64
	BlockBuildingToConsensus    atomic.Int64

	// Stage 4: Consensus (MonadBFT: Proposed → Voted → Finalized)
	ConsensusProposed       atomic.Int64
	ConsensusVoted          atomic.Int64
	ConsensusFinalized      atomic.Int64
	ConsensusRejected       atomic.Int64
	ConsensusToExecution    atomic.Int64

	// Stage 5: Execution (Parallel Processing)
	ExecutionParallelSuccess atomic.Int64
	ExecutionParallelRetry   atomic.Int64
	ExecutionReverted        atomic.Int64
	ExecutionToStateUpdate   atomic.Int64

	// Stage 6: State Update (Serial Commitment)
	StateAccountsUpdated atomic.Int64
	StateStorageWrites   atomic.Int64
	StateLogsEmitted     atomic.Int64
	StateToFinality      atomic.Int64

	// Stage 7: Finality (2-Block Confirmation)
	FinalityQueryable       atomic.Int64
	FinalityReceiptsGenerated atomic.Int64

	// Timing metrics (nanoseconds)
	MempoolPropagationLatencyNs atomic.Int64
	ConsensusLatencyNs          atomic.Int64
	ExecutionLatencyNs          atomic.Int64
	FinalityLatencyNs           atomic.Int64

	// Last reset time
	lastReset time.Time
	mu        sync.RWMutex
}

// NewMonadWaterfallMetrics creates a new Monad waterfall metrics tracker
func NewMonadWaterfallMetrics() *MonadWaterfallMetrics {
	return &MonadWaterfallMetrics{
		lastReset: time.Now(),
	}
}

// Snapshot returns current values as a structured map
func (m *MonadWaterfallMetrics) Snapshot() map[string]interface{} {
	return map[string]interface{}{
		"submission": map[string]interface{}{
			"rpc_received":     m.SubmissionRPCReceived.Load(),
			"p2p_received":     m.SubmissionP2PReceived.Load(),
			"invalid_sig":      m.SubmissionInvalidSig.Load(),
			"invalid_format":   m.SubmissionInvalidFormat.Load(),
		},
		"mempool": map[string]interface{}{
			"received":            m.MempoolReceived.Load(),
			"nonce_invalid":       m.MempoolNonceInvalid.Load(),
			"gas_too_high":        m.MempoolGasTooHigh.Load(),
			"propagation_failed":  m.MempoolPropagationFailed.Load(),
			"to_block_building":   m.MempoolToBlockBuilding.Load(),
		},
		"block_building": map[string]interface{}{
			"received":              m.BlockBuildingReceived.Load(),
			"insufficient_balance":  m.BlockBuildingInsufficientBalance.Load(),
			"nonce_gap":             m.BlockBuildingNonceGap.Load(),
			"block_full":            m.BlockBuildingBlockFull.Load(),
			"to_consensus":          m.BlockBuildingToConsensus.Load(),
		},
		"consensus": map[string]interface{}{
			"proposed":       m.ConsensusProposed.Load(),
			"voted":          m.ConsensusVoted.Load(),
			"finalized":      m.ConsensusFinalized.Load(),
			"rejected":       m.ConsensusRejected.Load(),
			"to_execution":   m.ConsensusToExecution.Load(),
		},
		"execution": map[string]interface{}{
			"parallel_success":  m.ExecutionParallelSuccess.Load(),
			"parallel_retry":    m.ExecutionParallelRetry.Load(),
			"reverted":          m.ExecutionReverted.Load(),
			"to_state_update":   m.ExecutionToStateUpdate.Load(),
		},
		"state_update": map[string]interface{}{
			"accounts_updated": m.StateAccountsUpdated.Load(),
			"storage_writes":   m.StateStorageWrites.Load(),
			"logs_emitted":     m.StateLogsEmitted.Load(),
			"to_finality":      m.StateToFinality.Load(),
		},
		"finality": map[string]interface{}{
			"queryable":          m.FinalityQueryable.Load(),
			"receipts_generated": m.FinalityReceiptsGenerated.Load(),
		},
		"timing": map[string]interface{}{
			"mempool_propagation_latency_ns": m.MempoolPropagationLatencyNs.Load(),
			"consensus_latency_ns":           m.ConsensusLatencyNs.Load(),
			"execution_latency_ns":           m.ExecutionLatencyNs.Load(),
			"finality_latency_ns":            m.FinalityLatencyNs.Load(),
		},
	}
}

// Reset resets all counters
func (m *MonadWaterfallMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset all atomic counters
	m.SubmissionRPCReceived.Store(0)
	m.SubmissionP2PReceived.Store(0)
	m.SubmissionInvalidSig.Store(0)
	m.SubmissionInvalidFormat.Store(0)

	m.MempoolReceived.Store(0)
	m.MempoolNonceInvalid.Store(0)
	m.MempoolGasTooHigh.Store(0)
	m.MempoolPropagationFailed.Store(0)
	m.MempoolToBlockBuilding.Store(0)

	m.BlockBuildingReceived.Store(0)
	m.BlockBuildingInsufficientBalance.Store(0)
	m.BlockBuildingNonceGap.Store(0)
	m.BlockBuildingBlockFull.Store(0)
	m.BlockBuildingToConsensus.Store(0)

	m.ConsensusProposed.Store(0)
	m.ConsensusVoted.Store(0)
	m.ConsensusFinalized.Store(0)
	m.ConsensusRejected.Store(0)
	m.ConsensusToExecution.Store(0)

	m.ExecutionParallelSuccess.Store(0)
	m.ExecutionParallelRetry.Store(0)
	m.ExecutionReverted.Store(0)
	m.ExecutionToStateUpdate.Store(0)

	m.StateAccountsUpdated.Store(0)
	m.StateStorageWrites.Store(0)
	m.StateLogsEmitted.Store(0)
	m.StateToFinality.Store(0)

	m.FinalityQueryable.Store(0)
	m.FinalityReceiptsGenerated.Store(0)

	m.lastReset = time.Now()
}

// GetElapsedSeconds returns seconds since last reset
func (m *MonadWaterfallMetrics) GetElapsedSeconds() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Since(m.lastReset).Seconds()
}

// Global Monad waterfall metrics instance
var (
	monadWaterfallMetrics *MonadWaterfallMetrics
	monadWaterfallMutex   sync.RWMutex
)

func init() {
	monadWaterfallMetrics = NewMonadWaterfallMetrics()
}

// GetMonadWaterfallMetrics returns the global Monad waterfall metrics
func GetMonadWaterfallMetrics() *MonadWaterfallMetrics {
	monadWaterfallMutex.RLock()
	defer monadWaterfallMutex.RUnlock()
	return monadWaterfallMetrics
}

// GenerateMonadWaterfall generates waterfall data matching Monad's transaction lifecycle
// Priority: Prometheus > IPC > Block Estimation > Mock
func GenerateMonadWaterfall() map[string]interface{} {
	// Priority 1: Try Prometheus metrics (most comprehensive)
	promCollector := GetPrometheusCollector()
	if promCollector != nil && promCollector.IsHealthy() {
		promMetrics := promCollector.GetMetrics()
		// Check if we have ACTIVE txpool metrics
		if promMetrics.TPS60s > 0 && (promMetrics.InsertOwnedTxsRate > 0 || promMetrics.InsertForwardedTxsRate > 0) {
			return generateMonadWaterfallFromPrometheus(promMetrics)
		}
	}

	// Priority 2: Try IPC metrics
	ipcCollector := GetIPCCollector()
	if ipcCollector != nil && ipcCollector.IsHealthy() {
		return generateMonadWaterfallFromIPC(ipcCollector.GetMetrics())
	}

	// Priority 3: Fallback to block-based estimation
	if monadSubscriber != nil && monadSubscriber.IsConnected() {
		block := monadSubscriber.GetLatestBlock()
		if block != nil {
			return generateMonadWaterfallFromBlock(block)
		}
	}

	// Priority 4: Mock data for testing
	return generateMonadMockWaterfall()
}

// generateMonadWaterfallFromPrometheus generates Monad-aligned waterfall from Prometheus metrics
func generateMonadWaterfallFromPrometheus(metrics *PrometheusMetrics) map[string]interface{} {
	// Collection interval for rate-to-count conversion
	interval := 5.0

	// Stage 1: Submission
	rpcReceived := int64(metrics.InsertOwnedTxsRate * interval)
	p2pReceived := int64(metrics.InsertForwardedTxsRate * interval)
	invalidSig := int64(metrics.DropInvalidSignatureRate * interval)

	// Stage 2: Mempool
	toMempool := rpcReceived + p2pReceived - invalidSig
	nonceInvalid := int64(metrics.DropNonceTooLowRate * interval)

	// Stage 3: Block Building
	insufficientBalance := int64(metrics.DropInsufficientBalanceRate * interval)
	blockFull := int64(metrics.DropPoolFullRate * interval)
	feeDropped := int64(metrics.DropFeeTooLowRate * interval)

	// Stage 4: Consensus - get from ConsensusTracker
	consensusTracker := GetConsensusTracker()
	consensusState := consensusTracker.GetConsensusState()

	// Build nodes array for Sankey diagram
	nodes := []map[string]interface{}{
		{"id": "submission_rpc", "label": "RPC", "color": "#4CAF50"},
		{"id": "submission_p2p", "label": "P2P", "color": "#2196F3"},
		{"id": "mempool", "label": "Mempool", "color": "#FF9800"},
		{"id": "block_building", "label": "Block Building", "color": "#9C27B0"},
		{"id": "consensus_proposed", "label": "Proposed", "color": "#3F51B5"},
		{"id": "consensus_voted", "label": "Voted", "color": "#FFC107"},
		{"id": "consensus_finalized", "label": "Finalized", "color": "#4CAF50"},
		{"id": "execution", "label": "Execution", "color": "#F44336"},
		{"id": "state_update", "label": "State Update", "color": "#00BCD4"},
		{"id": "finality", "label": "Final (Queryable)", "color": "#8BC34A"},
		{"id": "dropped", "label": "Dropped", "color": "#757575"},
	}

	// Calculate flows
	toBlockBuilding := toMempool - nonceInvalid
	toConsensus := toBlockBuilding - insufficientBalance - blockFull - feeDropped
	toExecution := toConsensus
	toStateUpdate := toExecution
	toFinality := toStateUpdate

	// Build links array for Sankey diagram
	links := []map[string]interface{}{}

	// Submission → Mempool
	if rpcReceived > 0 {
		links = append(links, map[string]interface{}{
			"source": "submission_rpc",
			"target": "mempool",
			"value":  rpcReceived,
		})
	}
	if p2pReceived > 0 {
		links = append(links, map[string]interface{}{
			"source": "submission_p2p",
			"target": "mempool",
			"value":  p2pReceived,
		})
	}

	// Mempool → Block Building / Dropped
	if toBlockBuilding > 0 {
		links = append(links, map[string]interface{}{
			"source": "mempool",
			"target": "block_building",
			"value":  toBlockBuilding,
		})
	}
	if invalidSig+nonceInvalid > 0 {
		links = append(links, map[string]interface{}{
			"source": "mempool",
			"target": "dropped",
			"value":  invalidSig + nonceInvalid,
		})
	}

	// Block Building → Consensus / Dropped
	if toConsensus > 0 {
		links = append(links, map[string]interface{}{
			"source": "block_building",
			"target": "consensus_proposed",
			"value":  toConsensus,
		})
	}
	if insufficientBalance+blockFull+feeDropped > 0 {
		links = append(links, map[string]interface{}{
			"source": "block_building",
			"target": "dropped",
			"value":  insufficientBalance + blockFull + feeDropped,
		})
	}

	// Consensus: Proposed → Voted → Finalized
	if toConsensus > 0 {
		links = append(links, map[string]interface{}{
			"source": "consensus_proposed",
			"target": "consensus_voted",
			"value":  toConsensus,
		})
		links = append(links, map[string]interface{}{
			"source": "consensus_voted",
			"target": "consensus_finalized",
			"value":  toConsensus,
		})
		links = append(links, map[string]interface{}{
			"source": "consensus_finalized",
			"target": "execution",
			"value":  toExecution,
		})
	}

	// Execution → State Update
	if toStateUpdate > 0 {
		links = append(links, map[string]interface{}{
			"source": "execution",
			"target": "state_update",
			"value":  toStateUpdate,
		})
	}

	// State Update → Finality
	if toFinality > 0 {
		links = append(links, map[string]interface{}{
			"source": "state_update",
			"target": "finality",
			"value":  toFinality,
		})
	}

	return map[string]interface{}{
		"nodes": nodes,
		"links": links,
		"metadata": map[string]interface{}{
			"source":            "prometheus_metrics",
			"last_updated":      metrics.LastUpdated.Unix(),
			"tps":               metrics.TPS60s,
			"pending_txs":       int64(metrics.PendingTxs),
			"tracked_txs":       int64(metrics.TrackedTxs),
			"interval_seconds":  interval,
			"consensus_state":   consensusState,
		},
		"drops": map[string]interface{}{
			"invalid_signature":     invalidSig,
			"nonce_invalid":         nonceInvalid,
			"insufficient_balance":  insufficientBalance,
			"block_full":            blockFull,
			"fee_too_low":           feeDropped,
		},
	}
}

// generateMonadWaterfallFromIPC generates waterfall from IPC metrics
func generateMonadWaterfallFromIPC(metrics *MonadRealMetrics) map[string]interface{} {
	// Similar structure to Prometheus but using IPC data
	// TODO: Implement when IPC metrics are available
	return generateMonadMockWaterfall()
}

// generateMonadWaterfallFromBlock generates estimated waterfall from block data
func generateMonadWaterfallFromBlock(block *BlockHeader) map[string]interface{} {
	txCount := int64(block.Transactions)

	// Realistic estimation based on typical Monad behavior
	rpcReceived := txCount * 5 / 10  // 50% RPC
	p2pReceived := txCount * 5 / 10  // 50% P2P

	// Drops (realistic low percentages)
	invalidSig := (rpcReceived + p2pReceived) / 100    // 1%
	nonceInvalid := (rpcReceived + p2pReceived) / 200  // 0.5%
	insufficientBalance := (rpcReceived + p2pReceived) / 500  // 0.2%

	toMempool := rpcReceived + p2pReceived - invalidSig
	toBlockBuilding := toMempool - nonceInvalid
	toConsensus := toBlockBuilding - insufficientBalance

	nodes := []map[string]interface{}{
		{"id": "submission_rpc", "label": "RPC", "color": "#4CAF50"},
		{"id": "submission_p2p", "label": "P2P", "color": "#2196F3"},
		{"id": "mempool", "label": "Mempool", "color": "#FF9800"},
		{"id": "block_building", "label": "Block Building", "color": "#9C27B0"},
		{"id": "consensus_proposed", "label": "Proposed", "color": "#3F51B5"},
		{"id": "consensus_voted", "label": "Voted", "color": "#FFC107"},
		{"id": "consensus_finalized", "label": "Finalized", "color": "#4CAF50"},
		{"id": "execution", "label": "Execution", "color": "#F44336"},
		{"id": "state_update", "label": "State Update", "color": "#00BCD4"},
		{"id": "finality", "label": "Final (Queryable)", "color": "#8BC34A"},
		{"id": "dropped", "label": "Dropped", "color": "#757575"},
	}

	links := []map[string]interface{}{
		{"source": "submission_rpc", "target": "mempool", "value": rpcReceived},
		{"source": "submission_p2p", "target": "mempool", "value": p2pReceived},
		{"source": "mempool", "target": "block_building", "value": toBlockBuilding},
		{"source": "mempool", "target": "dropped", "value": invalidSig + nonceInvalid},
		{"source": "block_building", "target": "consensus_proposed", "value": toConsensus},
		{"source": "block_building", "target": "dropped", "value": insufficientBalance},
		{"source": "consensus_proposed", "target": "consensus_voted", "value": toConsensus},
		{"source": "consensus_voted", "target": "consensus_finalized", "value": toConsensus},
		{"source": "consensus_finalized", "target": "execution", "value": toConsensus},
		{"source": "execution", "target": "state_update", "value": toConsensus},
		{"source": "state_update", "target": "finality", "value": toConsensus},
	}

	consensusTracker := GetConsensusTracker()

	return map[string]interface{}{
		"nodes": nodes,
		"links": links,
		"metadata": map[string]interface{}{
			"source":         "block_estimation",
			"block_height":   block.Number,
			"block_hash":     block.Hash,
			"block_txs":      block.Transactions,
			"timestamp":      block.Timestamp,
			"consensus_state": consensusTracker.GetConsensusState(),
		},
	}
}

// generateMonadMockWaterfall generates mock data for testing
func generateMonadMockWaterfall() map[string]interface{} {
	nodes := []map[string]interface{}{
		{"id": "submission_rpc", "label": "RPC", "color": "#4CAF50"},
		{"id": "submission_p2p", "label": "P2P", "color": "#2196F3"},
		{"id": "mempool", "label": "Mempool", "color": "#FF9800"},
		{"id": "block_building", "label": "Block Building", "color": "#9C27B0"},
		{"id": "consensus_proposed", "label": "Proposed", "color": "#3F51B5"},
		{"id": "consensus_voted", "label": "Voted", "color": "#FFC107"},
		{"id": "consensus_finalized", "label": "Finalized", "color": "#4CAF50"},
		{"id": "execution", "label": "Execution", "color": "#F44336"},
		{"id": "state_update", "label": "State Update", "color": "#00BCD4"},
		{"id": "finality", "label": "Final (Queryable)", "color": "#8BC34A"},
		{"id": "dropped", "label": "Dropped", "color": "#757575"},
	}

	links := []map[string]interface{}{
		{"source": "submission_rpc", "target": "mempool", "value": 700},
		{"source": "submission_p2p", "target": "mempool", "value": 300},
		{"source": "mempool", "target": "block_building", "value": 950},
		{"source": "mempool", "target": "dropped", "value": 50},
		{"source": "block_building", "target": "consensus_proposed", "value": 930},
		{"source": "block_building", "target": "dropped", "value": 20},
		{"source": "consensus_proposed", "target": "consensus_voted", "value": 930},
		{"source": "consensus_voted", "target": "consensus_finalized", "value": 930},
		{"source": "consensus_finalized", "target": "execution", "value": 930},
		{"source": "execution", "target": "state_update", "value": 925},
		{"source": "execution", "target": "dropped", "value": 5},
		{"source": "state_update", "target": "finality", "value": 925},
	}

	return map[string]interface{}{
		"nodes": nodes,
		"links": links,
		"metadata": map[string]interface{}{
			"source": "mock_data",
			"consensus_state": map[string]interface{}{
				"current_block":    100,
				"finalized_block":  98,
				"blocks_behind":    2,
				"proposed_blocks":  1,
				"voted_blocks":     1,
				"finalized_blocks": 1,
			},
		},
	}
}
