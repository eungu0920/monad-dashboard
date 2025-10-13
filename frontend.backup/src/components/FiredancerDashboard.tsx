import React, { useState, useEffect } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import type { MonadMetrics, WaterfallData } from '../types';
import '../styles/firedancer.css';

interface LogEntry {
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'success';
  message: string;
}

export function FiredancerDashboard() {
  const { metrics, connected, error } = useWebSocket();
  const [waterfallData, setWaterfallData] = useState<WaterfallData | null>(null);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [eventRingsStatus, setEventRingsStatus] = useState<any>(null);

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

  // Simulate real-time logs
  useEffect(() => {
    const addLog = (level: LogEntry['level'], message: string) => {
      const timestamp = new Date().toISOString().split('T')[1].split('.')[0];
      const newLog: LogEntry = { timestamp, level, message };

      setLogs(prev => {
        const updated = [...prev, newLog];
        return updated.slice(-100); // Keep last 100 logs
      });
    };

    if (metrics) {
      addLog('success', `Block #${metrics.consensus.current_height.toLocaleString()} processed`);
      addLog('info', `TPS: ${metrics.execution.tps.toFixed(1)}, Mempool: ${metrics.waterfall.mempool_size}`);

      if (metrics.execution.parallel_success_rate < 0.8) {
        addLog('warn', `Parallel execution rate low: ${(metrics.execution.parallel_success_rate * 100).toFixed(1)}%`);
      }
    }

    if (connected) {
      addLog('success', 'Connected to Monad node');
    } else if (error) {
      addLog('error', `Connection error: ${error}`);
    }
  }, [metrics, connected, error]);

  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);

    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${mins}m`;
    return `${mins}m`;
  };

  const formatBytes = (bytes: number): string => {
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let size = bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }

    return `${size.toFixed(2)} ${units[unitIndex]}`;
  };

  const getStatusClass = (value: number, goodThreshold: number, warningThreshold?: number): string => {
    if (value >= goodThreshold) return 'good';
    if (warningThreshold && value >= warningThreshold) return 'warning';
    return 'error';
  };

  if (error) {
    return (
      <div className="firedancer-dashboard">
        <div className="firedancer-panel">
          <div className="firedancer-panel-header">CONNECTION ERROR</div>
          <div className="firedancer-panel-content">
            <div style={{ color: 'var(--fd-status-error)', textAlign: 'center', padding: '20px' }}>
              <div style={{ fontSize: '16px', marginBottom: '10px' }}>⚠ {error}</div>
              <button
                onClick={() => window.location.reload()}
                style={{
                  background: 'var(--fd-accent-blue)',
                  color: 'var(--fd-bg-primary)',
                  border: 'none',
                  padding: '8px 16px',
                  borderRadius: '4px',
                  fontFamily: 'inherit',
                  cursor: 'pointer'
                }}
              >
                RETRY CONNECTION
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!metrics) {
    return (
      <div className="firedancer-dashboard">
        <div className="firedancer-panel">
          <div className="firedancer-panel-header">INITIALIZING</div>
          <div className="firedancer-panel-content">
            <div style={{ color: 'var(--fd-status-info)', textAlign: 'center', padding: '20px' }}>
              <div style={{ fontSize: '16px' }}>◉ Connecting to Monad node...</div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="firedancer-dashboard">
      {/* Header */}
      <div className="firedancer-header">
        <div className="firedancer-title">
          <h1>MONAD // DASHBOARD</h1>
          <div className="firedancer-subtitle">
            {metrics.node_info.node_name} • Chain #{metrics.node_info.chain_id} • v{metrics.node_info.version}
          </div>
        </div>
        <div className="firedancer-status">
          <div className={`led-indicator ${connected ? 'connected' : 'disconnected'}`}></div>
          <span>{connected ? 'ONLINE' : 'OFFLINE'}</span>
          <span style={{ color: 'var(--fd-text-muted)', marginLeft: '10px' }}>
            UP {formatUptime(metrics.node_info.uptime)}
          </span>
        </div>
      </div>

      {/* Main Metrics Grid */}
      <div className="firedancer-grid">
        {/* Node Status */}
        <div className="firedancer-panel">
          <div className="firedancer-panel-header">NODE STATUS</div>
          <div className="firedancer-panel-content">
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Block Height</span>
              <span className="firedancer-metric-value good">
                #{metrics.consensus.current_height.toLocaleString()}
              </span>
            </div>
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Block Time</span>
              <span className={`firedancer-metric-value ${metrics.consensus.block_time < 2 ? 'good' : 'warning'}`}>
                {metrics.consensus.block_time.toFixed(2)}s
              </span>
            </div>
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Validator Count</span>
              <span className="firedancer-metric-value info">
                {metrics.consensus.validator_count}
              </span>
            </div>
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Participation</span>
              <span className={`firedancer-metric-value ${getStatusClass(metrics.consensus.participation_rate, 0.8, 0.6)}`}>
                {(metrics.consensus.participation_rate * 100).toFixed(1)}%
              </span>
            </div>
          </div>
        </div>

        {/* Execution Metrics */}
        <div className="firedancer-panel">
          <div className="firedancer-panel-header">EXECUTION ENGINE</div>
          <div className="firedancer-panel-content">
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Transaction Throughput</span>
              <span className="firedancer-metric-value good">
                {metrics.execution.tps.toFixed(0)} TPS
              </span>
            </div>
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Mempool Size</span>
              <span className={`firedancer-metric-value ${metrics.waterfall.mempool_size > 10000 ? 'warning' : 'good'}`}>
                {metrics.waterfall.mempool_size.toLocaleString()}
              </span>
            </div>
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Parallel Success</span>
              <span className={`firedancer-metric-value ${getStatusClass(metrics.execution.parallel_success_rate, 0.8, 0.6)}`}>
                {(metrics.execution.parallel_success_rate * 100).toFixed(1)}%
              </span>
            </div>
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Gas Price</span>
              <span className="firedancer-metric-value info">
                {metrics.execution.avg_gas_price} gwei
              </span>
            </div>
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">State Size</span>
              <span className="firedancer-metric-value info">
                {formatBytes(metrics.execution.state_size)}
              </span>
            </div>
          </div>
        </div>

        {/* Network Status */}
        <div className="firedancer-panel">
          <div className="firedancer-panel-header">NETWORK</div>
          <div className="firedancer-panel-content">
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Total Peers</span>
              <span className={`firedancer-metric-value ${getStatusClass(metrics.network.peer_count, 30, 10)}`}>
                {metrics.network.peer_count}
              </span>
            </div>
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Inbound / Outbound</span>
              <span className="firedancer-metric-value info">
                {metrics.network.inbound_peers} / {metrics.network.outbound_peers}
              </span>
            </div>
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Network Latency</span>
              <span className={`firedancer-metric-value ${getStatusClass(100 - metrics.network.network_latency, 50, 20)}`}>
                {metrics.network.network_latency.toFixed(0)}ms
              </span>
            </div>
            <div className="firedancer-metric">
              <span className="firedancer-metric-label">Data In / Out</span>
              <span className="firedancer-metric-value info">
                {formatBytes(metrics.network.bytes_in)} / {formatBytes(metrics.network.bytes_out)}
              </span>
            </div>
          </div>
        </div>

        {/* Event Rings Status */}
        {eventRingsStatus && (
          <div className="firedancer-panel">
            <div className="firedancer-panel-header">EVENT RINGS</div>
            <div className="firedancer-panel-content">
              {eventRingsStatus.connected ? (
                <>
                  <div className="firedancer-metric">
                    <span className="firedancer-metric-label">Status</span>
                    <span className="firedancer-metric-value good">CONNECTED</span>
                  </div>
                  <div className="firedancer-metric">
                    <span className="firedancer-metric-label">Events Received</span>
                    <span className="firedancer-metric-value info">
                      {eventRingsStatus.events_received?.toLocaleString() || 0}
                    </span>
                  </div>
                  <div className="firedancer-metric">
                    <span className="firedancer-metric-label">Buffer Usage</span>
                    <span className="firedancer-metric-value info">
                      {eventRingsStatus.buffer_size || 0}/1000
                    </span>
                  </div>
                  <div className="firedancer-metric">
                    <span className="firedancer-metric-label">Missed Events</span>
                    <span className={`firedancer-metric-value ${(eventRingsStatus.missed_events || 0) > 0 ? 'warning' : 'good'}`}>
                      {eventRingsStatus.missed_events || 0}
                    </span>
                  </div>
                </>
              ) : (
                <div className="firedancer-metric">
                  <span className="firedancer-metric-label">Status</span>
                  <span className="firedancer-metric-value warning">RPC MODE</span>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Transaction Waterfall */}
      {waterfallData && (
        <div className="firedancer-waterfall">
          <div className="firedancer-panel-header" style={{ marginBottom: '15px' }}>
            TRANSACTION PIPELINE
          </div>
          {waterfallData.stages.map((stage: any, index: number) => (
            <div key={index} className="waterfall-stage">
              <div className="waterfall-stage-name">{stage.name}</div>
              <div className="waterfall-stage-metrics">
                <div className="waterfall-metric">
                  <span>IN:</span>
                  <span className="waterfall-metric-value">{stage.in?.toLocaleString()}</span>
                </div>
                {stage.drop > 0 && (
                  <div className="waterfall-metric">
                    <span>DROP:</span>
                    <span className="waterfall-metric-value" style={{ color: 'var(--fd-status-warning)' }}>
                      {stage.drop?.toLocaleString()}
                    </span>
                  </div>
                )}
                <div className="waterfall-metric">
                  <span>SUCCESS:</span>
                  <span className="waterfall-metric-value">{stage.success?.toLocaleString()}</span>
                </div>
                {stage.parallel_rate && (
                  <div className="waterfall-metric">
                    <span>PARALLEL:</span>
                    <span className="waterfall-metric-value">{stage.parallel_rate.toFixed(1)}%</span>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Real-time Logs */}
      <div className="firedancer-panel">
        <div className="firedancer-panel-header">SYSTEM LOG</div>
        <div className="firedancer-log">
          {logs.map((log, index) => (
            <span key={index} className="log-line">
              <span className="log-timestamp">[{log.timestamp}]</span>
              <span className={`log-level-${log.level}`}>{log.message}</span>
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}