import AutoSizer from "react-virtualized-auto-sizer";
import { useMemo, useRef } from "react";
import { isDefined } from "../../../utils";
import { useAtomValue } from "jotai";
import { tpsDataAtom } from "./atoms";
import { maxTransactionChartPoints } from "./consts";
import {
  regularTextColor,
  transactionNonVotePathColor,
  transactionTxCountPathColor,
} from "../../../colors";

const getPath = (points: { x: number; y: number }[], height: number) => {
  if (!points.length) return "";

  const path = points.map(({ x, y }) => `L ${x} ${y}`).join(" ");

  return (
    "M" +
    path.slice(1) +
    `L ${points[points.length - 1].x} ${height} L ${points[0].x} ${height} L ${points[0].x} ${points[0].y}`
  );
};

export default function Chart() {
  const tpsData = useAtomValue(tpsDataAtom);
  const sizeRefs = useRef<{ height: number; width: number }>();

  // Calculate separate max values for independent scaling
  const maxTotalTps = Math.max(...tpsData.map((d) => d?.nonvote_success ?? 0));
  const maxTxCount = Math.max(...tpsData.map((d) => d?.tx_count ?? 0));

  const scaledPaths = useMemo(() => {
    if (!sizeRefs.current) return;
    if (!tpsData.length) return;

    const { height, width } = sizeRefs.current;
    // Use fixed maxTransactionChartPoints for x-axis, not actual data length
    // This ensures graph fills full width even when data is still accumulating
    const xRatio = (width + 2) / maxTransactionChartPoints;

    // Add padding and use full height for better visualization
    const padding = 15;
    const tpsYRatio = (height - padding * 2) / (maxTotalTps || 1);
    const txYRatio = (height - padding * 2) / (maxTxCount || 1);

    const points = tpsData
      .map((d, i) => {
        if (d === undefined) return undefined;

        // Calculate both TPS and Tx count with independent scaling
        const tpsValue = d.nonvote_success * tpsYRatio;
        const txValue = (d.tx_count ?? 0) * txYRatio;
        return {
          x: i * xRatio, // Keep original index for correct x position
          avgY: height - padding - tpsValue, // Avg TPS (green) - independent scale
          txY: height - padding - txValue,   // Tx count (blue) - independent scale
        };
      })
      .filter(isDefined);

    // Position the max line near the top with padding
    const maxTotalY = padding;

    return {
      avgPath: getPath(
        points.map((p) => ({ x: p.x, y: p.avgY })),
        height,
      ),
      txPath: getPath(
        points.map((p) => ({ x: p.x, y: p.txY })),
        height,
      ),
      totalTpsY: maxTotalY,
      maxTotalTps,
      maxTxCount,
    };
  }, [maxTotalTps, maxTxCount, tpsData]);

  return (
    <>
      <AutoSizer>
        {({ height, width }) => {
          sizeRefs.current = { height, width };
          if (!scaledPaths) return null;
          return (
            <>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width={width}
                height={height}
                fill="none"
              >
                {/* Avg TPS path (green) */}
                <path
                  fillRule="evenodd"
                  clipRule="evenodd"
                  d={scaledPaths.avgPath}
                  fill={transactionNonVotePathColor}
                  opacity={0.7}
                />

                {/* Tx count path (blue) */}
                <path
                  fillRule="evenodd"
                  clipRule="evenodd"
                  d={scaledPaths.txPath}
                  fill={transactionTxCountPathColor}
                  opacity={0.7}
                />

                {scaledPaths.totalTpsY && (
                  <>
                    <line
                      x1="0"
                      y1={scaledPaths.totalTpsY}
                      x2={width}
                      y2={scaledPaths.totalTpsY}
                      strokeDasharray="4"
                      stroke="rgba(255, 255, 255, 0.30)"
                    />
                    {/* TPS max label (left, green) */}
                    <text
                      x="0"
                      y={scaledPaths.totalTpsY - 3}
                      fill={transactionNonVotePathColor}
                      fontSize="8"
                      fontFamily="Inter Tight"
                    >
                      {scaledPaths.maxTotalTps.toLocaleString()} TPS
                    </text>
                    {/* Tx max label (right, blue) */}
                    <text
                      x={width - 40}
                      y={scaledPaths.totalTpsY - 3}
                      fill={transactionTxCountPathColor}
                      fontSize="8"
                      fontFamily="Inter Tight"
                      textAnchor="end"
                    >
                      {scaledPaths.maxTxCount.toLocaleString()} Tx
                    </text>
                  </>
                )}
              </svg>
            </>
          );
        }}
      </AutoSizer>
    </>
  );
}
