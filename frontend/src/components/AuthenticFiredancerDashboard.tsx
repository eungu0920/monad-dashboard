import React, { useState, useEffect, useRef } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import { TPUWaterfall } from './TPUWaterfall';
import type { MonadMetrics, WaterfallData } from '../types';
import '../styles/firedancer-authentic.css';

interface TPUStage {
  name: string;
  in: number;
  out: number;
  drop: number;
  rate?: number;
}

export function AuthenticFiredancerDashboard() {
  const { metrics, connected, error } = useWebSocket();
  const [waterfallData, setWaterfallData] = useState<WaterfallData | null>(null);
  const [eventRingsStatus, setEventRingsStatus] = useState<any>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  // Fetch waterfall data
  useEffect(() => {
    const fetchWaterfall = async () => {
      try {
        const response = await fetch('/api/v1/waterfall');
        if (response.ok) {
          const data = await response.json();
          setWaterfallData(data);
        }
      } catch (err) {
        console.error('Failed to fetch waterfall data:', err);
      }
    };

    fetchWaterfall();
    const interval = setInterval(fetchWaterfall, 2000);
    return () => clearInterval(interval);
  }, []);

  // Fetch event rings status
  useEffect(() => {
    const fetchEventRings = async () => {
      try {
        const response = await fetch('/api/v1/event-rings');
        if (response.ok) {
          const data = await response.json();
          setEventRingsStatus(data);
        }
      } catch (err) {
        console.error('Failed to fetch event rings status:', err);
      }
    };

    fetchEventRings();
    const interval = setInterval(fetchEventRings, 5000);
    return () => clearInterval(interval);
  }, []);

  // Draw TPU Waterfall Canvas
  useEffect(() => {
    if (!waterfallData || !canvasRef.current) return;

    const canvas = canvasRef.current;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Set canvas size
    canvas.width = canvas.offsetWidth * window.devicePixelRatio;
    canvas.height = canvas.offsetHeight * window.devicePixelRatio;
    ctx.scale(window.devicePixelRatio, window.devicePixelRatio);

    const width = canvas.offsetWidth;
    const height = canvas.offsetHeight;

    // Clear canvas
    ctx.fillStyle = '#1a1f2b';
    ctx.fillRect(0, 0, width, height);

    // Draw flow diagram
    drawTPUWaterfall(ctx, width, height, waterfallData);
  }, [waterfallData]);

  const drawTPUWaterfall = (ctx: CanvasRenderingContext2D, width: number, height: number, data: any) => {
    // This is a simplified version - in a real implementation, you'd create a proper Sankey diagram
    const stages = data.stages || [];
    const stageWidth = width / stages.length;
    const centerY = height / 2;

    ctx.strokeStyle = '#22d3ee';
    ctx.fillStyle = '#22d3ee';
    ctx.lineWidth = 2;

    stages.forEach((stage: any, index: number) => {
      const x = index * stageWidth + stageWidth / 2;
      const y = centerY;

      // Draw stage box
      ctx.fillRect(x - 40, y - 20, 80, 40);

      // Draw stage name
      ctx.fillStyle = '#0f1419';
      ctx.font = '12px Inter';
      ctx.textAlign = 'center';
      ctx.fillText(stage.name, x, y + 4);

      // Draw connections
      if (index < stages.length - 1) {
        ctx.strokeStyle = '#22d3ee';
        ctx.beginPath();
        ctx.moveTo(x + 40, y);
        ctx.lineTo(x + stageWidth - 40, y);
        ctx.stroke();
      }

      // Reset fill style
      ctx.fillStyle = '#22d3ee';
    });
  };

  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);

    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${mins}m`;
    return `${mins}m`;
  };

  if (error) {
    return (
      <div className="firedancer-authentic">
        <div className="firedancer-navbar">
          <div className="firedancer-logo">
            <div className="logo-icon">M</div>
            <span className="logo-text">monad</span>
          </div>
          <div className="navbar-right">
            <span style={{ color: 'var(--fd-status-error)' }}>CONNECTION ERROR: {error}</span>
          </div>
        </div>
      </div>
    );
  }

  if (!metrics) {
    return (
      <div className="firedancer-authentic">
        <div className="firedancer-navbar">
          <div className="firedancer-logo">
            <div className="logo-icon">M</div>
            <span className="logo-text">monad</span>
          </div>
          <div className="navbar-right">
            <span>Connecting to node...</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="firedancer-authentic">
      {/* Navigation Bar */}
      <div className="firedancer-navbar">
        <div className="firedancer-logo">
          <div className="logo-icon">M</div>
          <span className="logo-text">monad</span>
          <span className="network-badge">v{metrics.node_info.version}</span>
        </div>

        <div className="navbar-center">
          <div className="nav-item active">Overview</div>
          <div className="nav-item">Validators</div>
          <div className="nav-item">Network</div>
        </div>

        <div className="navbar-right">
          <div className="epoch-info">
            <span>Block #{formatNumber(metrics.consensus.current_height)}</span>
            <span>•</span>
            <span>{connected ? 'ONLINE' : 'OFFLINE'}</span>
            <span>•</span>
            <span>Chain {metrics.node_info.chain_id}</span>
          </div>
          <div className="realtime-indicator">
            <div className="pulse-dot"></div>
            <span>Realtime</span>
          </div>
          <span>{metrics.node_info.node_name}</span>
          <span>UP {formatUptime(metrics.node_info.uptime)}</span>
          <span>TPS: {metrics.execution.tps.toFixed(1)}</span>
        </div>
      </div>

      {/* Main Content */}
      <div className="firedancer-main">
        {/* Status Panel */}
        <div className="firedancer-panel status-panel">
          <div className="panel-header">Node Status</div>
          <div className="panel-content">
            <div className="status-grid">
              <div className="status-item">
                <div className="status-label">Block Height</div>
                <div className="status-value">{formatNumber(metrics.consensus.current_height)}</div>
              </div>

              <div className="status-item">
                <div className="status-label">Consensus Status</div>
                <div className="status-subvalue">{connected ? 'synced' : 'syncing'}</div>
                <div className="status-secondary">{(metrics.consensus.participation_rate * 100).toFixed(1)}% participation</div>
              </div>

              <div className="status-item">
                <div className="status-label">Block Time</div>
                <div className="status-value">{metrics.consensus.block_time.toFixed(1)}s</div>
              </div>

              <div className="status-item">
                <div className="status-label">Last Block Time</div>
                <div className="status-value">{Math.floor((Date.now() / 1000 - metrics.consensus.last_block_time))}s ago</div>
              </div>
            </div>
          </div>
        </div>

        {/* Transactions Panel */}
        <div className="firedancer-panel transactions-panel">
          <div className="panel-header">Transactions</div>
          <div className="panel-content">
            <div className="status-item">
              <div className="status-label">Last TPS</div>
              <div className="status-value">{metrics.execution.tps.toFixed(2)}</div>
            </div>

            <div className="transactions-chart">
              {/* Placeholder for TPS chart */}
              <svg width="100%" height="100%">
                <defs>
                  <linearGradient id="tpsGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                    <stop offset="0%" stopColor="#22d3ee" stopOpacity="0.3"/>
                    <stop offset="100%" stopColor="#22d3ee" stopOpacity="0"/>
                  </linearGradient>
                </defs>
                <path
                  d="M0,80 Q30,60 60,70 T120,65 T180,55 T240,60 T300,50"
                  stroke="#22d3ee"
                  strokeWidth="2"
                  fill="none"
                />
                <path
                  d="M0,80 Q30,60 60,70 T120,65 T180,55 T240,60 T300,50 L300,120 L0,120 Z"
                  fill="url(#tpsGradient)"
                />
              </svg>
            </div>

            <div className="tps-metrics">
              <div className="tps-metric">
                <span className="tps-value success">
                  {Math.floor(metrics.waterfall.evm_parallel_executed / 10)}
                </span>
                <div className="tps-label">Parallel Success</div>
              </div>
              <div className="tps-metric">
                <span className="tps-value error">
                  {Math.floor(metrics.waterfall.signature_failed + metrics.waterfall.gas_invalid)}
                </span>
                <div className="tps-label">Validation Fail</div>
              </div>
              <div className="tps-metric">
                <span className="tps-value info">
                  {Math.floor(metrics.waterfall.mempool_size / 100)}
                </span>
                <div className="tps-label">Mempool Size</div>
              </div>
            </div>
          </div>
        </div>

        {/* Network Panel */}
        <div className="firedancer-panel validators-panel">
          <div className="panel-header">Network & Consensus</div>
          <div className="panel-content">
            <div className="validators-grid">
              <div className="validator-metric">
                <div className="status-label">Active Validators</div>
                <div className="validator-count">{metrics.consensus.validator_count}</div>
              </div>
              <div className="validator-metric">
                <div className="status-label">Total Voting Power</div>
                <div className="validator-stake">{formatNumber(metrics.consensus.voting_power)} MONAD</div>
              </div>
              <div className="validator-metric">
                <div className="status-label">Network Peers</div>
                <div className="validator-count">{metrics.network.peer_count}</div>
              </div>
              <div className="validator-metric">
                <div className="status-label">Network Latency</div>
                <div className="validator-stake">
                  {metrics.network.network_latency.toFixed(0)}ms
                </div>
              </div>
            </div>

            {/* Circular progress indicator */}
            <div style={{ textAlign: 'center', marginTop: '16px' }}>
              <div className="validator-percentage">
                {(metrics.execution.parallel_success_rate * 100).toFixed(1)}%
              </div>
              <div className="status-secondary">Parallel Execution Rate</div>
            </div>
          </div>
        </div>

        {/* Transaction Processing Waterfall */}
        <div className="firedancer-panel waterfall-panel">
          <div className="panel-header">Monad Transaction Processing Waterfall</div>
          <div className="panel-content">
            <div className="waterfall-canvas">
              <TPUWaterfall data={waterfallData} />
            </div>
          </div>
        </div>

        {/* Pipeline Stages */}
        <div className="pipeline-panel">
          <div className="pipeline-stages">
            {[
              { name: 'rpc', count: metrics ? Math.floor(metrics.waterfall.rpc_received) : 8916 },
              { name: 'gossip', count: metrics ? Math.floor(metrics.waterfall.gossip_received) : 822 },
              { name: 'mempool', count: metrics ? Math.floor(metrics.waterfall.mempool_size) : 9738 },
              { name: 'sig_verify', count: metrics ? Math.floor(metrics.waterfall.signature_failed) : 107 },
              { name: 'nonce_dedup', count: metrics ? Math.floor(metrics.waterfall.nonce_duplicate) : 203 },
              { name: 'gas_check', count: metrics ? Math.floor(metrics.waterfall.gas_invalid) : 238 },
              { name: 'evm_parallel', count: metrics ? Math.floor(metrics.waterfall.evm_parallel_executed) : 3800 },
              { name: 'evm_sequential', count: metrics ? Math.floor(metrics.waterfall.evm_sequential_fallback) : 700 },
              { name: 'state_conflict', count: metrics ? Math.floor(metrics.waterfall.state_conflicts) : 450 },
              { name: 'bft_consensus', count: metrics ? Math.floor(metrics.waterfall.bft_committed) : 13 },
              { name: 'state_update', count: metrics ? Math.floor(metrics.waterfall.state_updated) : 13 },
              { name: 'broadcast', count: metrics ? Math.floor(metrics.waterfall.blocks_broadcast) : 13 }
            ].map((stage, index) => (
              <div key={index} className="pipeline-stage">
                <div className="stage-name">{stage.name}</div>
                <div className="stage-chart">
                  <svg className="sparkline" width="100%" height="100%">
                    <path
                      d={`M0,${40-Math.random()*20} ${Array.from({length: 20}, (_, i) =>
                        `L${i*5},${40-Math.random()*30}`).join(' ')}`}
                      fill="none"
                      stroke={stage.count > 0 ? "#22d3ee" : "#4b5563"}
                      strokeWidth="1"
                    />
                  </svg>
                </div>
                <div className="stage-metrics">
                  <span className="stage-value">{stage.count}</span>
                  <span>0%</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}