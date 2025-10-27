import { Card, Flex, Text, Tabs, Box } from "@radix-ui/themes";
import { useAtomValue } from "jotai";
import { transactionLogsAtom } from "./atoms";
import TransactionFlowList from "./TransactionFlowList";
import TransactionFlowTimeline from "./TransactionFlowTimeline";

export default function TransactionFlowCard() {
  const logs = useAtomValue(transactionLogsAtom);

  return (
    <Card size="2" style={{ gridColumn: "1 / -1" }}>
      <Flex direction="column" gap="3">
        <Flex justify="between" align="center">
          <Text size="5" weight="bold">
            Transaction Flow
          </Text>
          <Text size="2" color="gray">
            {logs.length} transactions
          </Text>
        </Flex>

        <Tabs.Root defaultValue="timeline">
          <Tabs.List>
            <Tabs.Trigger value="timeline">Timeline</Tabs.Trigger>
            <Tabs.Trigger value="list">List</Tabs.Trigger>
          </Tabs.List>

          <Box pt="3">
            <Tabs.Content value="timeline">
              <TransactionFlowTimeline />
            </Tabs.Content>

            <Tabs.Content value="list">
              <TransactionFlowList />
            </Tabs.Content>
          </Box>
        </Tabs.Root>
      </Flex>
    </Card>
  );
}
