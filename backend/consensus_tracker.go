package main

import (
	"sync"
	"time"
)

// BlockConsensusState represents the consensus phase state of a block
type BlockConsensusState struct {
	BlockNumber uint64     `json:"block_number"`
	BlockHash   string     `json:"block_hash"`
	Phase       string     `json:"phase"` // "proposed", "voted", "finalized"
	ProposedAt  time.Time  `json:"proposed_at"`
	VotedAt     *time.Time `json:"voted_at,omitempty"`
	FinalizedAt *time.Time `json:"finalized_at,omitempty"`
	TxCount     int        `json:"tx_count"`
}

// ConsensusTracker tracks MonadBFT consensus phases for blocks
type ConsensusTracker struct {
	blocks         map[uint64]*BlockConsensusState
	currentBlock   uint64
	finalizedBlock uint64
	mu             sync.RWMutex
	maxHistory     int // Maximum number of blocks to track
}

// Global consensus tracker instance
var consensusTracker *ConsensusTracker

// InitializeConsensusTracker creates a new consensus tracker
func InitializeConsensusTracker() *ConsensusTracker {
	consensusTracker = &ConsensusTracker{
		blocks:     make(map[uint64]*BlockConsensusState),
		maxHistory: 20, // Track last 20 blocks
	}
	return consensusTracker
}

// GetConsensusTracker returns the global consensus tracker instance
func GetConsensusTracker() *ConsensusTracker {
	if consensusTracker == nil {
		return InitializeConsensusTracker()
	}
	return consensusTracker
}

// OnBlockProposed records when a block is proposed
func (ct *ConsensusTracker) OnBlockProposed(blockNum uint64, hash string, txCount int) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Update current block
	if blockNum > ct.currentBlock {
		ct.currentBlock = blockNum
	}

	// Create or update block state
	if _, exists := ct.blocks[blockNum]; !exists {
		ct.blocks[blockNum] = &BlockConsensusState{
			BlockNumber: blockNum,
			BlockHash:   hash,
			Phase:       "proposed",
			ProposedAt:  time.Now(),
			TxCount:     txCount,
		}
	}

	// Automatically mark previous blocks as voted/finalized based on MonadBFT rules
	ct.updatePhases(blockNum)

	// Clean up old blocks
	ct.cleanupOldBlocks()
}

// updatePhases automatically updates block phases based on MonadBFT timing
// Voted: after 1 block
// Finalized: after 2 blocks
func (ct *ConsensusTracker) updatePhases(currentBlockNum uint64) {
	now := time.Now()

	// Block N-1 should be voted
	if currentBlockNum >= 1 {
		votedBlockNum := currentBlockNum - 1
		if block, exists := ct.blocks[votedBlockNum]; exists {
			if block.Phase == "proposed" {
				block.Phase = "voted"
				block.VotedAt = &now
			}
		}
	}

	// Block N-2 should be finalized
	if currentBlockNum >= 2 {
		finalizedBlockNum := currentBlockNum - 2
		if block, exists := ct.blocks[finalizedBlockNum]; exists {
			if block.Phase != "finalized" {
				block.Phase = "finalized"
				block.FinalizedAt = &now
				ct.finalizedBlock = finalizedBlockNum
			}
		}
	}
}

// OnBlockVoted explicitly marks a block as voted (if real consensus data is available)
func (ct *ConsensusTracker) OnBlockVoted(blockNum uint64) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if block, exists := ct.blocks[blockNum]; exists {
		now := time.Now()
		block.Phase = "voted"
		block.VotedAt = &now
	}
}

// OnBlockFinalized explicitly marks a block as finalized (if real consensus data is available)
func (ct *ConsensusTracker) OnBlockFinalized(blockNum uint64) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if block, exists := ct.blocks[blockNum]; exists {
		now := time.Now()
		block.Phase = "finalized"
		block.FinalizedAt = &now
		if blockNum > ct.finalizedBlock {
			ct.finalizedBlock = blockNum
		}
	}
}

