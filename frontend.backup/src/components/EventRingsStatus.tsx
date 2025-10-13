import React, { useState, useEffect } from 'react';

interface EventRingsStats {
  connected: boolean;
  events_received: number;
  bytes_received: number;
  missed_events: number;
  parse_errors: number;
  last_sequence: number;
  buffer_size: number;
  message?: string;
}

export function EventRingsStatus() {
  const [stats, setStats] = useState<EventRingsStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch('/api/v1/event-rings');
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        setStats(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch event rings status');
      } finally {
        setLoading(false);
      }
    };

    // Initial fetch
    fetchStats();

    // Poll every 5 seconds
    const interval = setInterval(fetchStats, 5000);

    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="bg-gray-800 rounded-lg p-4">
        <h3 className="text-lg font-semibold text-gray-200 mb-2">Event Rings Status</h3>
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-gray-800 rounded-lg p-4">
        <h3 className="text-lg font-semibold text-gray-200 mb-2">Event Rings Status</h3>
        <div className="text-red-400">Error: {error}</div>
      </div>
    );
  }

  const formatNumber = (num: number): string => {
    return num.toLocaleString();
  };

  const formatBytes = (bytes: number): string => {
    const units = ['B', 'KB', 'MB', 'GB'];
    let size = bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }

    return `${size.toFixed(2)} ${units[unitIndex]}`;
  };

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <h3 className="text-lg font-semibold text-gray-200 mb-4">Event Rings Status</h3>

      {/* Connection Status */}
      <div className="mb-4">
        <div className="flex items-center gap-2 mb-2">
          <div className={`w-3 h-3 rounded-full ${stats?.connected ? 'bg-green-500' : 'bg-red-500'}`}></div>
          <span className="text-gray-200">
            {stats?.connected ? 'Connected to Event Rings' : 'Disconnected'}
          </span>
        </div>
        {stats?.message && (
          <div className="text-gray-400 text-sm ml-5">{stats.message}</div>
        )}
      </div>

      {stats?.connected && (
        <div className="grid grid-cols-2 gap-4">
          {/* Performance Metrics */}
          <div>
            <h4 className="text-sm font-medium text-gray-300 mb-2">Performance</h4>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-gray-400 text-sm">Events Received:</span>
                <span className="text-gray-200 text-sm">{formatNumber(stats.events_received)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400 text-sm">Data Received:</span>
                <span className="text-gray-200 text-sm">{formatBytes(stats.bytes_received)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400 text-sm">Buffer Size:</span>
                <span className="text-gray-200 text-sm">{stats.buffer_size}/1000</span>
              </div>
            </div>
          </div>

          {/* Error Metrics */}
          <div>
            <h4 className="text-sm font-medium text-gray-300 mb-2">Quality</h4>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-gray-400 text-sm">Missed Events:</span>
                <span className={`text-sm ${stats.missed_events > 0 ? 'text-yellow-400' : 'text-green-400'}`}>
                  {formatNumber(stats.missed_events)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400 text-sm">Parse Errors:</span>
                <span className={`text-sm ${stats.parse_errors > 0 ? 'text-red-400' : 'text-green-400'}`}>
                  {formatNumber(stats.parse_errors)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400 text-sm">Last Sequence:</span>
                <span className="text-gray-200 text-sm">#{stats.last_sequence}</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Real-time Event Stream Indicator */}
      {stats?.connected && (
        <div className="mt-4 pt-4 border-t border-gray-700">
          <div className="flex items-center justify-between">
            <span className="text-gray-400 text-sm">Real-time Events:</span>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
              <span className="text-green-400 text-sm">Streaming</span>
            </div>
          </div>
        </div>
      )}

      {/* Fallback Mode Indicator */}
      {!stats?.connected && (
        <div className="mt-4 p-3 bg-yellow-900/20 border border-yellow-600/30 rounded">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-yellow-400 rounded-full"></div>
            <span className="text-yellow-400 text-sm">Running in RPC-only mode</span>
          </div>
          <div className="text-gray-400 text-xs mt-1">
            Dashboard will use standard RPC calls for metrics collection
          </div>
        </div>
      )}
    </div>
  );
}