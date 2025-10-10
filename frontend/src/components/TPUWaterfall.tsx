import React, { useEffect, useRef, useState } from 'react';

interface FlowNode {
  id: string;
  name: string;
  x: number;
  y: number;
  width: number;
  height: number;
  value: number;
  color: string;
}

interface FlowLink {
  source: string;
  target: string;
  value: number;
  color: string;
  path?: string;
}

interface TPUWaterfallProps {
  data?: any;
  width?: number;
  height?: number;
}

export function TPUWaterfall({ data, width = 800, height = 250 }: TPUWaterfallProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [nodes, setNodes] = useState<FlowNode[]>([]);
  const [links, setLinks] = useState<FlowLink[]>([]);

  // Convert Monad pipeline to Firedancer-like flow
  useEffect(() => {
    if (!data) {
      // Create default Monad pipeline
      createMonadPipeline();
    } else {
      createPipelineFromData(data);
    }
  }, [data]);

  const createMonadPipeline = () => {
    const pipelineNodes: FlowNode[] = [
      { id: 'rpc', name: 'RPC\nIngress', x: 50, y: 50, width: 80, height: 60, value: 8916, color: '#22d3ee' },
      { id: 'gossip', name: 'Gossip\nReceived', x: 50, y: 140, width: 80, height: 60, value: 822, color: '#3b82f6' },
      { id: 'mempool', name: 'Mempool', x: 200, y: 95, width: 80, height: 60, value: 9738, color: '#14b8a6' },
      { id: 'verify', name: 'Signature\nVerify', x: 350, y: 95, width: 80, height: 60, value: 9631, color: '#10b981' },
      { id: 'dedup', name: 'Nonce\nDedup', x: 500, y: 95, width: 80, height: 60, value: 4834, color: '#84cc16' },
      { id: 'evm', name: 'EVM\nExecution', x: 650, y: 95, width: 80, height: 60, value: 1627, color: '#eab308' },
      { id: 'pack', name: 'Block\nPacking', x: 800, y: 50, width: 80, height: 60, value: 1622, color: '#f97316' },
      { id: 'consensus', name: 'BFT\nConsensus', x: 800, y: 140, width: 80, height: 60, value: 13, color: '#ef4444' },
      { id: 'store', name: 'State\nStore', x: 950, y: 95, width: 80, height: 60, value: 13, color: '#8b5cf6' }
    ];

    const pipelineLinks: FlowLink[] = [
      { source: 'rpc', target: 'mempool', value: 8916, color: '#22d3ee' },
      { source: 'gossip', target: 'mempool', value: 822, color: '#3b82f6' },
      { source: 'mempool', target: 'verify', value: 9631, color: '#14b8a6' },
      { source: 'verify', target: 'dedup', value: 4834, color: '#10b981' },
      { source: 'dedup', target: 'evm', value: 1627, color: '#84cc16' },
      { source: 'evm', target: 'pack', value: 1622, color: '#eab308' },
      { source: 'evm', target: 'consensus', value: 13, color: '#eab308' },
      { source: 'pack', target: 'store', value: 13, color: '#f97316' },
      { source: 'consensus', target: 'store', value: 13, color: '#ef4444' }
    ];

    // Calculate link paths
    const linksWithPaths = pipelineLinks.map(link => ({
      ...link,
      path: calculateLinkPath(
        pipelineNodes.find(n => n.id === link.source)!,
        pipelineNodes.find(n => n.id === link.target)!,
        link.value
      )
    }));

    setNodes(pipelineNodes);
    setLinks(linksWithPaths);
  };

  const createPipelineFromData = (waterfallData: any) => {
    // Convert actual waterfall data to pipeline visualization
    // This would map real Monad metrics to the flow diagram
    createMonadPipeline(); // Fallback to default for now
  };

  const calculateLinkPath = (source: FlowNode, target: FlowNode, value: number): string => {
    const sourceX = source.x + source.width;
    const sourceY = source.y + source.height / 2;
    const targetX = target.x;
    const targetY = target.y + target.height / 2;

    const controlX1 = sourceX + (targetX - sourceX) * 0.3;
    const controlX2 = targetX - (targetX - sourceX) * 0.3;

    // Calculate path thickness based on value
    const thickness = Math.max(2, Math.min(20, value / 500));

    return `M${sourceX},${sourceY} C${controlX1},${sourceY} ${controlX2},${targetY} ${targetX},${targetY}`;
  };

  // Draw the waterfall
  useEffect(() => {
    if (!canvasRef.current || nodes.length === 0) return;

    const canvas = canvasRef.current;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Set canvas size for high DPI displays
    const rect = canvas.getBoundingClientRect();
    canvas.width = rect.width * window.devicePixelRatio;
    canvas.height = rect.height * window.devicePixelRatio;
    ctx.scale(window.devicePixelRatio, window.devicePixelRatio);

    // Clear canvas
    ctx.fillStyle = '#1a1f2b';
    ctx.fillRect(0, 0, rect.width, rect.height);

    // Draw links first (so they appear behind nodes)
    drawLinks(ctx);

    // Draw nodes
    drawNodes(ctx);
  }, [nodes, links]);

  const drawLinks = (ctx: CanvasRenderingContext2D) => {
    links.forEach(link => {
      const source = nodes.find(n => n.id === link.source);
      const target = nodes.find(n => n.id === link.target);

      if (!source || !target) return;

      const sourceX = source.x + source.width;
      const sourceY = source.y + source.height / 2;
      const targetX = target.x;
      const targetY = target.y + target.height / 2;

      const controlX1 = sourceX + (targetX - sourceX) * 0.4;
      const controlX2 = targetX - (targetX - sourceX) * 0.4;

      // Calculate stroke width based on value
      const strokeWidth = Math.max(1, Math.min(12, link.value / 800));

      // Draw flow line
      ctx.strokeStyle = link.color + '80'; // Semi-transparent
      ctx.lineWidth = strokeWidth;
      ctx.lineCap = 'round';

      ctx.beginPath();
      ctx.moveTo(sourceX, sourceY);
      ctx.bezierCurveTo(controlX1, sourceY, controlX2, targetY, targetX, targetY);
      ctx.stroke();

      // Draw flow highlight
      ctx.strokeStyle = link.color;
      ctx.lineWidth = Math.max(1, strokeWidth / 3);

      ctx.beginPath();
      ctx.moveTo(sourceX, sourceY);
      ctx.bezierCurveTo(controlX1, sourceY, controlX2, targetY, targetX, targetY);
      ctx.stroke();

      // Draw value label
      const midX = (sourceX + targetX) / 2;
      const midY = (sourceY + targetY) / 2 - 10;

      ctx.fillStyle = link.color;
      ctx.font = '10px Inter';
      ctx.textAlign = 'center';
      ctx.fillText(link.value.toLocaleString(), midX, midY);
    });
  };

  const drawNodes = (ctx: CanvasRenderingContext2D) => {
    nodes.forEach(node => {
      // Draw node box
      ctx.fillStyle = node.color + '20'; // Semi-transparent background
      ctx.strokeStyle = node.color;
      ctx.lineWidth = 2;

      ctx.fillRect(node.x, node.y, node.width, node.height);
      ctx.strokeRect(node.x, node.y, node.width, node.height);

      // Draw node name
      ctx.fillStyle = '#ffffff';
      ctx.font = '11px Inter';
      ctx.textAlign = 'center';

      const lines = node.name.split('\n');
      lines.forEach((line, index) => {
        ctx.fillText(
          line,
          node.x + node.width / 2,
          node.y + node.height / 2 - (lines.length - 1) * 6 + index * 12
        );
      });

      // Draw value
      ctx.fillStyle = node.color;
      ctx.font = 'bold 12px Inter';
      ctx.fillText(
        node.value.toLocaleString(),
        node.x + node.width / 2,
        node.y + node.height + 15
      );
    });
  };

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <canvas
        ref={canvasRef}
        style={{
          width: '100%',
          height: '100%',
          background: 'transparent'
        }}
      />

      {/* Legend */}
      <div style={{
        position: 'absolute',
        top: 10,
        right: 10,
        background: 'rgba(26, 31, 43, 0.8)',
        border: '1px solid #374151',
        borderRadius: '4px',
        padding: '8px',
        fontSize: '10px',
        color: '#b8bcc8'
      }}>
        <div>Flow: Transactions/sec</div>
        <div>Drops: Failed validations</div>
      </div>

      {/* Hover info */}
      <div style={{
        position: 'absolute',
        bottom: 10,
        left: 10,
        fontSize: '11px',
        color: '#8b909c'
      }}>
        next leader slot: {nodes.find(n => n.id === 'consensus')?.value || 0}
      </div>
    </div>
  );
}