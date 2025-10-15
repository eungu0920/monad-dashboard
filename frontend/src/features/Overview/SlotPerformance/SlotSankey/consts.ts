export const enum SlotNode {
  // Monad: Not used but kept for compatibility
  IncPackCranked = "Crank:inc",
  IncPackRetained = "Buffered:inc",
  IncResolvRetained = "Unresolved:inc",

  // Monad: Network ingress
  IncQuic = "RPC",           // Changed: QUIC → RPC
  IncUdp = "P2P",            // Changed: UDP → P2P
  IncGossip = "P2P",         // Changed: Gossip → P2P (same as UDP)
  IncBlockEngine = "Relayer", // Changed: Jito → Relayer

  SlotStart = "Received",
  SlotEnd = "Packed",
  End = "End",

  // Monad: Pipeline stages
  Networking = "networking:tile",
  QUIC = "RPC:tile",         // Changed: QUIC → RPC
  Verification = "verify:tile",
  Dedup = "pool:tile",       // Changed: dedup → pool
  Resolv = "validate:tile",  // Changed: resolv → validate
  Pack = "pack:tile",
  Bank = "EVM:tile",         // Changed: bank → EVM

  // Monad: Network drops
  NetOverrun = "Too slow:net",
  QUICOverrun = "Too slow:rpc",     // Changed
  QUICInvalid = "Malformed:rpc",    // Changed
  QUICTooManyFrags = "Out of buffers:rpc", // Changed
  QUICAbandoned = "Abandoned:rpc",  // Changed

  // Monad: Verification drops
  VerifyOverrun = "Too slow:verify",
  VerifyParse = "Unparseable",
  VerifyFailed = "Sig Failed",      // Changed: Bad signature → Sig Failed
  VerifyDuplicate = "Nonce Failed", // Changed: Duplicate → Nonce Failed

  // Monad: Pool/Validation drops
  DedupDeuplicate = "Nonce Failed", // Changed: Duplicate → Nonce Failed
  ResolvFailed = "Balance Failed",  // Changed: Bad LUT → Balance Failed
  ResolvExpired = "Fee Too Low",    // Changed: Expired → Fee Too Low
  ResolvNoLedger = "No ledger",
  ResolvRetained = "Unresolved",    // Changed

  // Monad: Packing drops
  PackInvalid = "Invalid Tx",       // Changed: Unpackable → Invalid Tx
  PackInvalidBundle = "Bad Bundle",
  PackExpired = "Expired:pack",
  PackRetained = "Buffered:pack",
  PackLeaderSlow = "Buffer full",
  PackWaitFull = "Pool Full",       // Changed: Storage full → Pool Full

  // Monad: EVM execution
  BankInvalid = "Exec Failed",      // Changed: Unexecutable → Exec Failed

  // Monad: Final outcomes
  BlockSuccess = "Parallel",        // Changed: Success → Parallel
  BlockFailure = "Sequential",      // Changed: Failure → Sequential

  Votes = "Votes",
  NonVoteSuccess = "Non-vote Success",
  NonVoteFailure = "Non-vote Failure",
}

export const startEndNodes: SlotNode[] = [SlotNode.SlotStart, SlotNode.SlotEnd];

export const tileNodes: SlotNode[] = [
  SlotNode.Networking,
  SlotNode.QUIC,
  SlotNode.Verification,
  SlotNode.Dedup,
  SlotNode.Resolv,
  SlotNode.Pack,
  SlotNode.Bank,
];

export const droppedSlotNodes: SlotNode[] = [
  SlotNode.NetOverrun,
  SlotNode.QUICOverrun,
  SlotNode.QUICInvalid,
  SlotNode.QUICTooManyFrags,
  SlotNode.QUICAbandoned,
  SlotNode.VerifyOverrun,
  SlotNode.VerifyParse,
  SlotNode.VerifyFailed,
  SlotNode.VerifyDuplicate,
  SlotNode.DedupDeuplicate,
  SlotNode.ResolvFailed,
  SlotNode.ResolvExpired,
  SlotNode.ResolvNoLedger,
  SlotNode.ResolvRetained,
  SlotNode.PackInvalid,
  SlotNode.PackInvalidBundle,
  SlotNode.PackExpired,
  SlotNode.PackRetained,
  SlotNode.PackLeaderSlow,
  SlotNode.PackWaitFull,
  SlotNode.BankInvalid,
];

