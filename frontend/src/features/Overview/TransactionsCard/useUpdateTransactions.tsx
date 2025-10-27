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

    // Map each history point to a single data point (1 point per block)
    const historyData = tpsHistory.map<EstimatedTps>(
      ([total, vote, nonvote_success, nonvote_failed, tx_count]) => ({
        total,
        vote,
        nonvote_success,
        nonvote_failed,
        tx_count,
      }),
    );

    // Keep only the most recent maxTransactionChartPoints
    const recentData = historyData.slice(-maxTransactionChartPoints);

    // Pad with undefined at the start if we don't have enough data yet
    const empty: undefined[] = new Array<undefined>(
      Math.max(0, maxTransactionChartPoints - recentData.length),
    ).fill(undefined);

    // Start with empty slots on the left, data fills from left to right
    const tps = [...empty, ...recentData];

    setTpsData(tps);
  }, [setTpsData, tpsHistory]);
}
