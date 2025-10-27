import { useEffect, useRef } from "react";
import { useSetAtom } from "jotai";
import { transactionLogsAtom } from "./atoms";
import type { TransactionLogData } from "./atoms";

const MAX_LOGS = 200; // Keep only the most recent 200 logs

export default function useTransactionFlow() {
  const setTransactionLogs = useSetAtom(transactionLogsAtom);
  const throttleTimeoutRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    // This hook will receive WebSocket messages from the parent component
    // For now, we'll set up the atom update logic

    const handleTransactionLog = (log: TransactionLogData) => {
      // Throttle updates to every 100ms for performance
      if (throttleTimeoutRef.current) {
        clearTimeout(throttleTimeoutRef.current);
      }

      throttleTimeoutRef.current = setTimeout(() => {
        setTransactionLogs((prev) => {
          // Add new log at the beginning (newest first)
          const updated = [log, ...prev];

          // Keep only the most recent MAX_LOGS
          return updated.slice(0, MAX_LOGS);
        });
      }, 100);
    };

    // Cleanup
    return () => {
      if (throttleTimeoutRef.current) {
        clearTimeout(throttleTimeoutRef.current);
      }
    };
  }, [setTransactionLogs]);

  return null;
}
