package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PrometheusCollector collects metrics from Monad's Prometheus/OTEL endpoint
type PrometheusCollector struct {
	endpoint   string
	httpClient *http.Client
	mu         sync.RWMutex

	// Real metrics from Prometheus
	metrics *PrometheusMetrics
}

// PrometheusMetrics contains parsed Prometheus metrics
type PrometheusMetrics struct {
	// Transaction commit metrics
	TxCommits       float64 // monad_execution_ledger_num_tx_commits
	TxCommitsTotal  float64 // Total counter value (cumulative)

	// TPS calculation (rate over 60s)
	TPS60s          float64 // Calculated from rate

	// Other execution metrics
	BlocksCommitted float64 // monad_execution_ledger_num_blocks_committed

	// TxPool metrics - CUMULATIVE counters from Prometheus
	InsertOwnedTxsTotal       float64 // monad_bft_txpool_pool_insert_owned_txs (cumulative)
	InsertForwardedTxsTotal   float64 // monad_bft_txpool_pool_insert_forwarded_txs (cumulative)
	DropInvalidSignatureTotal float64 // monad_bft_txpool_pool_drop_not_well_formed (cumulative)
	DropNonceTooLowTotal      float64 // monad_bft_txpool_pool_drop_nonce_too_low (cumulative)
	DropFeeTooLowTotal        float64 // monad_bft_txpool_pool_drop_fee_too_low (cumulative)
	DropInsufficientBalanceTotal float64 // monad_bft_txpool_pool_drop_insufficient_balance (cumulative)
	DropPoolFullTotal         float64 // monad_bft_txpool_pool_drop_pool_full (cumulative)
	PendingTxs                float64 // monad_bft_txpool_pool_pending_txs (gauge, not cumulative)
	TrackedTxs                float64 // monad_bft_txpool_pool_tracked_txs (gauge, not cumulative)

	// TxPool metrics - RATE (change per collection interval)
	InsertOwnedTxsRate       float64 // Rate of RPC transactions
	InsertForwardedTxsRate   float64 // Rate of P2P transactions
	DropInvalidSignatureRate float64 // Rate of signature failures
	DropNonceTooLowRate      float64 // Rate of nonce failures
	DropFeeTooLowRate        float64 // Rate of fee failures
	DropInsufficientBalanceRate float64 // Rate of balance failures
	DropPoolFullRate         float64 // Rate of pool full drops

	// Timestamps
	LastUpdated     time.Time
	LastUpdateTime  time.Time
}

// NewPrometheusCollector creates a new Prometheus metrics collector
func NewPrometheusCollector(endpoint string) *PrometheusCollector {
	return &PrometheusCollector{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		metrics: &PrometheusMetrics{
			LastUpdated:    time.Now(),
			LastUpdateTime: time.Now(),
		},
	}
}

// Start begins collecting metrics periodically
func (c *PrometheusCollector) Start() {
	// Collect immediately
	if err := c.collectMetrics(); err != nil {
		log.Printf("Initial Prometheus metrics collection failed: %v", err)
	}

	// Then collect every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			if err := c.collectMetrics(); err != nil {
				log.Printf("Prometheus metrics collection error: %v", err)
			}
		}
	}()
}

