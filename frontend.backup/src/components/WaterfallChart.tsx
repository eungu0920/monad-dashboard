import React, { useEffect, useState } from 'react';
import type { WaterfallData, WaterfallStage } from '../types';

interface WaterfallChartProps {
  data?: WaterfallData | null;
}

export function WaterfallChart({ data }: WaterfallChartProps) {
  const [animatedValues, setAnimatedValues] = useState<number[]>([]);

  useEffect(() => {
    if (data?.stages) {
      // Animate stage values
      const newValues = data.stages.map(stage => stage.success);
      setAnimatedValues(newValues);
    }
  }, [data]);

  if (!data) {
    return (
      <div className="waterfall-container">
        <div className="loading">Loading waterfall data...</div>
      </div>
    );
  }

  const formatNumber = (num: number): string => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return num.toString();
  };

  const getStageColor = (stage: WaterfallStage, index: number): string => {
    // Color progression through the pipeline
    const colors = [
      '#6c5ce7', // Monad Primary
      '#a29bfe', // Monad Secondary
      '#74b9ff', // Light Blue
      '#00b894', // Monad Accent
      '#00cec9', // Teal
      '#55a3ff', // Blue
      '#fd79a8', // Pink
      '#fdcb6e', // Yellow
    ];

    return colors[index % colors.length];
  };

  const getIntensity = (value: number, maxValue: number): number => {
    return Math.max(0.3, Math.min(1, value / maxValue));
  };

  const maxValue = Math.max(...data.stages.map(s => s.success));

  return (
    <div className="waterfall-container">
      <div className="waterfall-header">
        <h2 className="waterfall-title">Transaction Waterfall</h2>
        <div className="waterfall-stats">
          <span>Total In: {formatNumber(data.summary.total_in)}</span>
          <span>Success: {formatNumber(data.summary.total_success)}</span>
          <span>Dropped: {formatNumber(data.summary.total_dropped)}</span>
          <span>Success Rate: {(data.summary.success_rate * 100).toFixed(1)}%</span>
        </div>
      </div>

      <div className="waterfall-flow">
        {data.stages.map((stage, index) => (
          <React.Fragment key={stage.name}>
            <div className="waterfall-stage">
              <div className="stage-name">{stage.name}</div>
              <div
                className="stage-box"
                style={{
                  background: `linear-gradient(135deg, ${getStageColor(stage, index)}, ${getStageColor(stage, index)}99)`,
                  opacity: getIntensity(stage.success, maxValue),
                  boxShadow: `0 4px 12px ${getStageColor(stage, index)}33`,
                }}
              >
                <div className="stage-throughput">
                  {formatNumber(stage.success)}
                </div>
                {stage.drop > 0 && (
                  <div className="stage-drops">
                    -{formatNumber(stage.drop)} dropped
                  </div>
                )}
                {stage.parallel_rate !== undefined && (
                  <div className="stage-drops">
                    {stage.parallel_rate.toFixed(1)}% parallel
                  </div>
                )}
              </div>
              <div style={{
                fontSize: '0.75rem',
                color: 'var(--monad-text-muted)',
                textAlign: 'center'
              }}>
                {stage.in > 0 && `In: ${formatNumber(stage.in)}`}
              </div>
            </div>

            {index < data.stages.length - 1 && (
              <div className="stage-arrow">→</div>
            )}
          </React.Fragment>
        ))}
      </div>

      {/* Detailed breakdown */}
      <div style={{
        marginTop: '2rem',
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
        gap: '1rem'
      }}>
        {data.stages.map((stage, index) => (
          <div
            key={stage.name}
            style={{
              padding: '1rem',
              background: 'rgba(255,255,255,0.05)',
              borderRadius: '8px',
              border: `1px solid ${getStageColor(stage, index)}33`
            }}
          >
            <h4 style={{
              color: getStageColor(stage, index),
              marginBottom: '0.5rem',
              fontSize: '0.875rem',
              fontWeight: '600'
            }}>
              {stage.name}
            </h4>
            <div style={{ fontSize: '0.75rem', color: 'var(--monad-text-muted)' }}>
              <div>Input: {formatNumber(stage.in)}</div>
              <div>Success: {formatNumber(stage.success)}</div>
              {stage.drop > 0 && <div style={{ color: 'var(--monad-danger)' }}>
                Dropped: {formatNumber(stage.drop)}
              </div>}
              {stage.parallel_rate !== undefined && (
                <div style={{ color: 'var(--monad-accent)' }}>
                  Parallel Rate: {stage.parallel_rate.toFixed(1)}%
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}