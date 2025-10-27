// Timeline constants
export const TIMELINE_DURATION_SECONDS = 10; // Show last 10 seconds
export const MAX_VISIBLE_TXS = 50; // Maximum transactions to render at once

// Colors
export const TX_STATUS_COLORS = {
  pending: "#FFA500", // Orange
  processing: "#FFFF00", // Yellow
  success: "#00FF00", // Green
  failed: "#FF0000", // Red
} as const;
