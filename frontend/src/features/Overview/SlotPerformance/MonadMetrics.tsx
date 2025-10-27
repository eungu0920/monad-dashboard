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

  const metadata = waterfallV2.metadata as any;
  const drops = waterfallV2.drops as any;

  // Calculate ingress metrics
  const rpcSubmit = Number(metadata.rpc_submit) || 0;
  const p2pGossip = Number(metadata.p2p_gossip) || 0;
  const totalIngress = rpcSubmit + p2pGossip;

  return (
    <Flex gap="2" wrap="wrap" style={{ marginTop: "16px" }}>
      {/* 1. Transaction Ingress */}
      {(rpcSubmit > 0 || p2pGossip > 0) && (
        <MetricCard
          title="Tx Ingress"
          metrics={[
            { label: "RPC Submit", value: rpcSubmit },
            { label: "P2P Gossip", value: p2pGossip },
            { label: "Total", value: totalIngress },
          ]}
        />
      )}

      {/* 2. Mempool */}
      <MetricCard
        title="Mempool"
        metrics={[
          { label: "Pending", value: Number(metadata.pending_txs) || 0 },
          { label: "Tracked", value: Number(metadata.tracked_txs) || 0 },
        ]}
      />

      {/* 3. Consensus (MonadBFT) */}
      {consensus && (
        <MetricCard
          title="Consensus"
          metrics={[
            { label: "Proposed", value: Number((consensus as any).proposed_blocks) || 0 },
            { label: "Voted", value: Number((consensus as any).voted_blocks) || 0 },
            { label: "Finalized", value: Number((consensus as any).finalized_blocks) || 0 },
          ]}
        />
      )}

      {/* 4. Execution */}
      <MetricCard
        title="Execution"
        metrics={[
          { label: "TPS", value: metadata.tps ? Number(metadata.tps).toFixed(2) : "0.00" },
          { label: "Blocks", value: Number(metadata.blocks_committed || metadata.block_height) || 0 },
        ]}
      />

      {/* 5. Transaction Drops */}
      {drops && (
        <MetricCard
          title="Tx Drops"
          metrics={[
            { label: "Signature", value: Number(drops.invalid_signature) || 0 },
            { label: "Nonce", value: Number(drops.nonce_invalid) || 0 },
            { label: "Fee Too Low", value: Number(drops.fee_too_low) || 0 },
            { label: "Block Full", value: Number(drops.block_full || drops.pool_full) || 0 },
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
