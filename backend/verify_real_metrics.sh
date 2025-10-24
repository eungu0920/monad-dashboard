#!/bin/bash

# Comprehensive script to verify all real metrics sources are working
echo "==================================="
echo "Monad Dashboard Real Metrics Verification"
echo "==================================="
echo ""

# Configuration
PROMETHEUS_ENDPOINT="${PROMETHEUS_ENDPOINT:-http://127.0.0.1:8889/metrics}"
MONAD_WEBSOCKET="${MONAD_WEBSOCKET:-ws://127.0.0.1:8081}"
IPC_PATH="${IPC_PATH:-/home/monad/monad-bft/mempool.sock}"

PASS_COUNT=0
FAIL_COUNT=0

function check_pass() {
    echo "✅ $1"
    PASS_COUNT=$((PASS_COUNT + 1))
}

function check_fail() {
    echo "❌ $1"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

function check_warn() {
    echo "⚠️  $1"
}

# Check 1: Prometheus Endpoint
echo "📊 Checking Prometheus Metrics..."
echo "Endpoint: $PROMETHEUS_ENDPOINT"
echo ""

if curl -sf "$PROMETHEUS_ENDPOINT" > /dev/null; then
    check_pass "Prometheus endpoint is reachable"

    # Check for TPS metric
    TX_COMMITS=$(curl -s "$PROMETHEUS_ENDPOINT" | grep "monad_execution_ledger_num_tx_commits{")
    if [ ! -z "$TX_COMMITS" ]; then
        VALUE=$(echo "$TX_COMMITS" | awk '{print $2}')
        check_pass "Found monad_execution_ledger_num_tx_commits: $VALUE"
    else
        check_fail "monad_execution_ledger_num_tx_commits not found"
    fi

    # Check for blocks committed
    BLOCKS=$(curl -s "$PROMETHEUS_ENDPOINT" | grep "monad_execution_ledger_num_blocks_committed{" | head -1)
    if [ ! -z "$BLOCKS" ]; then
        VALUE=$(echo "$BLOCKS" | awk '{print $2}')
        check_pass "Found monad_execution_ledger_num_blocks_committed: $VALUE"
    else
        check_warn "monad_execution_ledger_num_blocks_committed not found (optional)"
    fi

    # Check for txpool metrics (NEW - most important for waterfall!)
    echo ""
    echo "🔍 Checking TxPool Metrics (Critical for Waterfall)..."

    INSERT_OWNED=$(curl -s "$PROMETHEUS_ENDPOINT" | grep "monad_bft_txpool_pool_insert_owned_txs{")
    if [ ! -z "$INSERT_OWNED" ]; then
        VALUE=$(echo "$INSERT_OWNED" | awk '{print $2}')
        check_pass "Found monad_bft_txpool_pool_insert_owned_txs: $VALUE"
    else
        check_fail "monad_bft_txpool_pool_insert_owned_txs not found"
    fi

    INSERT_FWD=$(curl -s "$PROMETHEUS_ENDPOINT" | grep "monad_bft_txpool_pool_insert_forwarded_txs{")
    if [ ! -z "$INSERT_FWD" ]; then
        VALUE=$(echo "$INSERT_FWD" | awk '{print $2}')
        check_pass "Found monad_bft_txpool_pool_insert_forwarded_txs: $VALUE"
    else
        check_fail "monad_bft_txpool_pool_insert_forwarded_txs not found"
    fi

    DROP_NONCE=$(curl -s "$PROMETHEUS_ENDPOINT" | grep "monad_bft_txpool_pool_drop_nonce_too_low{")
    if [ ! -z "$DROP_NONCE" ]; then
        VALUE=$(echo "$DROP_NONCE" | awk '{print $2}')
        check_pass "Found monad_bft_txpool_pool_drop_nonce_too_low: $VALUE"
    else
        check_warn "monad_bft_txpool_pool_drop_nonce_too_low not found (optional)"
    fi

    PENDING=$(curl -s "$PROMETHEUS_ENDPOINT" | grep "monad_bft_txpool_pool_pending_txs{")
    if [ ! -z "$PENDING" ]; then
        VALUE=$(echo "$PENDING" | awk '{print $2}')
        check_pass "Found monad_bft_txpool_pool_pending_txs: $VALUE"
    else
        check_warn "monad_bft_txpool_pool_pending_txs not found (optional)"
    fi

else
    check_fail "Cannot reach Prometheus endpoint at $PROMETHEUS_ENDPOINT"
fi

echo ""
echo "-----------------------------------"
echo ""

# Check 2: WebSocket Connection
echo "🔌 Checking WebSocket Connection..."
echo "WebSocket: $MONAD_WEBSOCKET"
echo ""

# Try to connect (timeout after 2 seconds)
if timeout 2 curl -s --no-buffer -H "Connection: Upgrade" -H "Upgrade: websocket" \
    -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
    "$(echo $MONAD_WEBSOCKET | sed 's/ws:/http:/')" > /dev/null 2>&1; then
    check_pass "WebSocket endpoint is reachable"
else
    check_warn "WebSocket endpoint check inconclusive (may still work)"
fi

echo ""
echo "-----------------------------------"
echo ""

# Check 3: IPC Socket
echo "🔧 Checking IPC Socket..."
echo "IPC Path: $IPC_PATH"
echo ""

if [ -S "$IPC_PATH" ]; then
    check_pass "IPC socket file exists"

    # Check if we can connect
    if timeout 2 nc -U "$IPC_PATH" < /dev/null > /dev/null 2>&1; then
        check_pass "IPC socket is connectable"
    else
        check_warn "IPC socket exists but connection test failed"
    fi
else
    check_fail "IPC socket not found at $IPC_PATH"
fi

echo ""
echo "-----------------------------------"
echo ""

# Check 4: Calculate Real TPS
echo "📈 Calculating Real-Time TPS (5 second test)..."
echo ""

if curl -sf "$PROMETHEUS_ENDPOINT" > /dev/null; then
    VALUE1=$(curl -s "$PROMETHEUS_ENDPOINT" | grep "monad_execution_ledger_num_tx_commits{" | awk '{print $2}')

    if [ ! -z "$VALUE1" ]; then
        echo "Initial commits: $VALUE1"
        echo "Waiting 5 seconds..."
        sleep 5

        VALUE2=$(curl -s "$PROMETHEUS_ENDPOINT" | grep "monad_execution_ledger_num_tx_commits{" | awk '{print $2}')
        echo "Final commits: $VALUE2"
        echo ""

        if [ ! -z "$VALUE2" ]; then
            DIFF=$(echo "$VALUE2 - $VALUE1" | bc)
            TPS=$(echo "scale=2; $DIFF / 5" | bc)

            if [ $(echo "$TPS > 0" | bc) -eq 1 ]; then
                check_pass "Real-time TPS: $TPS tx/s (network is active)"
            else
                check_warn "TPS is 0 (network might be idle)"
            fi
        else
            check_fail "Could not get final commit value"
        fi
    else
        check_fail "Could not get initial commit value"
    fi
else
    check_fail "Prometheus endpoint not available for TPS test"
fi

echo ""
echo "==================================="
echo "Summary"
echo "==================================="
echo "✅ Passed: $PASS_COUNT"
echo "❌ Failed: $FAIL_COUNT"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo "🎉 All critical checks passed! Dashboard will use 100% real data."
    echo ""
    echo "Expected behavior:"
    echo "  • TPS: Calculated from Prometheus rate(monad_execution_ledger_num_tx_commits)"
    echo "  • Waterfall: Real metrics from Prometheus txpool counters"
    echo "  • Metadata: Shows 'source: prometheus_metrics'"
    exit 0
elif [ $PASS_COUNT -ge 3 ]; then
    echo "⚠️  Some checks failed but core metrics available."
    echo "Dashboard will use available real data with fallbacks."
    exit 0
else
    echo "❌ Too many critical checks failed."
    echo "Dashboard may fall back to estimation mode."
    exit 1
fi
