package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// TransactionLog represents a transaction log event from monadLogs
type TransactionLog struct {
	BlockNumber      int64    `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex int      `json:"transactionIndex"`
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	Timestamp        int64    `json:"timestamp"`
}

// MonadSubscriber handles real-time subscriptions to Monad node
type MonadSubscriber struct {
	wsURL            string
	conn             *websocket.Conn
	headsSubID       string // Subscription ID for monadNewHeads
	logsSubID        string // Subscription ID for monadLogs

	blockChan        chan *BlockHeader
	logsChan         chan *TransactionLog
	errorChan        chan error

	mu             sync.RWMutex
	latestBlock    *BlockHeader
	isConnected    bool

	// TPS calculation - track recent blocks
	recentBlocks    []BlockTxInfo
	maxRecentBlocks int

	// TPS history for charting
	tpsHistory      [][5]float64 // [total, vote, avg, instant, txCount]
	maxHistorySize  int

	ctx            context.Context
	cancel         context.CancelFunc
}

// BlockTxInfo stores transaction count and timestamp for TPS calculation
type BlockTxInfo struct {
	Timestamp    int64
	Transactions int
}

// BlockHeader represents a new block header
type BlockHeader struct {
	Number       int64  `json:"number"`
	Hash         string `json:"hash"`
	Timestamp    int64  `json:"timestamp"`
	Transactions int    `json:"transactionCount"`
	GasUsed      int64  `json:"gasUsed"`
}

// NewMonadSubscriber creates a new subscriber
func NewMonadSubscriber(wsURL string) *MonadSubscriber {
	ctx, cancel := context.WithCancel(context.Background())
	return &MonadSubscriber{
		wsURL:           wsURL,
		blockChan:       make(chan *BlockHeader, 100),
		logsChan:        make(chan *TransactionLog, 1000), // Larger buffer for logs
		errorChan:       make(chan error, 10),
		recentBlocks:    make([]BlockTxInfo, 0, 10),
		maxRecentBlocks: 10, // Track last 10 blocks (~4 seconds of data)
		tpsHistory:      make([][5]float64, 0, 200),
		maxHistorySize:  200, // Keep 200 data points for chart (80 seconds of data)
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Connect establishes WebSocket connection and subscribes to new blocks
func (s *MonadSubscriber) Connect() error {
	log.Printf("Connecting to Monad WebSocket at %s...", s.wsURL)

	conn, _, err := websocket.DefaultDialer.Dial(s.wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Monad WebSocket: %w", err)
	}

	s.conn = conn
	s.isConnected = true

	// Subscribe to newHeads (standard Ethereum block headers)
	headsSubMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "eth_subscribe",
		"params":  []interface{}{"newHeads"},
	}

	if err := conn.WriteJSON(headsSubMsg); err != nil {
		return fmt.Errorf("failed to send newHeads subscribe message: %w", err)
	}

	// Read newHeads subscription confirmation
	var headsSubResponse struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  string `json:"result"`
	}

	if err := conn.ReadJSON(&headsSubResponse); err != nil {
		return fmt.Errorf("failed to read newHeads subscription response: %w", err)
	}

	s.headsSubID = headsSubResponse.Result
	log.Printf("Successfully subscribed to newHeads with subscription ID: %s", s.headsSubID)

	// Subscribe to monadLogs (transaction logs for flow visualization)
	logsSubMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "eth_subscribe",
		"params":  []interface{}{"monadLogs", map[string]interface{}{}}, // Empty filter = all logs
	}

	if err := conn.WriteJSON(logsSubMsg); err != nil {
		return fmt.Errorf("failed to send monadLogs subscribe message: %w", err)
	}

	// Read monadLogs subscription confirmation
	var logsSubResponse struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  string `json:"result"`
	}

	if err := conn.ReadJSON(&logsSubResponse); err != nil {
		return fmt.Errorf("failed to read monadLogs subscription response: %w", err)
	}

	s.logsSubID = logsSubResponse.Result
	log.Printf("Successfully subscribed to monadLogs with subscription ID: %s", s.logsSubID)

	// Start listening for messages
	go s.listen()

	return nil
}

// listen continuously reads messages from WebSocket
func (s *MonadSubscriber) listen() {
	defer func() {
		s.mu.Lock()
		s.isConnected = false
		s.mu.Unlock()
	}()

	for {
		select {
		case <-s.ctx.Done():
			log.Println("Monad subscriber context cancelled, stopping listener")
			return
		default:
			var msg map[string]interface{}
			if err := s.conn.ReadJSON(&msg); err != nil {
				log.Printf("Error reading from Monad WebSocket: %v", err)
				s.errorChan <- err

				// Try to reconnect after error
				time.Sleep(2 * time.Second)
				if err := s.reconnect(); err != nil {
					log.Printf("Failed to reconnect: %v", err)
					continue
				}
				return
			}

			// Check if this is a subscription message
			if method, ok := msg["method"].(string); ok && method == "eth_subscription" {
				// Determine which subscription this is for
				params, ok := msg["params"].(map[string]interface{})
				if !ok {
					continue
				}

				subID, ok := params["subscription"].(string)
				if !ok {
					continue
				}

				// Route to appropriate handler
				if subID == s.headsSubID {
					s.handleBlockMessage(msg)
				} else if subID == s.logsSubID {
					s.handleLogsMessage(msg)
				}
			}
		}
	}
}

// handleBlockMessage processes incoming block headers from newHeads
func (s *MonadSubscriber) handleBlockMessage(msg map[string]interface{}) {
	params, ok := msg["params"].(map[string]interface{})
	if !ok {
		return
	}

	result, ok := params["result"].(map[string]interface{})
	if !ok {
		return
	}

	// Parse block header
	header := s.parseBlockHeader(result)
	if header == nil {
		return
	}

	// Update latest block
	s.mu.Lock()
	s.latestBlock = header
	s.mu.Unlock()

	// Fetch full block details to get transaction count, then send to channel
	// newHeads subscription doesn't include tx count, so we need to fetch it
	go func() {
		// Enrich with transaction count first
		s.enrichBlockWithTransactions(header)

		// Now send the enriched block to the channel for metrics update
		select {
		case s.blockChan <- header:
		default:
			// Channel full, skip this block
			log.Printf("Block channel full, skipping block %d", header.Number)
		}
	}()

	log.Printf("Received new block: height=%d, hash=%s (enriching...)",
		header.Number, header.Hash[:10])
}

// handleLogsMessage processes incoming transaction logs from monadLogs
func (s *MonadSubscriber) handleLogsMessage(msg map[string]interface{}) {
	params, ok := msg["params"].(map[string]interface{})
	if !ok {
		return
	}

	result, ok := params["result"].(map[string]interface{})
	if !ok {
		return
	}

	// Parse transaction log
	txLog := s.parseTransactionLog(result)
	if txLog == nil {
		return
	}

	// Send to logs channel
	select {
	case s.logsChan <- txLog:
	default:
		// Channel full, skip this log
		log.Printf("Logs channel full, skipping log for tx %s", txLog.TransactionHash[:10])
	}
}

// parseTransactionLog converts JSON to TransactionLog
func (s *MonadSubscriber) parseTransactionLog(result map[string]interface{}) *TransactionLog {
	blockNumberStr, ok := result["blockNumber"].(string)
	if !ok {
		return nil
	}

	blockNumber, err := parseHexToInt64(blockNumberStr)
	if err != nil {
		log.Printf("Failed to parse block number in log: %v", err)
		return nil
	}

	txHash, _ := result["transactionHash"].(string)
	address, _ := result["address"].(string)
	data, _ := result["data"].(string)

	// Parse transaction index
	txIndex := 0
	if txIndexStr, ok := result["transactionIndex"].(string); ok {
		if idx, err := parseHexToInt64(txIndexStr); err == nil {
			txIndex = int(idx)
		}
	}

	// Parse topics array
	topics := []string{}
	if topicsArr, ok := result["topics"].([]interface{}); ok {
		for _, t := range topicsArr {
			if topicStr, ok := t.(string); ok {
				topics = append(topics, topicStr)
			}
		}
	}

	return &TransactionLog{
		BlockNumber:      blockNumber,
		TransactionHash:  txHash,
		TransactionIndex: txIndex,
		Address:          address,
		Topics:           topics,
		Data:             data,
		Timestamp:        time.Now().Unix(), // Use current time as approximation
	}
}

// enrichBlockWithTransactions fetches full block details to get transaction count
func (s *MonadSubscriber) enrichBlockWithTransactions(header *BlockHeader) {
	// Use monadClient to fetch full block with transaction count
	blockResp, err := monadClient.rpcCall(monadClient.ExecutionRPCUrl, "eth_getBlockByNumber",
		[]interface{}{fmt.Sprintf("0x%x", header.Number), false})
	if err != nil {
		log.Printf("Failed to fetch block details for enrichment: %v", err)
		return
	}

	var block struct {
		Result struct {
			Transactions []string `json:"transactions"`
		} `json:"result"`
	}

	if err := json.Unmarshal(blockResp, &block); err != nil {
		log.Printf("Failed to decode block for enrichment: %v", err)
		return
	}

	// Update transaction count
	header.Transactions = len(block.Result.Transactions)

	// Add to recent blocks for TPS calculation
	s.addRecentBlock(header.Timestamp, header.Transactions)

	// Calculate TPS metrics for logging
	epoch := header.Number / 50000 // 50,000 blocks per epoch
	instantTPS := float64(header.Transactions) / 0.4
	avgTPS := s.calculateAverageTPS()

	log.Printf("Block %d: Epoch %d, Instant TPS: %.2f, Avg TPS: %.2f (txs=%d)",
		header.Number, epoch, instantTPS, avgTPS, header.Transactions)

	// NOTE: Do NOT call updateMetricsFromBlock here!
	// It will be called from processSubscribedBlocks to avoid duplicate updates
}

// addRecentBlock adds a block to the recent blocks list for TPS calculation
func (s *MonadSubscriber) addRecentBlock(timestamp int64, txCount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add new block
	s.recentBlocks = append(s.recentBlocks, BlockTxInfo{
		Timestamp:    timestamp,
		Transactions: txCount,
	})

	// Keep only the most recent blocks
	if len(s.recentBlocks) > s.maxRecentBlocks {
		s.recentBlocks = s.recentBlocks[1:]
	}
}

// calculateAverageTPS calculates TPS based on recent blocks (all available data)
func (s *MonadSubscriber) calculateAverageTPS() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.recentBlocks) < 2 {
		return 0
	}

	// Calculate total transactions and time span
	totalTx := 0
	for _, block := range s.recentBlocks {
		totalTx += block.Transactions
	}

	// Time difference between first and last block
	firstBlock := s.recentBlocks[0]
	lastBlock := s.recentBlocks[len(s.recentBlocks)-1]
	timeSpanSeconds := float64(lastBlock.Timestamp - firstBlock.Timestamp)

	if timeSpanSeconds <= 0 {
		// Fallback: use block count * 0.4s
		timeSpanSeconds = float64(len(s.recentBlocks)-1) * 0.4
	}

	return float64(totalTx) / timeSpanSeconds
}

// calculateOneSecondTPS calculates TPS for exactly 1 second of recent blocks
func (s *MonadSubscriber) calculateOneSecondTPS() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.recentBlocks) < 2 {
		return 0
	}

	lastBlock := s.recentBlocks[len(s.recentBlocks)-1]
	oneSecondAgo := lastBlock.Timestamp - 1 // 1 second ago

	// Sum transactions from blocks within the last 1 second
	totalTx := 0
	for i := len(s.recentBlocks) - 1; i >= 0; i-- {
		block := s.recentBlocks[i]
		if block.Timestamp >= oneSecondAgo {
			totalTx += block.Transactions
		} else {
			break
		}
	}

	return float64(totalTx) // Already per second
}

// getInstantTPS returns TPS for the most recent block only
func (s *MonadSubscriber) getInstantTPS() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.recentBlocks) == 0 {
		return 0
	}

	lastBlock := s.recentBlocks[len(s.recentBlocks)-1]
	return float64(lastBlock.Transactions) / 0.4 // Per 0.4s block time
}

// addTPSToHistory adds current TPS metrics to history for charting
func (s *MonadSubscriber) addTPSToHistory(oneSecondTPS, avgTPS, instantTPS float64, txCount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add new data point: [total, vote, avg, instant, txCount]
	s.tpsHistory = append(s.tpsHistory, [5]float64{oneSecondTPS, 0, avgTPS, instantTPS, float64(txCount)})

	// Keep only the most recent points
	if len(s.tpsHistory) > s.maxHistorySize {
		s.tpsHistory = s.tpsHistory[1:]
	}
}

// getTPSHistory returns the full TPS history for charting
func (s *MonadSubscriber) getTPSHistory() [][5]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Make a copy to avoid race conditions
	historyCopy := make([][5]float64, len(s.tpsHistory))
	copy(historyCopy, s.tpsHistory)
	return historyCopy
}

// parseBlockHeader converts JSON to BlockHeader
func (s *MonadSubscriber) parseBlockHeader(result map[string]interface{}) *BlockHeader {
	numberStr, ok := result["number"].(string)
	if !ok {
		return nil
	}

	number, err := parseHexToInt64(numberStr)
	if err != nil {
		log.Printf("Failed to parse block number: %v", err)
		return nil
	}

	timestampStr, ok := result["timestamp"].(string)
	if !ok {
		return nil
	}

	timestamp, err := parseHexToInt64(timestampStr)
	if err != nil {
		log.Printf("Failed to parse timestamp: %v", err)
		return nil
	}

	hash, _ := result["hash"].(string)

	// Parse transaction count
	txCount := 0
	if txs, ok := result["transactions"].([]interface{}); ok {
		txCount = len(txs)
	}

	// Parse gas used
	gasUsed := int64(0)
	if gasUsedStr, ok := result["gasUsed"].(string); ok {
		gasUsed, _ = parseHexToInt64(gasUsedStr)
	}

	return &BlockHeader{
		Number:       number,
		Hash:         hash,
		Timestamp:    timestamp,
		Transactions: txCount,
		GasUsed:      gasUsed,
	}
}

// GetLatestBlock returns the most recent block header
func (s *MonadSubscriber) GetLatestBlock() *BlockHeader {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestBlock
}

// IsConnected returns connection status
func (s *MonadSubscriber) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isConnected
}

// BlockChannel returns the channel for receiving new blocks
func (s *MonadSubscriber) BlockChannel() <-chan *BlockHeader {
	return s.blockChan
}

// LogsChannel returns the channel for receiving transaction logs
func (s *MonadSubscriber) LogsChannel() <-chan *TransactionLog {
	return s.logsChan
}

// reconnect attempts to reconnect to the WebSocket
func (s *MonadSubscriber) reconnect() error {
	log.Println("Attempting to reconnect to Monad WebSocket...")

	s.mu.Lock()
	if s.conn != nil {
		s.conn.Close()
	}
	s.isConnected = false
	s.mu.Unlock()

	return s.Connect()
}

// Close closes the WebSocket connection
func (s *MonadSubscriber) Close() error {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		// Unsubscribe from newHeads
		if s.headsSubID != "" {
			unsubMsg := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      3,
				"method":  "eth_unsubscribe",
				"params":  []string{s.headsSubID},
			}
			s.conn.WriteJSON(unsubMsg)
		}

		// Unsubscribe from monadLogs
		if s.logsSubID != "" {
			unsubMsg := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      4,
				"method":  "eth_unsubscribe",
				"params":  []string{s.logsSubID},
			}
			s.conn.WriteJSON(unsubMsg)
		}

		return s.conn.Close()
	}

	return nil
}

// ToConsensusMetrics converts BlockHeader to ConsensusMetrics
func (h *BlockHeader) ToConsensusMetrics() *ConsensusMetrics {
	return &ConsensusMetrics{
		CurrentHeight:     h.Number,
		LastBlockTime:     h.Timestamp,
		BlockTime:         0.4,
		ValidatorCount:    100,
		VotingPower:       1000000,
		ParticipationRate: 0.9,
	}
}

// ToExecutionMetrics converts BlockHeader to ExecutionMetrics
// Note: Prioritizes Prometheus TPS for accuracy
func (h *BlockHeader) ToExecutionMetrics() *ExecutionMetrics {
	// Priority 1: Use Prometheus TPS (most accurate)
	var tps float64
	promCollector := GetPrometheusCollector()
	if promCollector != nil && promCollector.IsHealthy() {
		tps = promCollector.GetTPS()
		// log.Printf("Using Prometheus TPS: %.2f", tps)
	} else if monadSubscriber != nil {
		// Priority 2: Use subscriber's average TPS
		tps = monadSubscriber.calculateAverageTPS()
		// log.Printf("Using subscriber average TPS: %.2f", tps)
	} else {
		// Priority 3: Fallback to instant TPS
		tps = float64(h.Transactions) / 0.4
		// log.Printf("Using instant TPS: %.2f", tps)
	}

	return &ExecutionMetrics{
		TPS:                 tps,
		PendingTxCount:      0, // Would need separate call
		ParallelSuccessRate: 0.85,
		AvgGasPrice:         21,
		AvgExecutionTime:    5.0,
		StateSize:           1000000000,
	}
}

// Global subscriber instance
var monadSubscriber *MonadSubscriber

// InitializeSubscriber creates and starts the subscriber
func InitializeSubscriber(wsURL string) error {
	monadSubscriber = NewMonadSubscriber(wsURL)

	if err := monadSubscriber.Connect(); err != nil {
		return err
	}

	// Start processing blocks
	go processSubscribedBlocks()

	return nil
}

// processSubscribedBlocks processes incoming blocks, logs, and updates metrics
func processSubscribedBlocks() {
	for {
		select {
		case block := <-monadSubscriber.BlockChannel():
			if block != nil {
				updateMetricsFromBlock(block)
			}
		case txLog := <-monadSubscriber.LogsChannel():
			if txLog != nil {
				broadcastTransactionLog(txLog)
			}
		case err := <-monadSubscriber.errorChan:
			log.Printf("Subscriber error: %v", err)
		}
	}
}

// broadcastTransactionLog sends transaction log to all connected WebSocket clients
func broadcastTransactionLog(txLog *TransactionLog) {
	// Format as Firedancer protocol message
	msg := map[string]interface{}{
		"topic": "tx_flow",
		"key":   "transaction_log",
		"value": map[string]interface{}{
			"block_number":      txLog.BlockNumber,
			"transaction_hash":  txLog.TransactionHash,
			"transaction_index": txLog.TransactionIndex,
			"address":           txLog.Address,
			"topics":            txLog.Topics,
			"data":              txLog.Data,
			"timestamp":         txLog.Timestamp,
		},
	}

	// Broadcast to all connected clients (defined in main.go)
	broadcastToAllClients(msg)
}

// updateMetricsFromBlock updates global metrics from a new block
func updateMetricsFromBlock(block *BlockHeader) {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	// Update consensus tracker with new block
	consensusTracker := GetConsensusTracker()
	if consensusTracker != nil {
		consensusTracker.OnBlockProposed(uint64(block.Number), block.Hash, block.Transactions)
	}

	now := time.Now()

	// Get network metrics (these don't change per block)
	network, _ := monadClient.GetNetworkMetrics()
	if network == nil {
		network = &NetworkMetrics{
			PeerCount:      50,
			InboundPeers:   25,
			OutboundPeers:  25,
			BytesIn:        1000000,
			BytesOut:       1000000,
			NetworkLatency: 50.0,
		}
	}

	consensus := block.ToConsensusMetrics()
	execution := block.ToExecutionMetrics()

	// Update current metrics with real-time data
	currentMetrics = MonadMetrics{
		Timestamp: now.Unix(),
		NodeInfo: NodeInfo{
			Version:  "0.1.0",
			ChainID:  20143,
			NodeName: getNodeName(),
			Status:   "running",
			Uptime:   int64(now.Sub(startTime).Seconds()),
		},
		Waterfall: generateWaterfallFromExecution(execution),
		Consensus: *consensus,
		Execution: *execution,
		Network:   *network,
	}

	log.Printf("Updated metrics from real-time block: height=%d, tps=%.2f",
		block.Number, execution.TPS)
}
