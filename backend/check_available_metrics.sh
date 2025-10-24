#!/bin/bash

# Check what metrics are actually available from Monad
echo "=== Checking Available Monad Metrics ==="
echo ""

ENDPOINT="${PROMETHEUS_ENDPOINT:-http://127.0.0.1:8889/metrics}"

echo "Endpoint: $ENDPOINT"
echo ""

# Get all monad metrics
echo "=== All Monad Metrics ==="
curl -s "$ENDPOINT" | grep "^monad_" | grep -v "^#" | sort | head -50

echo ""
echo "=== TxPool Related Metrics ==="
curl -s "$ENDPOINT" | grep "monad.*tx" | grep -v "^#"

echo ""
echo "=== Execution Related Metrics ==="
curl -s "$ENDPOINT" | grep "monad.*execution" | grep -v "^#"

echo ""
echo "=== Pool/Drop Related Metrics ==="
curl -s "$ENDPOINT" | grep -E "monad.*(pool|drop|insert)" | grep -v "^#"

echo ""
echo "=== State Related Metrics ==="
curl -s "$ENDPOINT" | grep -E "monad.*(state|read|write)" | grep -v "^#"
