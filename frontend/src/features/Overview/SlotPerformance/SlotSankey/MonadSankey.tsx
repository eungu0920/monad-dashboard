import { useMemo } from "react";
import { Sankey } from "../../../../sankey";
import AutoSizer from "react-virtualized-auto-sizer";
import { useAtomValue } from "jotai";
import { monadWaterfallV2Atom } from "../../../../api/atoms";
import { Flex, Spinner, Text } from "@radix-ui/themes";

/**
 * MonadSankey - Monad-specific transaction waterfall visualization
 *
 * Uses the Monad 7-stage transaction lifecycle:
 * 1. Submission (RPC/P2P)
 * 2. Mempool (Validation)
 * 3. Leader Propagation
 * 4. Block Building
 * 5. Consensus (Proposed → Voted → Finalized)
 * 6. Execution (Parallel)
 * 7. State Update → 2-Block Finality
 */
export default function MonadSankey() {
  const waterfallV2 = useAtomValue(monadWaterfallV2Atom);

  const data = useMemo(() => {
    if (!waterfallV2) return null;

    // Backend already provides nodes and links in Sankey-compatible format
    return {
      nodes: waterfallV2.nodes,
      links: waterfallV2.links,
    };
  }, [waterfallV2]);

  if (!data || !data.links.length) {
    return (
      <Flex justify="center" align="center" height="100%">
        <Spinner style={{ height: 50, width: 50 }} />
      </Flex>
    );
  }

  return (
    <AutoSizer>
      {({ height, width }) => {
        const isRotated = width < 600;
        if (isRotated) {
          const swap = height;
          height = width;
          width = swap;
        }
        return (
          <Sankey
            height={height}
            width={width}
            data={data}
            margin={{
              top: 20,
              right: isRotated ? 180 : 160,
              bottom: 50,
              left: 120,
            }}
            align="center"
            isInteractive={false}
            nodeThickness={0}
            nodeSpacing={getNodeSpacing(height)}
            nodeBorderWidth={1}
            sort="input"
            nodeBorderRadius={3}
            linkOpacity={1}
            enableLinkGradient
            labelPosition="outside"
            labelPadding={20}
          />
        );
      }}
    </AutoSizer>
  );
}

function getNodeSpacing(height: number) {
  if (height < 275) {
    return 32;
  } else if (height < 300) {
    return 36;
  } else if (height < 325) {
    return 40;
  } else if (height < 350) {
    return 48;
  }
  return 52;
}
