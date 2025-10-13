import React, { useState, useEffect } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import { TPUWaterfall } from '../components/TPUWaterfall';
import type { WaterfallData } from '../types';

export function Overview() {
  const { metrics } = useWebSocket();
  const [waterfallData, setWaterfallData] = useState<WaterfallData | null>(null);

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

  const formatNumber = (num: number): string => {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num?.toString() || '0';
  };

  if (!metrics) {
    return (
      <div className="firedancer-loading">
        <div className="loading-spinner"></div>
        <p>Loading metrics...</p>
      </div>
    );
  }

  return (
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
              <div className="status-subvalue">synced</div>
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
            { name: 'rpc', count: Math.floor(metrics.waterfall.rpc_received) },
            { name: 'gossip', count: Math.floor(metrics.waterfall.gossip_received) },
            { name: 'mempool', count: Math.floor(metrics.waterfall.mempool_size) },
            { name: 'sig_verify', count: Math.floor(metrics.waterfall.signature_failed) },
            { name: 'nonce_dedup', count: Math.floor(metrics.waterfall.nonce_duplicate) },
            { name: 'gas_check', count: Math.floor(metrics.waterfall.gas_invalid) },
            { name: 'evm_parallel', count: Math.floor(metrics.waterfall.evm_parallel_executed) },
            { name: 'evm_sequential', count: Math.floor(metrics.waterfall.evm_sequential_fallback) },
            { name: 'state_conflict', count: Math.floor(metrics.waterfall.state_conflicts) },
            { name: 'bft_consensus', count: Math.floor(metrics.waterfall.bft_committed) },
            { name: 'state_update', count: Math.floor(metrics.waterfall.state_updated) },
            { name: 'broadcast', count: Math.floor(metrics.waterfall.blocks_broadcast) }
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
  );
}