// collectMetrics fetches and parses Prometheus metrics
func (c *PrometheusCollector) collectMetrics() error {
	resp, err := c.httpClient.Get(c.endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return c.parseMetrics(resp.Body)
}

// parseMetrics parses Prometheus text format
func (c *PrometheusCollector) parseMetrics(body io.Reader) error {
	scanner := bufio.NewScanner(body)

	newMetrics := &PrometheusMetrics{
		LastUpdated: time.Now(),
	}

	// Keep previous values for rate calculation
	c.mu.RLock()
	prevMetrics := *c.metrics // Copy all previous values
	prevTime := c.metrics.LastUpdateTime
	c.mu.RUnlock()

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Parse metric line: metric_name{labels} value timestamp
		// Example: monad_execution_ledger_num_tx_commits{job="testnet"} 863080221 1761214210873
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		// Extract metric name (before '{' or space)
		metricNameFull := parts[0]
		metricName := metricNameFull
		if idx := strings.Index(metricNameFull, "{"); idx > 0 {
			metricName = metricNameFull[:idx]
		}

		// Parse value (second field)
		valueStr := parts[1]
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}

		// Extract relevant metrics (CUMULATIVE values)
		switch metricName {
		case "monad_execution_ledger_num_tx_commits":
			newMetrics.TxCommitsTotal = value
		case "monad_execution_ledger_num_blocks_committed":
			newMetrics.BlocksCommitted = value
		case "monad_execution_ledger_num_commits":
			newMetrics.BlocksCommitted = value // Alternative metric name

		// TxPool metrics (actual Monad metric names with "pool_" prefix)
		case "monad_bft_txpool_pool_insert_owned_txs":
			newMetrics.InsertOwnedTxsTotal = value
		case "monad_bft_txpool_pool_insert_forwarded_txs":
			newMetrics.InsertForwardedTxsTotal = value
		case "monad_bft_txpool_pool_drop_not_well_formed":
			newMetrics.DropInvalidSignatureTotal = value
		case "monad_bft_txpool_pool_drop_nonce_too_low":
			newMetrics.DropNonceTooLowTotal = value
		case "monad_bft_txpool_pool_drop_fee_too_low":
			newMetrics.DropFeeTooLowTotal = value
		case "monad_bft_txpool_pool_drop_insufficient_balance":
			newMetrics.DropInsufficientBalanceTotal = value
		case "monad_bft_txpool_pool_drop_pool_full":
			newMetrics.DropPoolFullTotal = value
		case "monad_bft_txpool_pool_pending_txs":
			newMetrics.PendingTxs = value // Gauge, not cumulative
		case "monad_bft_txpool_pool_tracked_txs":
			newMetrics.TrackedTxs = value // Gauge, not cumulative
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading metrics: %w", err)
	}

	// Calculate rates for ALL counters
	now := time.Now()
	timeDiff := now.Sub(prevTime).Seconds()

	if timeDiff > 0 && prevMetrics.TxCommitsTotal > 0 {
		// TPS calculation
		txDiff := newMetrics.TxCommitsTotal - prevMetrics.TxCommitsTotal
		newMetrics.TPS60s = txDiff / timeDiff

		// TxPool rates (change since last collection)
		newMetrics.InsertOwnedTxsRate = (newMetrics.InsertOwnedTxsTotal - prevMetrics.InsertOwnedTxsTotal) / timeDiff
		newMetrics.InsertForwardedTxsRate = (newMetrics.InsertForwardedTxsTotal - prevMetrics.InsertForwardedTxsTotal) / timeDiff
		newMetrics.DropInvalidSignatureRate = (newMetrics.DropInvalidSignatureTotal - prevMetrics.DropInvalidSignatureTotal) / timeDiff
		newMetrics.DropNonceTooLowRate = (newMetrics.DropNonceTooLowTotal - prevMetrics.DropNonceTooLowTotal) / timeDiff
		newMetrics.DropFeeTooLowRate = (newMetrics.DropFeeTooLowTotal - prevMetrics.DropFeeTooLowTotal) / timeDiff
		newMetrics.DropInsufficientBalanceRate = (newMetrics.DropInsufficientBalanceTotal - prevMetrics.DropInsufficientBalanceTotal) / timeDiff
		newMetrics.DropPoolFullRate = (newMetrics.DropPoolFullTotal - prevMetrics.DropPoolFullTotal) / timeDiff

		log.Printf("📊 Prometheus TPS: %.2f tx/s (commits: %.0f -> %.0f, diff: %.0f over %.1fs)",
			newMetrics.TPS60s, prevMetrics.TxCommitsTotal, newMetrics.TxCommitsTotal, txDiff, timeDiff)
	} else if newMetrics.TxCommitsTotal > 0 && prevMetrics.TxCommitsTotal == 0 {
		// First collection
		log.Printf("📊 Prometheus: Initial collection - tx_commits: %.0f, insert_owned: %.0f, insert_forwarded: %.0f",
			newMetrics.TxCommitsTotal, newMetrics.InsertOwnedTxsTotal, newMetrics.InsertForwardedTxsTotal)
	} else if newMetrics.TxCommitsTotal == 0 {
		log.Printf("⚠️  Prometheus: monad_execution_ledger_num_tx_commits not found in metrics")
	}

	newMetrics.LastUpdateTime = now

	// Update stored metrics
	c.mu.Lock()
	c.metrics = newMetrics
	c.mu.Unlock()

	return nil
}

// GetMetrics returns current metrics snapshot
func (c *PrometheusCollector) GetMetrics() *PrometheusMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy
	metricsCopy := *c.metrics
	return &metricsCopy
}

// GetTPS returns the current TPS calculated from Prometheus metrics
func (c *PrometheusCollector) GetTPS() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metrics.TPS60s
}

// IsHealthy checks if metrics are recent
func (c *PrometheusCollector) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Metrics should be updated within last 10 seconds
	return time.Since(c.metrics.LastUpdated) < 10*time.Second
}

// Global Prometheus collector instance
var (
	prometheusCollector   *PrometheusCollector
	prometheusCollectorMu sync.RWMutex
)

// InitializePrometheusCollector initializes the Prometheus metrics collector
func InitializePrometheusCollector(endpoint string) error {
	prometheusCollectorMu.Lock()
	defer prometheusCollectorMu.Unlock()

	prometheusCollector = NewPrometheusCollector(endpoint)

	// Test connection
	if err := prometheusCollector.collectMetrics(); err != nil {
		return fmt.Errorf("failed to connect to Prometheus endpoint: %w", err)
	}

	// Start background collection
	prometheusCollector.Start()

	log.Printf("✅ Prometheus collector initialized at %s", endpoint)
	return nil
}

// GetPrometheusCollector returns the global Prometheus collector
func GetPrometheusCollector() *PrometheusCollector {
	prometheusCollectorMu.RLock()
	defer prometheusCollectorMu.RUnlock()
	return prometheusCollector
}
