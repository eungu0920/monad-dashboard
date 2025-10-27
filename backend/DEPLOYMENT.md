# Monad Dashboard Deployment Guide

## Recent Changes (Waterfall Redesign)

**Commit**: `feat: Redesign waterfall to match Monad transaction lifecycle`

### What Changed

1. **New Files**:
   - `consensus_tracker.go` - MonadBFT consensus phase tracker
   - `waterfall_metrics_v2.go` - New Monad lifecycle-aligned waterfall
   - `WATERFALL_REDESIGN.md` - Complete technical specification

2. **Modified Files**:
   - `main.go` - Added new API endpoints and consensus tracker initialization
   - `firedancer_protocol.go` - New WebSocket messages for v2 waterfall and consensus
   - `monad_subscriber.go` - Integration with consensus tracker
   - `metrics.go` - New API handlers

---

## Server Deployment Steps

### 1. Pull Latest Changes

```bash
cd ~/monad-dashboard
git pull origin main
```

Expected output:
```
remote: Enumerating objects: 15, done.
remote: Counting objects: 100% (15/15), done.
Updating abc1234..6186aaf
 backend/WATERFALL_REDESIGN.md   | 398 +++++++++++
 backend/consensus_tracker.go    | 268 +++++++
 backend/waterfall_metrics_v2.go | 497 +++++++++++++
 ...
 10 files changed, 1286 insertions(+), 35 deletions(-)
```

---

### 2. Build Backend

```bash
cd ~/monad-dashboard/backend
go mod tidy
go build -o monad-dashboard .
```

Expected:
- No errors
- Binary created: `monad-dashboard`

Check build:
```bash
ls -lh monad-dashboard
# Should show ~20-30MB binary
```

---

### 3. Stop Current Dashboard (if running)

Find running process:
```bash
ps aux | grep monad-dashboard
```

Stop gracefully:
```bash
# If using systemd
sudo systemctl stop monad-dashboard

# OR if running directly
pkill monad-dashboard
```

---

### 4. Start New Dashboard

```bash
cd ~/monad-dashboard/backend
./monad-dashboard
```

Expected startup logs:
```
✅ MonadBFT Consensus Tracker initialized
Attempting to connect to Prometheus endpoint at http://127.0.0.1:8889/metrics...
✅ Prometheus collector initialized - using accurate TPS from monad_execution_ledger_num_tx_commits
Attempting to connect to Monad IPC at /home/monad/monad-bft/mempool.sock...
IPC metrics collector not available: ...
Attempting to connect to Monad WebSocket at ws://127.0.0.1:8081...
Successfully initialized real-time WebSocket subscription
Monad Dashboard starting on :4000
```

---

### 5. Verify New Endpoints

#### Check Health
```bash
curl http://localhost:4000/api/v1/health
```

Expected:
```json
{"status":"ok","timestamp":1234567890,"version":"0.1.0"}
```

#### Check New Waterfall V2
```bash
curl http://localhost:4000/api/v1/waterfall/v2 | jq '.metadata.source'
```

Expected:
```
"prometheus_metrics"  # or "block_estimation" or "mock_data"
```

#### Check Consensus State
```bash
curl http://localhost:4000/api/v1/consensus | jq
```

Expected:
```json
{
  "current_block": 12345,
  "finalized_block": 12343,
  "blocks_behind": 2,
  "proposed_blocks": 1,
  "voted_blocks": 1,
  "finalized_blocks": 8,
  "recent_blocks": [...]
}
```

---

### 6. Monitor Logs

Watch for new waterfall messages:
```bash
# In dashboard logs
grep "Monad Waterfall source" logs/dashboard.log

# Should see:
🌊 Monad Waterfall source: prometheus_metrics
```

Watch for consensus tracking:
```bash
# New blocks should trigger consensus updates
# Look for consensus state in WebSocket messages
```

---

## Verification Checklist

- [ ] Dashboard starts without errors
- [ ] `/api/v1/waterfall/v2` returns nodes/links format
- [ ] `/api/v1/consensus` returns consensus state with recent blocks
- [ ] WebSocket sends `monad_waterfall_v2` messages
- [ ] WebSocket sends `monad_consensus_state` messages
- [ ] Legacy `/api/v1/waterfall` still works (backward compat)
- [ ] Logs show "MonadBFT Consensus Tracker initialized"
- [ ] Consensus state updates as new blocks arrive

