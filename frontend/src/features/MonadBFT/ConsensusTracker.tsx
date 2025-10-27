import { useAtomValue } from "jotai";
import { monadConsensusStateAtom } from "../../api/atoms";
import { Box, Card, Flex, Text, Badge, Progress } from "@radix-ui/themes";
import type { BlockConsensusState } from "../../api/types";

/**
 * MonadBFT Consensus Tracker
 *
 * Displays the consensus state of recent blocks showing progression through:
 * - Proposed (33% - Block N)
 * - Voted (66% - Block N+1)
 * - Finalized (100% - Block N+2)
 */
export function ConsensusTracker() {
  const consensusState = useAtomValue(monadConsensusStateAtom);

  if (!consensusState) {
    return (
      <Card>
        <Text size="2" color="gray">
          Waiting for consensus data...
        </Text>
      </Card>
    );
  }

  const { current_block, finalized_block, blocks_behind, recent_blocks } =
    consensusState;

  // Handle optional fields
  const blocksBehind = blocks_behind ?? 0;
  const currentBlock = current_block ?? 0;
  const finalizedBlock = finalized_block ?? 0;
  const recentBlocks = recent_blocks ?? [];

  return (
    <Card>
      <Flex direction="column" gap="3">
        {/* Header */}
        <Flex justify="between" align="center">
          <Text size="4" weight="bold">
            MonadBFT Consensus
          </Text>
          <Badge color={blocksBehind <= 2 ? "green" : "yellow"} size="2">
            {blocksBehind} blocks behind
          </Badge>
        </Flex>

        {/* Summary */}
        <Flex gap="4">
          <Flex direction="column" gap="1">
            <Text size="1" color="gray">
              Current Block
            </Text>
            <Text size="3" weight="bold">
              {currentBlock.toLocaleString()}
            </Text>
          </Flex>
          <Flex direction="column" gap="1">
            <Text size="1" color="gray">
              Finalized Block
            </Text>
            <Text size="3" weight="bold">
              {finalizedBlock.toLocaleString()}
            </Text>
          </Flex>
        </Flex>

        {/* Recent Blocks */}
        <Flex direction="column" gap="2">
          <Text size="2" weight="medium">
            Recent Blocks
          </Text>
          {recentBlocks.slice(0, 10).map((block) => (
            <BlockProgress key={block.block_number} block={block} />
          ))}
        </Flex>
      </Flex>
    </Card>
  );
}

interface BlockProgressProps {
  block: BlockConsensusState;
}

function BlockProgress({ block }: BlockProgressProps) {
  const { block_number, phase, tx_count } = block;

  // Handle optional fields with defaults
  const blockNumber = block_number ?? 0;
  const blockPhase = phase ?? "proposed";
  const txCount = tx_count ?? 0;

  // Calculate progress percentage
  const progressValue = blockPhase === "proposed" ? 33 : blockPhase === "voted" ? 66 : 100;

  // Color based on phase
  const color =
    blockPhase === "proposed"
      ? "blue"
      : blockPhase === "voted"
      ? "yellow"
      : "green";

  return (
    <Box>
      <Flex justify="between" align="center" mb="1">
        <Flex gap="2" align="center">
          <Text size="1" weight="medium">
            Block #{blockNumber}
          </Text>
          <Badge color={color} size="1">
            {blockPhase.charAt(0).toUpperCase() + blockPhase.slice(1)}
          </Badge>
        </Flex>
        <Text size="1" color="gray">
          {txCount} txs
        </Text>
      </Flex>
      <Progress value={progressValue} color={color} />
    </Box>
  );
}

export default ConsensusTracker;
