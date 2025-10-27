# Monad Waterfall Redesign - Technical Specification

## Overview

This document outlines the redesign of the transaction waterfall to accurately reflect Monad's architecture as described in the official documentation.

---

## New Waterfall Structure

Based on Monad's transaction lifecycle, we define **7 stages** that transactions flow through:

### Stage 1: **Submission** (Network Ingress)
**Description**: User transactions enter the system via RPC or P2P gossip

**Inputs**:
- `rpc_received`: Transactions from eth_sendTransaction/eth_sendRawTransaction
- `p2p_received`: Transactions from peer gossip network

**Outputs**:
- `invalid_signature`: Failed signature verification
- `invalid_format`: Malformed transaction data
- `to_mempool`: Valid transactions forwarded to mempool

**Prometheus Metrics**:
- `monad_bft_txpool_pool_insert_owned_txs` (RPC)
- `monad_bft_txpool_pool_insert_forwarded_txs` (P2P)

---

### Stage 2: **Mempool** (Validation & Leader Propagation)
**Description**: RPC nodes validate and propagate to N leaders (with K retries)

**Inputs**:
- `from_submission`: Transactions from Stage 1

**Validation Checks**:
- Signature verification
- Nonce ordering
- Gas limit validation

**Outputs**:
- `nonce_invalid`: Non-sequential nonce
- `gas_too_high`: Gas limit exceeds block maximum
- `propagation_failed`: Failed after K leader attempts
- `to_block_building`: Successfully propagated to leaders

**Prometheus Metrics**:
- `monad_bft_txpool_pool_drop_nonce_too_low`
- `monad_bft_txpool_pool_drop_gas_limit_exceeded` (if available)
- `monad_bft_mempool_propagation_attempts` (if available)

---

### Stage 3: **Block Building** (Inclusion Checks)
**Description**: Leaders perform dynamic checks at consensus time

**Inputs**:
- `from_mempool`: Transactions ready for inclusion

**Dynamic Checks**:
- Account balance sufficient
- Nonce is contiguous
- Block has available space

**Outputs**:
- `insufficient_balance`: Account cannot pay gas
- `nonce_gap`: Nonce not contiguous
- `block_full`: No space in current block
- `to_consensus`: Transactions included in block

**Prometheus Metrics**:
- `monad_bft_txpool_pool_drop_insufficient_balance`
- `monad_bft_txpool_pool_drop_pool_full`
- `monad_bft_block_building_txs_selected` (if available)

---

### Stage 4: **Consensus** (RaptorCast & MonadBFT)
**Description**: Block propagates through 3-phase consensus

**Inputs**:
- `from_block_building`: Block with included transactions

**MonadBFT Phases**:
1. **Proposed** (0 blocks): Leader proposes block
2. **Voted** (after 1 block): Validators vote on block
3. **Finalized** (after 2 blocks): Block achieves finality

**Outputs**:
- `consensus_failed`: Block rejected by consensus
- `raptor_cast_latency`: Network propagation time
- `to_execution`: Finalized blocks ready for execution

**Prometheus Metrics**:
- `monad_bft_consensus_blocks_proposed`
- `monad_bft_consensus_blocks_voted`
- `monad_bft_consensus_blocks_finalized`
- `monad_bft_raptor_cast_latency_ns` (if available)

**NEW: MonadBFT Tracker**:
Track blocks through phases for visualization:
```go
type BlockConsensusState struct {
    BlockNumber  uint64
    BlockHash    string
    Phase        string  // "proposed", "voted", "finalized"
    ProposedAt   time.Time
    VotedAt      *time.Time
    FinalizedAt  *time.Time
}
```

---

### Stage 5: **Execution** (Parallel Processing)
**Description**: Transactions execute optimistically in parallel

**Inputs**:
- `from_consensus`: Finalized blocks

**Execution Modes**:
- **Parallel**: Optimistic concurrent execution
- **Serial Commit**: Results committed in order

**Outputs**:
- `parallel_success`: Successfully executed in parallel
- `parallel_retry`: Required serial retry due to conflicts
- `execution_failed`: Transaction reverted
- `to_state_update`: Successful executions

