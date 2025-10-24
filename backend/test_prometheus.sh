#!/bin/bash

# Test Prometheus metrics endpoint
echo "=== Testing Prometheus Metrics Endpoint ==="
echo ""

ENDPOINT="${PROMETHEUS_ENDPOINT:-http://127.0.0.1:8889/metrics}"

echo "Endpoint: $ENDPOINT"
echo ""

# Check if endpoint is reachable
if ! curl -sf "$ENDPOINT" > /dev/null; then
    echo "❌ Cannot reach Prometheus endpoint"
    exit 1
fi

echo "✅ Endpoint is reachable"
echo ""

# Get tx_commits metric
echo "=== Transaction Commits Metric ==="
TX_COMMITS=$(curl -s "$ENDPOINT" | grep "monad_execution_ledger_num_tx_commits{")
if [ -z "$TX_COMMITS" ]; then
    echo "❌ monad_execution_ledger_num_tx_commits not found"
else
    echo "$TX_COMMITS"

    # Extract value
    VALUE=$(echo "$TX_COMMITS" | awk '{print $2}')
    echo ""
    echo "Current value: $VALUE transactions"
fi

echo ""

# Get blocks_committed metric
echo "=== Blocks Committed Metric ==="
BLOCKS=$(curl -s "$ENDPOINT" | grep "monad_execution_ledger_num_blocks_committed{")
if [ -z "$BLOCKS" ]; then
    echo "⚠️  monad_execution_ledger_num_blocks_committed not found"
else
    echo "$BLOCKS"
fi

echo ""

# Calculate TPS over 5 seconds
echo "=== Calculating TPS over 5 seconds ==="
VALUE1=$(curl -s "$ENDPOINT" | grep "monad_execution_ledger_num_tx_commits{" | awk '{print $2}')
echo "Initial commits: $VALUE1"
sleep 5
VALUE2=$(curl -s "$ENDPOINT" | grep "monad_execution_ledger_num_tx_commits{" | awk '{print $2}')
echo "Final commits: $VALUE2"

if [ ! -z "$VALUE1" ] && [ ! -z "$VALUE2" ]; then
    DIFF=$(echo "$VALUE2 - $VALUE1" | bc)
    TPS=$(echo "scale=2; $DIFF / 5" | bc)
    echo ""
    echo "📊 TPS: $TPS tx/s"
else
    echo "❌ Could not calculate TPS"
fi
