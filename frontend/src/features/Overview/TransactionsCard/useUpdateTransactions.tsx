import { useAtomValue, useSetAtom } from "jotai";
import { tpsHistoryAtom } from "../../../api/atoms";
import { useEffect } from "react";
import { tpsDataAtom } from "./atoms";
import type { EstimatedTps } from "../../../api/types";
import { maxTransactionChartPoints } from "./consts";

export default function useUpdateTransactions() {
  const tpsHistory = useAtomValue(tpsHistoryAtom);
  const setTpsData = useSetAtom(tpsDataAtom);

  // Update chart whenever tpsHistory changes (per-block updates from backend)
  useEffect(() => {
    if (!tpsHistory) return;

    const empty: undefined[] = new Array<undefined>(
      maxTransactionChartPoints,
    ).fill(undefined);

    // Map each history point to a single data point (1 point per block)
    const tps = [
      ...empty,
      ...tpsHistory.map<EstimatedTps>(
        ([total, vote, nonvote_success, nonvote_failed, tx_count]) => ({
          total,
          vote,
          nonvote_success,
          nonvote_failed,
          tx_count,
        }),
      ),
    ];
    setTpsData(tps.slice(tps.length - maxTransactionChartPoints));
  }, [setTpsData, tpsHistory]);
}
