import { atom } from "jotai";

export interface TransactionLogData {
  blockNumber: number;
  transactionHash: string;
  transactionIndex: number;
  address: string;
  topics: string[];
  data: string;
  timestamp: number;
  status?: "pending" | "processing" | "success" | "failed";
}

// Store recent transaction logs (max 200)
export const transactionLogsAtom = atom<TransactionLogData[]>([]);

// Filter options
export const txFlowFilterAtom = atom<{
  showSuccess: boolean;
  showFailed: boolean;
}>({
  showSuccess: true,
  showFailed: true,
});
