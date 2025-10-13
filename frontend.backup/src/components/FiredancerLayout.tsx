import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useWebSocket } from '../hooks/useWebSocket';
import '../styles/firedancer-authentic.css';

interface FiredancerLayoutProps {
  children: React.ReactNode;
}

export function FiredancerLayout({ children }: FiredancerLayoutProps) {
  const { metrics, connected } = useWebSocket();
  const location = useLocation();

  const formatNumber = (num: number): string => {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num?.toString() || '0';
  };

  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${mins}m`;
    return `${mins}m`;
  };

  const isActive = (path: string) => location.pathname === path;

  return (
    <div className="firedancer-authentic">
      {/* Navigation Bar */}
      <div className="firedancer-navbar">
        <div className="firedancer-logo">
          <div className="logo-icon">M</div>
          <span className="logo-text">monad</span>
          {metrics && (
            <span className="network-badge">v{metrics.node_info.version}</span>
          )}
        </div>

        <div className="navbar-center">
          <Link to="/" className={`nav-item ${isActive('/') ? 'active' : ''}`}>
            Overview
          </Link>
          <Link to="/leader-schedule" className={`nav-item ${isActive('/leader-schedule') ? 'active' : ''}`}>
            Leader Schedule
          </Link>
          <Link to="/gossip" className={`nav-item ${isActive('/gossip') ? 'active' : ''}`}>
            Gossip
          </Link>
        </div>

        <div className="navbar-right">
          {metrics ? (
            <>
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
            </>
          ) : (
            <span>Connecting...</span>
          )}
        </div>
      </div>

      {/* Main Content */}
      <div className="firedancer-content">
        {children}
      </div>
    </div>
  );
}
