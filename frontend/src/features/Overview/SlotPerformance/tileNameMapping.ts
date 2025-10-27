/**
 * Maps Firedancer tile types to Monad-equivalent processing stages
 *
 * Monad Transaction Lifecycle:
 * 1. Submission (RPC/P2P)
 * 2. Mempool Validation
 * 3. Leader Propagation
 * 4. Block Building
 * 5. Consensus (MonadBFT)
 * 6. Execution (Parallel)
 * 7. State Update
 */

export const TILE_NAME_MAPPING: Record<string, string> = {
  // Network I/O
  sock: "Network", // sock → Network I/O
  net: "Network", // net → Network

  // RPC/Connection handling
  quic: "RPC", // QUIC → RPC connections

  // Transaction processing
  bundle: "Bundle", // Keep bundle (if using)
  verify: "Verify", // Signature verification
  dedup: "Mempool", // dedup → Mempool validation
  resolv: "Validate", // resolv (address lookup) → Validation

  // Block production
  pack: "Block Build", // pack → Block Building
  poh: "Consensus", // PoH → MonadBFT Consensus

  // Execution & finalization
  bank: "Execution", // bank → EVM Execution
  shred: "Propagate", // shred → Block Propagation
  store: "State Update", // store → State Update

  // Other tiles
  sign: "Sign",
  metric: "Metrics",
  http: "HTTP",
  plugin: "Plugin",
  gui: "GUI",
  cswtch: "Context Switch",
};

/**
 * Get Monad-friendly display name for a tile
 */
export function getMonadTileName(tileKind: string): string {
  return TILE_NAME_MAPPING[tileKind] || tileKind;
}
