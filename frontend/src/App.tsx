import React, { useState, useEffect } from 'react';
import { useWebSocket } from './hooks/useWebSocket';
import { MetricCard } from './components/MetricCard';
import { WaterfallChart } from './components/WaterfallChart';
import { ThroughputChart } from './components/ThroughputChart';
import type { MonadMetrics, WaterfallData, ExecutionMetrics } from './types';

function App() {
  const { metrics, connected, error } = useWebSocket();
  const [waterfallData, setWaterfallData] = useState<WaterfallData | null>(null);
  const [throughputHistory, setThroughputHistory] = useState<ExecutionMetrics[]>([]);

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

  // Track throughput history
  useEffect(() => {
    if (metrics?.execution) {
      setThroughputHistory(prev => {
        const newHistory = [...prev, metrics.execution];
        // Keep last 60 data points (1 minute at 1 second intervals)
        return newHistory.slice(-60);
      });
    }
  }, [metrics]);

  const formatUptime = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const formatBytes = (bytes: number): string => {
    if (bytes >= 1024 * 1024 * 1024) {
      return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
    } else if (bytes >= 1024 * 1024) {
      return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
    } else if (bytes >= 1024) {
      return `${(bytes / 1024).toFixed(2)} KB`;
    }
    return `${bytes} B`;
  };

  if (error) {
    return (
      <div className="dashboard">
        <div className="error">
          Connection Error: {error}
          <br />
          <button
            onClick={() => window.location.reload()}
            style={{
              marginTop: '1rem',
              padding: '0.5rem 1rem',
              background: 'var(--monad-primary)',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer'
            }}
          >
            Retry Connection
          </button>
        </div>
      </div>
    );
  }

  if (!metrics) {
    return (
      <div className="dashboard">
        <div className="loading">Connecting to Monad node...</div>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <div className="dashboard-title">
          <h1>Monad Dashboard</h1>
          <span style={{
            fontSize: '0.875rem',
            color: 'var(--monad-text-muted)',
            background: 'var(--monad-card)',
            padding: '0.25rem 0.75rem',
            borderRadius: '4px',
            border: '1px solid var(--monad-border)'
          }}>
            {metrics.node_info.node_name} • Chain ID {metrics.node_info.chain_id}
          </span>
        </div>

        <div className="status-indicator">
          <div className="status-dot" style={{
            backgroundColor: connected ? 'var(--monad-accent)' : 'var(--monad-danger)'
          }} />
          <span>{connected ? 'Connected' : 'Disconnected'}</span>
          <span style={{ color: 'var(--monad-text-muted)', marginLeft: '1rem' }}>
            v{metrics.node_info.version}
          </span>
        </div>
      </header>

      {/* Key Metrics */}
      <div className="metrics-grid">
        <MetricCard
          title="Transaction Throughput"
          value={metrics.execution.tps.toFixed(0)}
          unit="TPS"
          status="good"
        />
        <MetricCard
          title="Block Height"
          value={metrics.consensus.current_height.toLocaleString()}
          status="good"
        />
        <MetricCard
          title="Mempool Size"
          value={metrics.waterfall.mempool_size}
          unit="txns"
          status={metrics.waterfall.mempool_size > 10000 ? 'warning' : 'good'}
        />
        <MetricCard
          title="Parallel Success Rate"
          value={(metrics.execution.parallel_success_rate * 100).toFixed(1)}
          unit="%"
          status={metrics.execution.parallel_success_rate > 0.8 ? 'good' : 'warning'}
        />
        <MetricCard
          title="Peer Count"
          value={metrics.network.peer_count}
          status={metrics.network.peer_count > 30 ? 'good' : 'warning'}
        />
        <MetricCard
          title="Uptime"
          value={formatUptime(metrics.node_info.uptime)}
          status="good"
        />
      </div>

      {/* Waterfall Visualization */}
      <WaterfallChart data={waterfallData} />

      {/* Throughput Chart */}
      <ThroughputChart data={throughputHistory} />

      {/* Additional Metrics */}
      <div className="metrics-grid">
        <MetricCard
          title="Gas Price"
          value={metrics.execution.avg_gas_price}
          unit="gwei"
        />
        <MetricCard
          title="Block Time"
          value={metrics.consensus.block_time.toFixed(2)}
          unit="sec"
          status={metrics.consensus.block_time < 2 ? 'good' : 'warning'}
        />
        <MetricCard
          title="Validator Count"
          value={metrics.consensus.validator_count}
        />
        <MetricCard
          title="Network Latency"
          value={metrics.network.network_latency.toFixed(0)}
          unit="ms"
          status={metrics.network.network_latency < 100 ? 'good' : 'warning'}
        />
        <MetricCard
          title="State Size"
          value={formatBytes(metrics.execution.state_size)}
        />
        <MetricCard
          title="Participation Rate"
          value={(metrics.consensus.participation_rate * 100).toFixed(1)}
          unit="%"
          status={metrics.consensus.participation_rate > 0.8 ? 'good' : 'warning'}
        />
      </div>
    </div>
  );
}

export default App;