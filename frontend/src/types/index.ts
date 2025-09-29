export interface MonadMetrics {
  timestamp: number;
  node_info: NodeInfo;
  waterfall: WaterfallMetrics;
  consensus: ConsensusMetrics;
  execution: ExecutionMetrics;
  network: NetworkMetrics;
}

export interface NodeInfo {
  version: string;
  chain_id: number;
  node_name: string;
  status: string;
  uptime: number;
}

export interface WaterfallMetrics {
  // Ingress
  rpc_received: number;
  gossip_received: number;
  mempool_size: number;

  // Validation drops
  signature_failed: number;
  nonce_duplicate: number;
  gas_invalid: number;
  balance_insufficient: number;

  // Execution
  evm_parallel_executed: number;
  evm_sequential_fallback: number;
  gas_used_total: number;
  state_conflicts: number;

  // Consensus
  bft_proposed: number;
  bft_voted: number;
  bft_committed: number;

  // Persistence
  state_updated: number;
  triedb_written: number;
  blocks_broadcast: number;
}

export interface ConsensusMetrics {
  current_height: number;
  last_block_time: number;
  block_time: number;
  validator_count: number;
  voting_power: number;
  participation_rate: number;
}

export interface ExecutionMetrics {
  tps: number;
  pending_tx_count: number;
  parallel_success_rate: number;
  avg_gas_price: number;
  avg_execution_time: number;
  state_size: number;
}

export interface NetworkMetrics {
  peer_count: number;
  inbound_peers: number;
  outbound_peers: number;
  bytes_in: number;
  bytes_out: number;
  network_latency: number;
}

export interface WaterfallStage {
  name: string;
  in: number;
  out: number;
  drop: number;
  success: number;
  parallel_rate?: number;
}

export interface WaterfallData {
  timestamp: number;
  stages: WaterfallStage[];
  summary: {
    total_in: number;
    total_success: number;
    total_dropped: number;
    success_rate: number;
  };
}