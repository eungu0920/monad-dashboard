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
	TxCommitsTotal  float64 // Total counter value

	// TPS calculation (rate over 60s)
	TPS60s          float64 // Calculated from rate

	// Other execution metrics
	BlocksCommitted float64 // monad_execution_ledger_num_blocks_committed

	// Timestamps
	LastUpdated     time.Time
	LastTxCommits   float64 // Previous value for rate calculation
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
	prevTxCommits := c.metrics.TxCommitsTotal
	prevTime := c.metrics.LastUpdateTime
	c.mu.RUnlock()

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Parse metric line: metric_name{labels} value timestamp
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		metricName := strings.Split(parts[0], "{")[0]
		valueStr := parts[1]
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}

		// Extract relevant metrics
		switch metricName {
		case "monad_execution_ledger_num_tx_commits":
			newMetrics.TxCommitsTotal = value
		case "monad_execution_ledger_num_blocks_committed":
			newMetrics.BlocksCommitted = value
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading metrics: %w", err)
	}

	// Calculate TPS as rate over time
	now := time.Now()
	timeDiff := now.Sub(prevTime).Seconds()

	if timeDiff > 0 && newMetrics.TxCommitsTotal > 0 {
		txDiff := newMetrics.TxCommitsTotal - prevTxCommits
		newMetrics.TPS60s = txDiff / timeDiff

		log.Printf("📊 Prometheus TPS: %.2f tx/s (commits: %.0f -> %.0f over %.1fs)",
			newMetrics.TPS60s, prevTxCommits, newMetrics.TxCommitsTotal, timeDiff)
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