// GetRecentBlocks returns the N most recent blocks
func (ct *ConsensusTracker) GetRecentBlocks(count int) []BlockConsensusState {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if count > len(ct.blocks) {
		count = len(ct.blocks)
	}

	// Collect blocks and sort by block number
	blocks := make([]BlockConsensusState, 0, len(ct.blocks))
	for _, block := range ct.blocks {
		blocks = append(blocks, *block)
	}

	// Sort descending by block number
	for i := 0; i < len(blocks); i++ {
		for j := i + 1; j < len(blocks); j++ {
			if blocks[i].BlockNumber < blocks[j].BlockNumber {
				blocks[i], blocks[j] = blocks[j], blocks[i]
			}
		}
	}

	// Return top N
	if len(blocks) > count {
		blocks = blocks[:count]
	}

	return blocks
}

// GetConsensusState returns current consensus state summary
func (ct *ConsensusTracker) GetConsensusState() map[string]interface{} {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	proposedCount := 0
	votedCount := 0
	finalizedCount := 0

	for _, block := range ct.blocks {
		switch block.Phase {
		case "proposed":
			proposedCount++
		case "voted":
			votedCount++
		case "finalized":
			finalizedCount++
		}
	}

	return map[string]interface{}{
		"current_block":     ct.currentBlock,
		"finalized_block":   ct.finalizedBlock,
		"blocks_behind":     ct.currentBlock - ct.finalizedBlock,
		"proposed_blocks":   proposedCount,
		"voted_blocks":      votedCount,
		"finalized_blocks":  finalizedCount,
		"recent_blocks":     ct.GetRecentBlocks(10),
	}
}

// GetBlockPhase returns the consensus phase of a specific block
func (ct *ConsensusTracker) GetBlockPhase(blockNum uint64) string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if block, exists := ct.blocks[blockNum]; exists {
		return block.Phase
	}
	return "unknown"
}

// GetPhaseProgress returns progress percentage for a block (0-100)
func (ct *ConsensusTracker) GetPhaseProgress(blockNum uint64) int {
	phase := ct.GetBlockPhase(blockNum)
	switch phase {
	case "proposed":
		return 33
	case "voted":
		return 66
	case "finalized":
		return 100
	default:
		return 0
	}
}

// cleanupOldBlocks removes blocks older than maxHistory to prevent memory leak
func (ct *ConsensusTracker) cleanupOldBlocks() {
	if len(ct.blocks) <= ct.maxHistory {
		return
	}

	// Find the threshold block number
	threshold := ct.currentBlock - uint64(ct.maxHistory)

	// Remove blocks older than threshold
	for blockNum := range ct.blocks {
		if blockNum < threshold {
			delete(ct.blocks, blockNum)
		}
	}
}

// GetMetrics returns consensus metrics for monitoring
func (ct *ConsensusTracker) GetMetrics() map[string]interface{} {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	// Calculate average finalization time (proposed → finalized)
	var totalFinalizationTime time.Duration
	var finalizedBlocksCount int

	for _, block := range ct.blocks {
		if block.FinalizedAt != nil {
			duration := block.FinalizedAt.Sub(block.ProposedAt)
			totalFinalizationTime += duration
			finalizedBlocksCount++
		}
	}

	avgFinalizationTime := float64(0)
	if finalizedBlocksCount > 0 {
		avgFinalizationTime = totalFinalizationTime.Seconds() / float64(finalizedBlocksCount)
	}

	return map[string]interface{}{
		"current_block":           ct.currentBlock,
		"finalized_block":         ct.finalizedBlock,
		"finality_lag":            ct.currentBlock - ct.finalizedBlock,
		"avg_finalization_time":   avgFinalizationTime,
		"tracked_blocks":          len(ct.blocks),
	}
}
