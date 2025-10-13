import React, { useState, useEffect } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';

interface EpochData {
  epoch: number;
  start_slot: number;
  end_slot: number;
  my_slots: number;
  skipped_slots: number;
  leaders: Array<{
    validator: string;
    slots: number[];
  }>;
}

export function LeaderSchedule() {
  const { metrics } = useWebSocket();
  const [epochData, setEpochData] = useState<EpochData | null>(null);
  const [selectedEpoch, setSelectedEpoch] = useState<number>(0);

  useEffect(() => {
    // Mock epoch data - in real implementation, fetch from API
    if (metrics) {
      const mockEpoch: EpochData = {
        epoch: Math.floor(metrics.consensus.current_height / 432000),
        start_slot: Math.floor(metrics.consensus.current_height / 432000) * 432000,
        end_slot: (Math.floor(metrics.consensus.current_height / 432000) + 1) * 432000,
        my_slots: 42,
        skipped_slots: 3,
        leaders: Array.from({ length: 20 }, (_, i) => ({
          validator: `Validator${i + 1}`,
          slots: Array.from({ length: Math.floor(Math.random() * 100) + 10 }, (_, j) => j * 10 + i)
        }))
      };
      setEpochData(mockEpoch);
      setSelectedEpoch(mockEpoch.epoch);
    }
  }, [metrics]);

  if (!metrics || !epochData) {
    return (
      <div className="firedancer-loading">
        <div className="loading-spinner"></div>
        <p>Loading leader schedule...</p>
      </div>
    );
  }

  return (
    <div className="leader-schedule-page">
      <div className="firedancer-panel epoch-panel">
        <div className="panel-header">Epoch Information</div>
        <div className="panel-content">
          <div className="epoch-selector">
            <button
              onClick={() => setSelectedEpoch(selectedEpoch - 1)}
              className="epoch-button"
            >
              ← Previous
            </button>
            <div className="epoch-current">
              <span className="epoch-label">Current Epoch</span>
              <span className="epoch-number">{selectedEpoch}</span>
            </div>
            <button
              onClick={() => setSelectedEpoch(selectedEpoch + 1)}
              className="epoch-button"
            >
              Next →
            </button>
          </div>

          <div className="epoch-stats">
            <div className="epoch-stat">
              <div className="status-label">Start Slot</div>
              <div className="status-value">{epochData.start_slot.toLocaleString()}</div>
            </div>
            <div className="epoch-stat">
              <div className="status-label">End Slot</div>
              <div className="status-value">{epochData.end_slot.toLocaleString()}</div>
            </div>
            <div className="epoch-stat">
              <div className="status-label">My Leader Slots</div>
              <div className="status-value highlight">{epochData.my_slots}</div>
            </div>
            <div className="epoch-stat">
              <div className="status-label">Skipped Slots</div>
              <div className="status-value error">{epochData.skipped_slots}</div>
            </div>
          </div>
        </div>
      </div>

      <div className="firedancer-panel leaders-panel">
        <div className="panel-header">Leader Schedule</div>
        <div className="panel-content">
          <div className="leaders-table-container">
            <table className="leaders-table">
              <thead>
                <tr>
                  <th>Validator</th>
                  <th>Total Slots</th>
                  <th>Slots</th>
                </tr>
              </thead>
              <tbody>
                {epochData.leaders.map((leader, index) => (
                  <tr key={index} className={leader.validator.includes('1') ? 'my-validator' : ''}>
                    <td className="validator-name">{leader.validator}</td>
                    <td className="validator-count">{leader.slots.length}</td>
                    <td className="validator-slots">
                      <div className="slots-preview">
                        {leader.slots.slice(0, 10).map((slot, i) => (
                          <span key={i} className="slot-badge">{slot}</span>
                        ))}
                        {leader.slots.length > 10 && (
                          <span className="slot-more">+{leader.slots.length - 10} more</span>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div className="firedancer-panel slots-timeline-panel">
        <div className="panel-header">Slots Timeline</div>
        <div className="panel-content">
          <div className="slots-timeline">
            {Array.from({ length: 100 }, (_, i) => {
              const slot = epochData.start_slot + i * 100;
              const isMySlot = i % 7 === 0;
              const isSkipped = i % 23 === 0;
              return (
                <div
                  key={i}
                  className={`slot-block ${isMySlot ? 'my-slot' : ''} ${isSkipped ? 'skipped' : ''}`}
                  title={`Slot ${slot}`}
                ></div>
              );
            })}
          </div>
          <div className="timeline-legend">
            <div className="legend-item">
              <div className="legend-block my-slot"></div>
              <span>My Slots</span>
            </div>
            <div className="legend-item">
              <div className="legend-block skipped"></div>
              <span>Skipped</span>
            </div>
            <div className="legend-item">
              <div className="legend-block"></div>
              <span>Other Leaders</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
