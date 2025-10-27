import { Flex, Text, Box } from "@radix-ui/themes";
import { useAtomValue } from "jotai";
import { transactionLogsAtom, txFlowFilterAtom } from "./atoms";
import { TX_STATUS_COLORS } from "./consts";
import { useMemo } from "react";

export default function TransactionFlowList() {
  const logs = useAtomValue(transactionLogsAtom);
  const filter = useAtomValue(txFlowFilterAtom);

  const filteredLogs = useMemo(() => {
    return logs.filter((log) => {
      if (log.status === "success" && !filter.showSuccess) return false;
      if (log.status === "failed" && !filter.showFailed) return false;
      return true;
    });
  }, [logs, filter]);

  const displayLogs = filteredLogs.slice(0, 50); // Show max 50 for performance

  if (displayLogs.length === 0) {
    return (
      <Flex
        justify="center"
        align="center"
        style={{ height: "200px", opacity: 0.5 }}
      >
        <Text size="2" color="gray">
          No transaction logs yet. Waiting for transactions...
        </Text>
      </Flex>
    );
  }

  return (
    <Flex direction="column" gap="1" style={{ maxHeight: "400px", overflow: "auto" }}>
      {displayLogs.map((log, index) => {
        const statusColor = TX_STATUS_COLORS[log.status || "success"];
        const shortHash = `${log.transactionHash.slice(0, 6)}...${log.transactionHash.slice(-4)}`;

        return (
          <Flex
            key={`${log.transactionHash}-${index}`}
            align="center"
            gap="3"
            p="2"
            style={{
              borderLeft: `3px solid ${statusColor}`,
              backgroundColor: "var(--gray-a2)",
              borderRadius: "4px",
            }}
          >
            <Box
              style={{
                width: "8px",
                height: "8px",
                borderRadius: "50%",
                backgroundColor: statusColor,
                flexShrink: 0,
              }}
            />
            <Flex direction="column" gap="1" style={{ flex: 1, minWidth: 0 }}>
              <Flex align="center" gap="2">
                <Text size="2" weight="medium" style={{ fontFamily: "monospace" }}>
                  {shortHash}
                </Text>
                <Text size="1" color="gray">
                  Block #{log.blockNumber}
                </Text>
                <Text size="1" color="gray">
                  Index {log.transactionIndex}
                </Text>
              </Flex>
              {log.address && (
                <Text size="1" color="gray" style={{ fontFamily: "monospace" }}>
                  {log.address.slice(0, 42)}...
                </Text>
              )}
            </Flex>
            <Text size="1" color="gray" style={{ flexShrink: 0 }}>
              {new Date(log.timestamp * 1000).toLocaleTimeString()}
            </Text>
          </Flex>
        );
      })}
    </Flex>
  );
}