export const incomingSlotNodes: SlotNode[] = [
  SlotNode.IncQuic,
  SlotNode.IncUdp,
];

export const retainedSlotNodes: SlotNode[] = [
  SlotNode.IncPackRetained,
  SlotNode.PackRetained,
  SlotNode.IncResolvRetained,
  SlotNode.ResolvRetained,
];

export const successfulSlotNodes: SlotNode[] = [
  SlotNode.BlockSuccess,
  SlotNode.NonVoteSuccess,
];

export const failedSlotNodes: SlotNode[] = [
  SlotNode.BlockFailure,
  SlotNode.NonVoteFailure,
];

export const slotNodes = [
  {
    id: SlotNode.IncQuic,
  },
  {
    id: SlotNode.IncUdp,
  },
  { id: SlotNode.PackRetained, labelPositionOverride: "right" },
  { id: SlotNode.ResolvRetained, labelPositionOverride: "right" },
  { id: SlotNode.NetOverrun, labelPositionOverride: "right" },
  { id: SlotNode.QUICOverrun, labelPositionOverride: "right" },
  { id: SlotNode.QUICInvalid, labelPositionOverride: "right" },
  { id: SlotNode.QUICTooManyFrags, labelPositionOverride: "right" },
  { id: SlotNode.QUICAbandoned, labelPositionOverride: "right" },
  { id: SlotNode.VerifyOverrun, labelPositionOverride: "right" },
  { id: SlotNode.VerifyParse, labelPositionOverride: "right" },
  { id: SlotNode.VerifyFailed, labelPositionOverride: "right" },
  { id: SlotNode.VerifyDuplicate, labelPositionOverride: "right" },
  { id: SlotNode.DedupDeuplicate, labelPositionOverride: "right" },
  { id: SlotNode.ResolvFailed, labelPositionOverride: "right" },
  { id: SlotNode.ResolvExpired, labelPositionOverride: "right" },
  { id: SlotNode.ResolvNoLedger, labelPositionOverride: "right" },
  { id: SlotNode.PackInvalid, labelPositionOverride: "right" },
  { id: SlotNode.PackInvalidBundle, labelPositionOverride: "right" },
  { id: SlotNode.PackExpired, labelPositionOverride: "right" },
  { id: SlotNode.PackLeaderSlow, labelPositionOverride: "right" },
  { id: SlotNode.PackWaitFull, labelPositionOverride: "right" },
  { id: SlotNode.BankInvalid, labelPositionOverride: "right" },
  {
    id: SlotNode.SlotStart,
    alignLabelBottom: true,
    labelPositionOverride: "right",
  },
  {
    id: SlotNode.QUIC,
    alignLabelBottom: true,
  },
  {
    id: SlotNode.Verification,
    alignLabelBottom: true,
  },
  {
    id: SlotNode.Dedup,
    alignLabelBottom: true,
  },
  {
    id: SlotNode.Resolv,
    alignLabelBottom: true,
  },
  {
    id: SlotNode.IncGossip,
  },
  {
    id: SlotNode.IncBlockEngine,
  },
  {
    id: SlotNode.IncResolvRetained,
    labelPositionOverride: "left",
  },
  {
    id: SlotNode.IncPackCranked,
    labelPositionOverride: "left",
  },
  {
    id: SlotNode.IncPackRetained,
    labelPositionOverride: "left",
  },
  {
    id: SlotNode.Pack,
    alignLabelBottom: true,
  },
  {
    id: SlotNode.Bank,
    alignLabelBottom: true,
  },
  {
    id: SlotNode.End,
    hideLabel: true,
  },
  {
    id: SlotNode.SlotEnd,
    alignLabelBottom: true,
    labelPositionOverride: "left",
  },
  {
    id: SlotNode.BlockFailure,
  },
  {
    id: SlotNode.BlockSuccess,
  },
  {
    id: SlotNode.Votes,
  },
  {
    id: SlotNode.NonVoteFailure,
  },
  {
    id: SlotNode.NonVoteSuccess,
  },
];
