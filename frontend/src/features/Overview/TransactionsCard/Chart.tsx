import AutoSizer from "react-virtualized-auto-sizer";
import { useMemo, useRef } from "react";
import { isDefined } from "../../../utils";
import { useAtomValue } from "jotai";
import { tpsDataAtom } from "./atoms";
import {
  regularTextColor,
  transactionNonVotePathColor,
} from "../../../colors";

const getPath = (points: { x: number; y: number }[], height: number) => {
  if (!points.length) return "";

  const path = points.map(({ x, y }) => `L ${x} ${height - y}`).join(" ");

  return (
    "M" +
    path.slice(1) +
    `L ${points[points.length - 1].x} ${height} L ${points[0].x} ${height}, L ${points[0].x} ${points[0].y}`
  );
};

export default function Chart() {
  const tpsData = useAtomValue(tpsDataAtom);
  const sizeRefs = useRef<{ height: number; width: number }>();

  // Use Avg TPS (nonvote_success) for max scale instead of total
  const maxTotalTps = Math.max(...tpsData.map((d) => d?.nonvote_success ?? 0));

  const scaledPaths = useMemo(() => {
    if (!sizeRefs.current) return;
    if (!tpsData.length) return;

    const { height, width } = sizeRefs.current;
    const maxLength = tpsData.length;
    const xRatio = (width + 2) / maxLength;
    const yRatio = (height - 10) / (maxTotalTps || 1);

    const points = tpsData
      .map((d, i) => {
        if (d === undefined) return;

        // Only show Avg TPS (nonvote_success)
        return {
          x: i * xRatio,
          avgY: d.nonvote_success * yRatio, // Only Avg TPS
        };
      })
      .filter(isDefined);

    const maxTotalY = height - maxTotalTps * yRatio;

    return {
      avgPath: getPath(
        points.map((p) => ({ x: p.x, y: p.avgY })),
        height,
      ),
      totalTpsY: isNaN(maxTotalY) ? undefined : maxTotalY,
    };
  }, [maxTotalTps, tpsData]);

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
                <path
                  fillRule="evenodd"
                  clipRule="evenodd"
                  d={scaledPaths.avgPath}
                  fill={transactionNonVotePathColor}
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
                    <text
                      x="0"
                      y={scaledPaths.totalTpsY - 3}
                      fill={regularTextColor}
                      fontSize="8"
                      fontFamily="Inter Tight"
                    >
                      {maxTotalTps.toLocaleString()}
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
