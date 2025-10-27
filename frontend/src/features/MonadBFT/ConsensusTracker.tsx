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

  return (
    <Card>
      <Flex direction="column" gap="3">
        {/* Header */}
        <Flex justify="between" align="center">
          <Text size="4" weight="bold">
            MonadBFT Consensus
          </Text>
          <Badge color={blocks_behind <= 2 ? "green" : "yellow"} size="2">
            {blocks_behind} blocks behind
          </Badge>
        </Flex>

        {/* Summary */}
        <Flex gap="4">
          <Flex direction="column" gap="1">
            <Text size="1" color="gray">
              Current Block
            </Text>
            <Text size="3" weight="bold">
              {current_block.toLocaleString()}
            </Text>
          </Flex>
          <Flex direction="column" gap="1">
            <Text size="1" color="gray">
              Finalized Block
            </Text>
            <Text size="3" weight="bold">
              {finalized_block.toLocaleString()}
            </Text>
          </Flex>
        </Flex>

        {/* Recent Blocks */}
        <Flex direction="column" gap="2">
          <Text size="2" weight="medium">
            Recent Blocks
          </Text>
          {recent_blocks.slice(0, 10).map((block) => (
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

  // Calculate progress percentage
  const progressValue = phase === "proposed" ? 33 : phase === "voted" ? 66 : 100;

  // Color based on phase
  const color =
    phase === "proposed"
      ? "blue"
      : phase === "voted"
      ? "yellow"
      : "green";

  return (
    <Box>
      <Flex justify="between" align="center" mb="1">
        <Flex gap="2" align="center">
          <Text size="1" weight="medium">
            Block #{block_number}
          </Text>
          <Badge color={color} size="1">
            {phase.charAt(0).toUpperCase() + phase.slice(1)}
          </Badge>
        </Flex>
        <Text size="1" color="gray">
          {tx_count} txs
        </Text>
      </Flex>
      <Progress value={progressValue} color={color} />
    </Box>
  );
}

export default ConsensusTracker;
