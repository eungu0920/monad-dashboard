import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import type { ExecutionMetrics } from '../types';

interface ThroughputChartProps {
  data: ExecutionMetrics[];
}

export function ThroughputChart({ data }: ThroughputChartProps) {
  if (!data || data.length === 0) {
    return (
      <div className="chart-container">
        <div className="loading">Loading throughput data...</div>
      </div>
    );
  }

  const chartData = data.map((metrics, index) => ({
    time: new Date(Date.now() - (data.length - index - 1) * 1000).toLocaleTimeString(),
    tps: Math.round(metrics.tps),
    parallel_rate: Math.round(metrics.parallel_success_rate * 100),
    pending: metrics.pending_tx_count,
  }));

  return (
    <div className="chart-container">
      <h3 className="chart-title">Transaction Throughput</h3>
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--monad-border)" />
          <XAxis
            dataKey="time"
            stroke="var(--monad-text-muted)"
            fontSize={12}
          />
          <YAxis
            stroke="var(--monad-text-muted)"
            fontSize={12}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'var(--monad-card)',
              border: '1px solid var(--monad-border)',
              borderRadius: '8px',
              color: 'var(--monad-text)'
            }}
          />
          <Line
            type="monotone"
            dataKey="tps"
            stroke="var(--monad-primary)"
            strokeWidth={2}
            dot={{ r: 3 }}
            name="TPS"
          />
          <Line
            type="monotone"
            dataKey="parallel_rate"
            stroke="var(--monad-accent)"
            strokeWidth={2}
            dot={{ r: 3 }}
            name="Parallel Success %"
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}