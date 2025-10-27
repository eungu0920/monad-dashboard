import AutoSizer from "react-virtualized-auto-sizer";
import { useAtomValue } from "jotai";
import { transactionLogsAtom } from "./atoms";
import { useMemo, useState } from "react";
import { TIMELINE_DURATION_SECONDS, TX_STATUS_COLORS } from "./consts";
import type { TransactionLogData } from "./atoms";

export default function TransactionFlowTimeline() {
  const logs = useAtomValue(transactionLogsAtom);
  const [hoveredTx, setHoveredTx] = useState<TransactionLogData | null>(null);
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 });

  const timelineData = useMemo(() => {
    if (!logs.length) return [];

    const now = Date.now() / 1000; // Current time in seconds
    const startTime = now - TIMELINE_DURATION_SECONDS;

    // Filter logs within the timeline duration
    return logs
      .filter((log) => log.timestamp >= startTime)
      .map((log, index) => ({
        ...log,
        // Calculate relative position (0 = left edge, 1 = right edge/now)
        relativeTime: (log.timestamp - startTime) / TIMELINE_DURATION_SECONDS,
        displayIndex: index,
      }));
  }, [logs]);

  if (timelineData.length === 0) {
    return (
      <div
        style={{
          height: "300px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          opacity: 0.5,
        }}
      >
        <span style={{ fontSize: "14px", color: "var(--gray-11)" }}>
          No transactions in the last {TIMELINE_DURATION_SECONDS} seconds
        </span>
      </div>
    );
  }

  return (
    <div style={{ height: "300px", width: "100%" }}>
      <AutoSizer>
        {({ height, width }) => (
          <div style={{ position: "relative" }}>
            <svg
              width={width}
              height={height}
              style={{ backgroundColor: "var(--gray-a2)", borderRadius: "8px" }}
              onMouseMove={(e) => {
                const rect = e.currentTarget.getBoundingClientRect();
                setMousePos({
                  x: e.clientX - rect.left,
                  y: e.clientY - rect.top,
                });
              }}
              onMouseLeave={() => setHoveredTx(null)}
            >
            {/* Time axis */}
            <line
              x1={0}
              y1={height - 30}
              x2={width}
              y2={height - 30}
              stroke="var(--gray-6)"
              strokeWidth={2}
            />

            {/* Time labels */}
            <text
              x={width - 5}
              y={height - 10}
              fill="var(--gray-11)"
              fontSize="12"
              textAnchor="end"
            >
              Now
            </text>
            <text
              x={5}
              y={height - 10}
              fill="var(--gray-11)"
              fontSize="12"
              textAnchor="start"
            >
              -{TIMELINE_DURATION_SECONDS}s
            </text>

            {/* Transaction bars */}
            {timelineData.slice(0, 50).map((tx, index) => {
              const x = tx.relativeTime * width;
              const barWidth = 6;
              const barHeight = 20;
              const y = (height - 60) * (index / 50); // Distribute vertically

              const statusColor =
                TX_STATUS_COLORS[tx.status || "success"];
              const isHovered = hoveredTx?.transactionHash === tx.transactionHash;

              return (
                <g
                  key={`${tx.transactionHash}-${index}`}
                  style={{ cursor: "pointer" }}
                  onMouseEnter={() => setHoveredTx(tx)}
                  onClick={() => {
                    // Open in Monad explorer
                    window.open(
                      `https://explorer.testnet.monad.xyz/tx/${tx.transactionHash}`,
                      "_blank"
                    );
                  }}
                >
                  {/* Transaction bar */}
                  <rect
                    x={x - barWidth / 2}
                    y={y}
                    width={barWidth}
                    height={barHeight}
                    fill={statusColor}
                    opacity={isHovered ? 1 : 0.8}
                    rx={2}
                    stroke={isHovered ? "white" : "none"}
                    strokeWidth={isHovered ? 2 : 0}
                  />

                  {/* Connection line to time axis */}
                  <line
                    x1={x}
                    y1={y + barHeight}
                    x2={x}
                    y2={height - 30}
                    stroke={statusColor}
                    strokeWidth={1}
                    opacity={isHovered ? 0.6 : 0.3}
                    strokeDasharray="2,2"
                  />

                  {/* Invisible hover area for easier interaction */}
                  <rect
                    x={x - 15}
                    y={y - 5}
                    width={30}
                    height={barHeight + 10}
                    fill="transparent"
                  />
                </g>
              );
            })}
          </svg>

          {/* Tooltip */}
          {hoveredTx && (
            <div
              style={{
                position: "absolute",
                left: mousePos.x + 10,
                top: mousePos.y + 10,
                backgroundColor: "var(--gray-12)",
                color: "var(--gray-1)",
                padding: "8px 12px",
                borderRadius: "6px",
                fontSize: "12px",
                pointerEvents: "none",
                zIndex: 1000,
                boxShadow: "0 4px 12px rgba(0,0,0,0.3)",
                maxWidth: "300px",
              }}
            >
              <div style={{ fontFamily: "monospace", marginBottom: "4px" }}>
                <strong>Tx:</strong> {hoveredTx.transactionHash.slice(0, 10)}...{hoveredTx.transactionHash.slice(-8)}
              </div>
              <div>
                <strong>Block:</strong> #{hoveredTx.blockNumber}
              </div>
              <div>
                <strong>Index:</strong> {hoveredTx.transactionIndex}
              </div>
              <div>
                <strong>Time:</strong> {new Date(hoveredTx.timestamp * 1000).toLocaleTimeString()}
              </div>
              <div style={{ marginTop: "4px", fontSize: "10px", opacity: 0.7 }}>
                Click to view in explorer
              </div>
            </div>
          )}
        </div>
        )}
      </AutoSizer>
    </div>
  );
}
