# Frontend Integration Guide - Monad Waterfall V2

## What's New

### 1. **New WebSocket Messages**

The backend now sends two new message types:

#### `monad_waterfall_v2`
Replaces the old `live_txn_waterfall` format with a Sankey diagram-friendly structure:

```typescript
{
  nodes: Array<{ id: string; label: string; color?: string }>;
  links: Array<{ source: string; target: string; value: number }>;
  metadata: {
    source: "prometheus_metrics" | "block_estimation" | "mock_data";
    tps?: number;
    consensus_state?: ConsensusStateMetadata;
    // ... other metadata
  };
  drops?: {
    invalid_signature?: number;
    nonce_invalid?: number;
    // ... other drop reasons
  };
}
```

#### `monad_consensus_state`
Real-time MonadBFT consensus phase tracking:

```typescript
{
  current_block: number;
  finalized_block: number;
  blocks_behind: number;
  proposed_blocks: number;
  voted_blocks: number;
  finalized_blocks: number;
  recent_blocks: Array<{
    block_number: number;
    block_hash: string;
    phase: "proposed" | "voted" | "finalized";
    tx_count: number;
    // ... timestamps
  }>;
}
```

### 2. **New Atoms**

```typescript
// frontend/src/api/atoms.ts
export const monadWaterfallV2Atom: MonadWaterfallV2 | undefined;
export const monadConsensusStateAtom: MonadConsensusState | undefined;
```

### 3. **New Component**

```typescript
// frontend/src/features/MonadBFT/ConsensusTracker.tsx
<ConsensusTracker />
```

Displays recent blocks with their consensus phases and progress bars.

---

## Integration Steps

### Step 1: WebSocket Message Handling ✅

**Status**: Already implemented in `useSetAtomWsData.ts`

The following cases are already added:

```typescript
case "monad_waterfall_v2": {
  setDbMonadWaterfallV2(value);
  break;
}
case "monad_consensus_state": {
  setMonadConsensusState(value);
  break;
}
```

---

### Step 2: Add ConsensusTracker to UI

**Location**: Choose where to display the consensus tracker

#### Option A: Add to Overview Page

```tsx
// frontend/src/features/Overview/index.tsx
import { ConsensusTracker } from "../MonadBFT";

export default function Overview() {
  return (
    <>
      {/* Existing components */}
      <SlotPerformance />

      {/* NEW: Add consensus tracker */}
      <ConsensusTracker />
    </>
  );
}
```

#### Option B: Add to Sidebar/Dashboard

Create a dedicated monitoring section for consensus state.

---

### Step 3: Update Sankey Diagram (Optional)

**Current State**: The Sankey diagram uses the old `liveTxnWaterfallAtom` format.

**To Update**: Modify the Sankey diagram component to use `monadWaterfallV2Atom`:

```tsx
// Example: frontend/src/features/Overview/SlotPerformance/index.tsx
import { monadWaterfallV2Atom } from "../../../api/atoms";

function YourSankeyComponent() {
  const waterfallV2 = useAtomValue(monadWaterfallV2Atom);

  if (!waterfallV2) return <Loading />;

  // waterfallV2.nodes and waterfallV2.links are ready for Sankey!
  return <ResponsiveSankey data={{ nodes: waterfallV2.nodes, links: waterfallV2.links }} />;
}
```

**Benefit**: The new format aligns with Monad's 7-stage transaction lifecycle:
1. Submission
2. Mempool
3. Block Building
4. Consensus (Proposed/Voted/Finalized)
5. Execution
6. State Update
7. Finality

---

### Step 4: Test with Live Data

After deploying to the server:

1. **Open Browser DevTools** → Network → WS tab
2. **Filter messages** for `monad_waterfall_v2` and `monad_consensus_state`
3. **Verify data**: Check that `metadata.source` is `"prometheus_metrics"`
4. **Check consensus tracker**: Should show recent blocks with phases

Example WebSocket message:

```json
{
  "topic": "summary",
  "key": "monad_consensus_state",
  "value": {
    "current_block": 12345,
    "finalized_block": 12343,
    "blocks_behind": 2,
    "recent_blocks": [
      {
        "block_number": 12345,
        "phase": "proposed",
        "tx_count": 150
      }
    ]
  }
}
```

---

## File Changes Summary

### New Files
1. `frontend/src/features/MonadBFT/ConsensusTracker.tsx` - Consensus visualization component
2. `frontend/src/features/MonadBFT/index.ts` - Exports
3. `frontend/src/features/MonadBFT/README.md` - Component documentation

### Modified Files
1. `frontend/src/api/entities.ts` - Added Zod schemas for new message types
2. `frontend/src/api/types.ts` - Added TypeScript types
3. `frontend/src/api/atoms.ts` - Added new atoms
4. `frontend/src/api/useSetAtomWsData.ts` - Added message handlers

---

## Backward Compatibility

**Important**: The backend still sends `live_txn_waterfall` (legacy format) for backward compatibility.

You can:
- **Use both**: Keep old waterfall, add consensus tracker separately
- **Migrate gradually**: Switch one component at a time to use v2 data
- **Remove legacy**: After confirming v2 works, deprecate old format

---

## Troubleshooting

### ConsensusTracker shows "Waiting for data..."

**Check**:
1. WebSocket connection is established
2. Backend is sending `monad_consensus_state` messages
3. Atom is being updated in DevTools (React DevTools → Jotai atoms)

```typescript
// Debug in console
import { monadConsensusStateAtom } from './api/atoms';
console.log(useAtomValue(monadConsensusStateAtom));
```

### Waterfall V2 not appearing

**Check**:
1. Backend sends `monad_waterfall_v2` messages every 200ms
2. `monadWaterfallV2Atom` is populated
3. Zod validation passes (check console for errors)

```typescript
// Debug
import { monadWaterfallV2Atom } from './api/atoms';
const data = useAtomValue(monadWaterfallV2Atom);
console.log('Waterfall V2:', data);
console.log('Source:', data?.metadata?.source);
```

### TypeScript Errors

If you see type errors:
1. Ensure all imports are from `./api/types` or `./api/atoms`
2. Check that `MonadWaterfallV2` and `MonadConsensusState` types are exported
3. Run `npm run typecheck` (or `tsc --noEmit`)

---

## Next Steps

1. ✅ Deploy backend changes to server
2. ⏳ Add `<ConsensusTracker />` to a page
3. ⏳ Test with real Monad node data
4. ⏳ (Optional) Update Sankey to use `monadWaterfallV2Atom`
5. ⏳ (Optional) Add historical consensus metrics chart

---

## Performance Notes

- **Throttling**: `monadWaterfallV2Atom` uses `waterfallDebounceMs` throttling
- **Updates**: WebSocket messages arrive every 200ms
- **Memory**: Consensus tracker shows last 10 blocks (~2-3 KB)

---

## References

- Backend: `backend/WATERFALL_REDESIGN.md`
- Deployment: `backend/DEPLOYMENT.md`
- Component: `frontend/src/features/MonadBFT/README.md`