**Prometheus Metrics**:
- `monad_execution_ledger_num_tx_commits`
- `monad_execution_parallel_success` (if available)
- `monad_execution_serial_fallback` (if available)
- `monad_execution_reverted_txs` (if available)

---

### Stage 6: **State Update** (Commitment)
**Description**: State changes committed serially in original order

**Inputs**:
- `from_execution`: Executed transactions

**State Operations**:
- Account balance updates
- Storage writes
- Event log emissions
- State proof generation

**Outputs**:
- `accounts_updated`: Number of accounts modified
- `storage_writes`: Storage slots written
- `logs_emitted`: Events emitted
- `to_finality`: Transactions committed to state

**Prometheus Metrics**:
- `monad_state_accounts_updated` (if available)
- `monad_state_storage_writes` (if available)
- Estimated from tx_commits if not available

---

### Stage 7: **Finality** (2-Block Confirmation)
**Description**: Transaction results become queryable after 2 blocks

**Inputs**:
- `from_state_update`: Committed transactions

**Finality Timeline**:
- Block N: Transaction executed
- Block N+1: Voted
- Block N+2: **Finalized** (user can query via eth_getTransactionReceipt)

**Outputs**:
- `queryable`: Transactions available for user queries
- `receipts_generated`: Transaction receipts created

**Note**: This stage represents the 2-block delay before results are guaranteed final.

---

## Waterfall Data Structure

### JSON Format
```json
{
  "nodes": [
    {"id": "submission_rpc", "label": "RPC"},
    {"id": "submission_p2p", "label": "P2P"},
    {"id": "mempool", "label": "Mempool"},
    {"id": "block_building", "label": "Block Building"},
    {"id": "consensus_proposed", "label": "Proposed"},
    {"id": "consensus_voted", "label": "Voted"},
    {"id": "consensus_finalized", "label": "Finalized"},
    {"id": "execution", "label": "Execution"},
    {"id": "state_update", "label": "State Update"},
    {"id": "finality", "label": "Finalized (Queryable)"},
    {"id": "dropped", "label": "Dropped"}
  ],
  "links": [
    {"source": "submission_rpc", "target": "mempool", "value": 100},
    {"source": "submission_p2p", "target": "mempool", "value": 50},
    {"source": "mempool", "target": "dropped", "value": 10},
    {"source": "mempool", "target": "block_building", "value": 140},
    {"source": "block_building", "target": "dropped", "value": 5},
    {"source": "block_building", "target": "consensus_proposed", "value": 135},
    {"source": "consensus_proposed", "target": "consensus_voted", "value": 135},
    {"source": "consensus_voted", "target": "consensus_finalized", "value": 135},
    {"source": "consensus_finalized", "target": "execution", "value": 135},
    {"source": "execution", "target": "state_update", "value": 133},
    {"source": "execution", "target": "dropped", "value": 2},
    {"source": "state_update", "target": "finality", "value": 133}
  ],
  "metadata": {
    "source": "prometheus_metrics",
    "timestamp": 1234567890,
    "block_height": 12345,
    "tps": 45.2,
    "consensus_state": {
      "proposed_blocks": 5,
      "voted_blocks": 3,
      "finalized_blocks": 2
    }
  }
}
```

---

## MonadBFT Consensus Visualization

### Separate Component: `ConsensusStageTracker`

**Display Format**:
```
[Block N-2] ████████████ Finalized    (100%)
[Block N-1] ████████░░░░ Voted        (66%)
[Block N  ] ████░░░░░░░░ Proposed     (33%)
```

**Real-time Updates**:
- Track last 10 blocks
- Show progression: Proposed → Voted → Finalized
- Highlight current block phase
- Color coding: Proposed (blue), Voted (yellow), Finalized (green)

**Data Structure**:
```go
type ConsensusTracker struct {
    Blocks []BlockConsensusState
    CurrentBlock uint64
    FinalizedBlock uint64
}
```

---

## Implementation Plan

### Backend Changes

