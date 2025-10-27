import { Flex, Box } from "@radix-ui/themes";
import CardHeader from "../../../components/CardHeader";
import Card from "../../../components/Card";
import CardStat from "../../../components/CardStat";
import { useAtomValue } from "jotai";
import styles from "./statusCard.module.css";
import { currentSlotAtom } from "../../../atoms";
import { headerColor } from "../../../colors";

export default function SlotStatusCard() {
  return (
    <Card style={{ flexGrow: 1 }}>
      <Flex direction="column" height="100%" gap="2" align="start">
        <CardHeader text="Status" />
        <div className={styles.statRow}>
          <CurrentSlotText />
        </div>
      </Flex>
    </Card>
  );
}

function CurrentSlotText() {
  const currentSlot = useAtomValue(currentSlotAtom);

  return (
    <Box>
      <CardStat
        label="Block"
        value={currentSlot?.toString() ?? ""}
        valueColor={headerColor}
        large
      />
    </Box>
  );
}
