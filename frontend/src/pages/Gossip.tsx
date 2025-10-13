import React, { useState, useEffect } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';

interface GossipPeer {
  pubkey: string;
  ip: string;
  port: number;
  version: string;
  stake: number;
  rtt: number;
  last_seen: number;
}

export function Gossip() {
  const { metrics } = useWebSocket();
  const [peers, setPeers] = useState<GossipPeer[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState<'stake' | 'rtt' | 'last_seen'>('stake');

  useEffect(() => {
    // Mock gossip peers data - in real implementation, fetch from API
    if (metrics) {
      const mockPeers: GossipPeer[] = Array.from({ length: metrics.network.peer_count }, (_, i) => ({
        pubkey: `0x${Math.random().toString(16).substr(2, 40)}`,
        ip: `${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`,
        port: 30000 + Math.floor(Math.random() * 1000),
        version: `v1.${Math.floor(Math.random() * 5)}.${Math.floor(Math.random() * 10)}`,
        stake: Math.floor(Math.random() * 1000000) + 10000,
        rtt: Math.floor(Math.random() * 200) + 10,
        last_seen: Date.now() - Math.floor(Math.random() * 60000)
      }));
      setPeers(mockPeers);
    }
  }, [metrics]);

  const filteredPeers = peers
    .filter(peer =>
      peer.pubkey.toLowerCase().includes(searchTerm.toLowerCase()) ||
      peer.ip.includes(searchTerm) ||
      peer.version.includes(searchTerm)
    )
    .sort((a, b) => {
      if (sortBy === 'stake') return b.stake - a.stake;
      if (sortBy === 'rtt') return a.rtt - b.rtt;
      return b.last_seen - a.last_seen;
    });

  const formatStake = (stake: number): string => {
    if (stake >= 1000000) return (stake / 1000000).toFixed(2) + 'M';
    if (stake >= 1000) return (stake / 1000).toFixed(1) + 'K';
    return stake.toString();
  };

  const formatTimeSince = (timestamp: number): string => {
    const seconds = Math.floor((Date.now() - timestamp) / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    return `${Math.floor(seconds / 3600)}h ago`;
  };

  if (!metrics) {
    return (
      <div className="firedancer-loading">
        <div className="loading-spinner"></div>
        <p>Loading gossip network...</p>
      </div>
    );
  }

  return (
    <div className="gossip-page">
      <div className="firedancer-panel network-stats-panel">
        <div className="panel-header">Network Statistics</div>
        <div className="panel-content">
          <div className="network-stats-grid">
            <div className="network-stat">
              <div className="status-label">Total Peers</div>
              <div className="status-value highlight">{peers.length}</div>
            </div>
            <div className="network-stat">
              <div className="status-label">Active Validators</div>
              <div className="status-value">{metrics.consensus.validator_count}</div>
            </div>
            <div className="network-stat">
              <div className="status-label">Avg RTT</div>
              <div className="status-value">
                {peers.length > 0 ? Math.floor(peers.reduce((sum, p) => sum + p.rtt, 0) / peers.length) : 0}ms
              </div>
            </div>
            <div className="network-stat">
              <div className="status-label">Total Stake</div>
              <div className="status-value">
                {formatStake(peers.reduce((sum, p) => sum + p.stake, 0))} MONAD
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="firedancer-panel peers-panel">
        <div className="panel-header">
          Gossip Peers
          <div className="panel-controls">
            <input
              type="text"
              placeholder="Search peers..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="search-input"
            />
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as any)}
              className="sort-select"
            >
              <option value="stake">Sort by Stake</option>
              <option value="rtt">Sort by RTT</option>
              <option value="last_seen">Sort by Last Seen</option>
            </select>
          </div>
        </div>
        <div className="panel-content">
          <div className="peers-table-container">
            <table className="peers-table">
              <thead>
                <tr>
                  <th>Public Key</th>
                  <th>IP Address</th>
                  <th>Version</th>
                  <th>Stake</th>
                  <th>RTT</th>
                  <th>Last Seen</th>
                </tr>
              </thead>
              <tbody>
                {filteredPeers.map((peer, index) => (
                  <tr key={index}>
                    <td className="peer-pubkey" title={peer.pubkey}>
                      {peer.pubkey.substring(0, 10)}...{peer.pubkey.substring(peer.pubkey.length - 8)}
                    </td>
                    <td className="peer-ip">{peer.ip}:{peer.port}</td>
                    <td className="peer-version">
                      <span className="version-badge">{peer.version}</span>
                    </td>
                    <td className="peer-stake">{formatStake(peer.stake)} MONAD</td>
                    <td className={`peer-rtt ${peer.rtt < 50 ? 'good' : peer.rtt < 100 ? 'ok' : 'poor'}`}>
                      {peer.rtt}ms
                    </td>
                    <td className="peer-lastseen">{formatTimeSince(peer.last_seen)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div className="firedancer-panel network-map-panel">
        <div className="panel-header">Network Topology</div>
        <div className="panel-content">
          <div className="network-map">
            <svg width="100%" height="400" className="topology-svg">
              {/* Central node (this validator) */}
              <circle cx="50%" cy="50%" r="20" fill="#22d3ee" />
              <text x="50%" y="50%" textAnchor="middle" dy="5" fill="#0f1419" fontSize="12" fontWeight="600">
                ME
              </text>

              {/* Peer nodes */}
              {filteredPeers.slice(0, 12).map((peer, i) => {
                const angle = (i / 12) * 2 * Math.PI;
                const radius = 150;
                const x = `calc(50% + ${radius * Math.cos(angle)}px)`;
                const y = `calc(50% + ${radius * Math.sin(angle)}px)`;

                return (
                  <g key={i}>
                    <line
                      x1="50%"
                      y1="50%"
                      x2={x}
                      y2={y}
                      stroke="#374151"
                      strokeWidth="1"
                      opacity="0.3"
                    />
                    <circle
                      cx={x}
                      cy={y}
                      r="8"
                      fill={peer.rtt < 50 ? "#10b981" : peer.rtt < 100 ? "#f59e0b" : "#ef4444"}
                      opacity="0.8"
                    />
                  </g>
                );
              })}
            </svg>
          </div>
          <div className="network-legend">
            <div className="legend-item">
              <div className="legend-dot good"></div>
              <span>&lt; 50ms</span>
            </div>
            <div className="legend-item">
              <div className="legend-dot ok"></div>
              <span>50-100ms</span>
            </div>
            <div className="legend-item">
              <div className="legend-dot poor"></div>
              <span>&gt; 100ms</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