#### 1. New File: `consensus_tracker.go`
```go
// Track MonadBFT consensus phases for blocks
type ConsensusTracker struct {
    blocks map[uint64]*BlockConsensusState
    mu     sync.RWMutex
}

func (ct *ConsensusTracker) OnBlockProposed(blockNum uint64, hash string)
func (ct *ConsensusTracker) OnBlockVoted(blockNum uint64)
func (ct *ConsensusTracker) OnBlockFinalized(blockNum uint64)
func (ct *ConsensusTracker) GetRecentBlocks(count int) []BlockConsensusState
```

#### 2. Update: `waterfall_metrics.go`
- Replace old 7-stage structure with new Monad-aligned stages
- Add helper functions for each stage calculation
- Integrate ConsensusTracker for phase-aware metrics

#### 3. Update: `prometheus_collector.go`
- Add consensus metrics collection
- Track block phase transitions
- Calculate RaptorCast latencies if available

#### 4. Update: `firedancer_protocol.go`
- Add new message types for consensus phases
- Send MonadBFT state updates via WebSocket
- Include consensus_state in metadata

### Frontend Changes

#### 1. New Component: `MonadBFTTracker.tsx`
Location: `frontend/src/features/Overview/MonadBFT/`

**Features**:
- Display last 10 blocks with their consensus phases
- Progress bars for each block (33% / 66% / 100%)
- Real-time updates via WebSocket
- Click to expand block details

#### 2. Update: Sankey Diagram
- Modify node structure to include consensus sub-stages
- Add color coding for MonadBFT phases
- Update tooltips with phase information

#### 3. Update: `atoms.ts`
```typescript
export const consensusStateAtom = atom<ConsensusState>({
  blocks: [],
  currentBlock: 0,
  finalizedBlock: 0
});
```

---

## Prometheus Metrics Mapping

| Monad Stage | Primary Metric | Fallback |
|-------------|---------------|----------|
| Submission | `monad_bft_txpool_pool_insert_owned_txs` | Block data |
| Mempool | `monad_bft_txpool_pool_drop_nonce_too_low` | Estimation |
| Block Building | `monad_bft_txpool_pool_drop_insufficient_balance` | Estimation |
| Consensus | `monad_bft_consensus_blocks_*` | Block intervals |
| Execution | `monad_execution_ledger_num_tx_commits` | ✅ Available |
| State Update | (estimated from tx_commits) | 3:1 reads:writes |
| Finality | Block N+2 tracking | ✅ Calculable |

---

## Drop Reasons Taxonomy

### Submission Stage
- `invalid_signature`: Signature verification failed
- `invalid_format`: Malformed transaction data

### Mempool Stage
- `nonce_too_low`: Nonce below account nonce
- `nonce_gap`: Non-contiguous nonce
- `gas_limit_exceeded`: Gas > block gas limit
- `propagation_timeout`: Failed K leader attempts

### Block Building Stage
- `insufficient_balance`: Cannot pay gas fee
- `nonce_conflict`: Nonce already used
- `pool_full`: Block capacity reached

### Execution Stage
- `execution_reverted`: Smart contract reverted
- `out_of_gas`: Gas limit exhausted
- `invalid_opcode`: EVM execution error

---

## Testing Strategy

1. **Unit Tests**: Test each stage calculation independently
2. **Integration Tests**: Verify full waterfall flow
3. **Visual Validation**: Compare with actual node behavior
4. **Prometheus Validation**: Ensure metrics match reality

---

## Migration Notes

### Breaking Changes
- Old waterfall API structure changes
- Frontend expects new node/link format
- MonadBFT tracker requires new WebSocket messages

### Backward Compatibility
- Keep old `/api/v1/waterfall` endpoint for 1 version
- Add `/api/v1/waterfall/v2` with new structure
- Frontend detects and uses appropriate version

---

## Success Criteria

✅ Waterfall accurately reflects Monad documentation
✅ MonadBFT phases visible in real-time
✅ Drop reasons mapped to correct stages
✅ RaptorCast propagation tracked (if metrics available)
✅ 2-block finality delay visualized
✅ All stages use real Prometheus metrics where possible

---

## Future Enhancements

1. **Leader Propagation Details**: Show N×K retry visualization
2. **RaptorCast Network Map**: Display block propagation topology
3. **Historical Analysis**: Track waterfall metrics over time
4. **Alert System**: Notify on abnormal drop rates
5. **Comparison Mode**: Compare current vs. historical performance
