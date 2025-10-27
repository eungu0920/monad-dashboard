import { Box, Card, Flex, Text } from "@radix-ui/themes";
import { useAtomValue } from "jotai";
import { monadWaterfallV2Atom, monadConsensusStateAtom } from "../../../api/atoms";
import styles from "./tileCard.module.css";

/**
 * MonadMetrics - Display actual Monad metrics from Prometheus
 *
 * Shows only metrics that are actually available from Monad:
 * 1. Transaction Ingress (RPC + P2P)
 * 2. Mempool Status
 * 3. Consensus State (MonadBFT)
 * 4. Execution Performance
 * 5. Transaction Drops
 */
export default function MonadMetrics() {
  const waterfallV2 = useAtomValue(monadWaterfallV2Atom);
  const consensus = useAtomValue(monadConsensusStateAtom);

  if (!waterfallV2?.metadata) {
    return null;
  }

  const metadata = waterfallV2.metadata;

  return (
    <Flex gap="2" wrap="wrap" style={{ marginTop: "16px" }}>
      {/* 1. Transaction Ingress */}
      <MetricCard
        title="Tx Ingress"
        metrics={[
          { label: "RPC Submit", value: metadata.rpc_submit || 0 },
          { label: "P2P Gossip", value: metadata.p2p_gossip || 0 },
          { label: "Total", value: (metadata.rpc_submit || 0) + (metadata.p2p_gossip || 0) },
        ]}
      />

      {/* 2. Mempool */}
      <MetricCard
        title="Mempool"
        metrics={[
          { label: "Pending", value: metadata.pending_txs || 0 },
          { label: "Tracked", value: metadata.tracked_txs || 0 },
        ]}
      />

      {/* 3. Consensus (MonadBFT) */}
      {consensus && (
        <MetricCard
          title="Consensus"
          metrics={[
            { label: "Proposed", value: consensus.proposed_blocks || 0 },
            { label: "Voted", value: consensus.voted_blocks || 0 },
            { label: "Finalized", value: consensus.finalized_blocks || 0 },
          ]}
        />
      )}

      {/* 4. Execution */}
      <MetricCard
        title="Execution"
        metrics={[
          { label: "TPS", value: metadata.tps?.toFixed(2) || "0.00" },
          { label: "Blocks", value: metadata.blocks_committed || 0 },
        ]}
      />

      {/* 5. Transaction Drops */}
      {metadata.drops && (
        <MetricCard
          title="Tx Drops"
          metrics={[
            { label: "Signature", value: metadata.drops.invalid_signature || 0 },
            { label: "Nonce", value: metadata.drops.nonce_invalid || 0 },
            { label: "Fee Too Low", value: metadata.drops.fee_too_low || 0 },
            { label: "Pool Full", value: metadata.drops.pool_full || 0 },
          ]}
        />
      )}
    </Flex>
  );
}

interface MetricCardProps {
  title: string;
  metrics: Array<{
    label: string;
    value: number | string;
  }>;
}

function MetricCard({ title, metrics }: MetricCardProps) {
  return (
    <Card style={{ minWidth: "180px", flex: "1 1 180px" }}>
      <Flex direction="column" gap="2">
        <Text size="3" weight="bold" className={styles.header}>
          {title}
        </Text>
        {metrics.map((metric, i) => (
          <Flex key={i} justify="between" align="center">
            <Text size="1" color="gray">
              {metric.label}
            </Text>
            <Text size="2" weight="medium">
              {typeof metric.value === "number" && metric.value > 1000
                ? metric.value.toLocaleString()
                : metric.value}
            </Text>
          </Flex>
        ))}
      </Flex>
    </Card>
  );
}
