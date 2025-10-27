# MonadBFT Consensus Tracker

Visual component for tracking MonadBFT consensus phases of recent blocks.

## Features

- **3-Phase Consensus Display**: Shows Proposed → Voted → Finalized progression
- **Progress Bars**: Visual representation of consensus advancement (33% / 66% / 100%)
- **Real-time Updates**: Automatically updates as new blocks arrive via WebSocket
- **Finality Lag Indicator**: Shows how many blocks behind the current block is finalized

## Usage

```tsx
import { ConsensusTracker } from "../../features/MonadBFT";

function YourComponent() {
  return <ConsensusTracker />;
}
```

## Data Source

Subscribes to `monadConsensusStateAtom` which receives `monad_consensus_state` WebSocket messages from the backend.

### Message Format

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
    proposed_at: string;
    voted_at?: string;
    finalized_at?: string;
    tx_count: number;
  }>;
}
```

## Integration

### Add to Overview Page

```tsx
// In features/Overview/index.tsx
import { ConsensusTracker } from "../MonadBFT";

export default function Overview() {
  return (
    <>
      {/* ... existing components */}
      <ConsensusTracker />
    </>
  );
}
```

### Add to Sidebar

Or create a dedicated consensus monitoring page/section.

## Styling

Uses Radix UI components:
- `Card`: Container
- `Progress`: Progress bars with color-coded phases
- `Badge`: Phase labels and finality lag indicator
- `Flex`, `Box`, `Text`: Layout and typography

## Phase Colors

- **Proposed** (33%): Blue - Initial block proposal
- **Voted** (66%): Yellow - Validators have voted on the block
- **Finalized** (100%): Green - Block is finalized and immutable
