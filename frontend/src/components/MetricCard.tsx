import React from 'react';

interface MetricCardProps {
  title: string;
  value: string | number;
  unit?: string;
  change?: number;
  changeLabel?: string;
  status?: 'good' | 'warning' | 'danger';
}

export function MetricCard({
  title,
  value,
  unit,
  change,
  changeLabel,
  status = 'good'
}: MetricCardProps) {
  const formatValue = (val: string | number): string => {
    if (typeof val === 'number') {
      if (val >= 1000000000) {
        return (val / 1000000000).toFixed(2) + 'B';
      } else if (val >= 1000000) {
        return (val / 1000000).toFixed(2) + 'M';
      } else if (val >= 1000) {
        return (val / 1000).toFixed(1) + 'K';
      }
      return val.toFixed(0);
    }
    return val.toString();
  };

  const getStatusColor = () => {
    switch (status) {
      case 'warning': return 'var(--monad-warning)';
      case 'danger': return 'var(--monad-danger)';
      default: return 'var(--monad-accent)';
    }
  };

  return (
    <div className="metric-card">
      <h3>{title}</h3>
      <div className="metric-value" style={{ color: getStatusColor() }}>
        {formatValue(value)}
        {unit && <span className="metric-unit">{unit}</span>}
      </div>
      {change !== undefined && (
        <div
          style={{
            fontSize: '0.75rem',
            color: change >= 0 ? 'var(--monad-accent)' : 'var(--monad-danger)',
            display: 'flex',
            alignItems: 'center',
            gap: '0.25rem'
          }}
        >
          <span>{change >= 0 ? '▲' : '▼'}</span>
          <span>{Math.abs(change).toFixed(1)}%</span>
          {changeLabel && <span style={{ color: 'var(--monad-text-muted)' }}>
            {changeLabel}
          </span>}
        </div>
      )}
    </div>
  );
}