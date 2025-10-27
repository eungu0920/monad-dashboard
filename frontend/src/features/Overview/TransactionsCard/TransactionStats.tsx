import { Flex } from "@radix-ui/themes";
import CardStat from "../../../components/CardStat";
import { useAtomValue } from "jotai";
import { estimatedTpsAtom } from "../../../api/atoms";
import { headerColor } from "../../../colors";

export default function TransactionStats() {
  const estimated = useAtomValue(estimatedTpsAtom);
  return (
    <Flex direction="column" gap="2" minWidth="180px" style={{ width: "180px" }}>
      <CardStat
        label="Total TPS (1s)"
        value={
          estimated?.total.toLocaleString(undefined, {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
          }) ?? "-"
        }
        valueColor={headerColor}
        large
      />
    </Flex>
  );
}