---

## WebSocket Messages (New)

The dashboard now sends additional WebSocket messages:

### 1. Monad Waterfall V2
```json
{
  "topic": "summary",
  "key": "monad_waterfall_v2",
  "value": {
    "nodes": [...],
    "links": [...],
    "metadata": {
      "source": "prometheus_metrics",
      "consensus_state": {...}
    }
  }
}
```

### 2. MonadBFT Consensus State
```json
{
  "topic": "summary",
  "key": "monad_consensus_state",
  "value": {
    "current_block": 12345,
    "finalized_block": 12343,
    "blocks_behind": 2,
    "recent_blocks": [
      {
        "block_number": 12345,
        "block_hash": "0x...",
        "phase": "proposed",
        "tx_count": 150
      },
      ...
    ]
  }
}
```

### 3. Legacy Waterfall (Still Sent)
```json
{
  "topic": "summary",
  "key": "live_txn_waterfall",
  "value": {
    "waterfall": {
      "in": {...},
      "out": {...}
    }
  }
}
```

---

## Troubleshooting

### Issue: Build Fails
```bash
# Clean and rebuild
go clean
rm -f monad-dashboard
go mod tidy
go build -o monad-dashboard .
```

### Issue: Consensus State Empty
**Cause**: No blocks have been received yet

**Solution**: Wait for at least 3 blocks (1-2 seconds)
```bash
# Check if blocks are being received
curl http://localhost:4000/api/v1/metrics | jq '.consensus.current_height'
```

### Issue: Waterfall Source is "mock_data"
**Cause**:
1. Prometheus not available
2. IPC not available
3. WebSocket not connected

**Check**:
```bash
# 1. Check Prometheus
curl http://127.0.0.1:8889/metrics | grep monad_bft_txpool

# 2. Check IPC socket
ls -l /home/monad/monad-bft/mempool.sock

# 3. Check WebSocket
curl http://127.0.0.1:8081
```

### Issue: "Too Many Open Files"
```bash
# Increase file descriptor limit
ulimit -n 65536

# Make permanent (add to /etc/security/limits.conf)
* soft nofile 65536
* hard nofile 65536
```

---

## Frontend Updates (TODO)

The backend is now ready. Frontend needs:

1. **Subscribe to new WebSocket messages**:
   - Listen for `monad_waterfall_v2`
   - Listen for `monad_consensus_state`

2. **Create MonadBFT Component**:
   - Display last 10 blocks with phases
   - Show progress bars (Proposed 33%, Voted 66%, Finalized 100%)
   - Update in real-time

3. **Update Sankey Diagram**:
   - Accept nodes/links format instead of in/out
   - Show 7 Monad lifecycle stages
   - Color-code consensus phases

4. **Add Consensus Indicator**:
   - Show finality lag: "2 blocks behind"
   - Display current vs finalized block numbers

---

## Performance Notes

- **Consensus Tracker**: Tracks last 20 blocks (configurable)
- **Memory**: ~2-5 KB per block × 20 = ~100 KB
- **CPU**: Minimal (updates only on new blocks)
- **Network**: +2 WebSocket messages per update (200ms interval)

---

## Rollback (if needed)

If issues occur:

```bash
cd ~/monad-dashboard
git log --oneline -5

# Rollback to previous commit
git checkout <previous-commit-hash>
cd backend
go build -o monad-dashboard .
./monad-dashboard
```

---

## Next Release Plan

### Phase 2: Frontend Integration
- [ ] MonadBFT visualization component
- [ ] Sankey diagram update
- [ ] Consensus state display
- [ ] Testing with real node data

### Phase 3: Enhanced Metrics
- [ ] RaptorCast latency tracking
- [ ] Leader propagation visualization (N×K retries)
- [ ] Historical consensus metrics
- [ ] Alert on high finality lag

---

## Support

If you encounter issues:

1. Check logs: `tail -f logs/dashboard.log`
2. Verify endpoints: Use curl commands above
3. Check Prometheus: `curl http://127.0.0.1:8889/metrics`
4. Restart dashboard: `pkill monad-dashboard && ./monad-dashboard`

For questions about the redesign, see `WATERFALL_REDESIGN.md`.
