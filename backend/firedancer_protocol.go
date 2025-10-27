package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Firedancer protocol message types

type FiredancerMessage struct {
	Topic string      `json:"topic"`
	Key   string      `json:"key"`
	Value interface{} `json:"value,omitempty"`
	ID    *int        `json:"id,omitempty"`
}

// Summary messages
func sendInitialSummaryMessages(conn *websocket.Conn) error {
	messages := []FiredancerMessage{
		{
			Topic: "summary",
			Key:   "version",
			Value: "0.1.0",
		},
		{
			Topic: "summary",
			Key:   "cluster",
			Value: "development",
		},
		{
			Topic: "summary",
			Key:   "identity_key",
			Value: "MonadValidator1111111111111111111111111",
		},
		{
			Topic: "summary",
			Key:   "startup_time_nanos",
			Value: time.Now().UnixNano(),
		},
		{
			Topic: "summary",
			Key:   "startup_progress",
			Value: map[string]interface{}{
				"phase":                                                 "running",
				"downloading_full_snapshot_slot":                        nil,
				"downloading_full_snapshot_peer":                        nil,
				"downloading_full_snapshot_elapsed_secs":                nil,
				"downloading_full_snapshot_remaining_secs":              nil,
				"downloading_full_snapshot_throughput":                  nil,
				"downloading_full_snapshot_total_bytes":                 nil,
				"downloading_full_snapshot_current_bytes":               nil,
				"downloading_incremental_snapshot_slot":                 nil,
				"downloading_incremental_snapshot_peer":                 nil,
				"downloading_incremental_snapshot_elapsed_secs":         nil,
				"downloading_incremental_snapshot_remaining_secs":       nil,
				"downloading_incremental_snapshot_throughput":           nil,
				"downloading_incremental_snapshot_total_bytes":          nil,
				"downloading_incremental_snapshot_current_bytes":        nil,
				"ledger_slot":                                           nil,
				"ledger_max_slot":                                       nil,
				"waiting_for_supermajority_slot":                        nil,
				"waiting_for_supermajority_stake_percent":               nil,
			},
		},
		{
			Topic: "summary",
			Key:   "vote_state",
			Value: "non-voting",
		},
	}

	for _, msg := range messages {
		if err := conn.WriteJSON(msg); err != nil {
			return err
		}
	}

	return nil
}

// Send peers data to satisfy startup screen requirements
func sendPeersMessage(conn *websocket.Conn) error {
	// Get node name from config
	nodeName := getNodeName()

	// Send a simple peers update with at least one peer
	// This will make hasPeers === true in the frontend
	peersMsg := FiredancerMessage{
		Topic: "peers",
		Key:   "update",
		Value: map[string]interface{}{
			"add": []map[string]interface{}{
				{
					"identity_pubkey": "MonadValidator1111111111111111111111111",
					"gossip": map[string]interface{}{
						"wallclock":     time.Now().Unix(),
						"shred_version": 1,
						"version":       "1.0.0",
						"feature_set":   nil,
						"sockets":       map[string]string{},
					},
					"vote": []map[string]interface{}{},
					"info": map[string]interface{}{
						"name":     nodeName,
						"details":  nil,
						"website":  nil,
						"icon_url": nil,
					},
				},
			},
		},
	}

	return conn.WriteJSON(peersMsg)
}

// Send epoch information
func sendEpochMessage(conn *websocket.Conn) error {
	// Get current epoch from Monad
	epoch, err := monadClient.GetCurrentEpoch()
	if err != nil {
		log.Printf("Failed to get current epoch: %v, using default", err)
		epoch = 0
	}

	// Calculate epoch boundaries (50,000 blocks per epoch)
	epochSize := int64(50000)
	startSlot := epoch * epochSize
	endSlot := (epoch + 1) * epochSize

	epochMsg := FiredancerMessage{
		Topic: "epoch",
		Key:   "new",
		Value: map[string]interface{}{
			"epoch":                    epoch,
			"start_time_nanos":         nil,
			"end_time_nanos":           nil,
			"start_slot":               startSlot,
			"end_slot":                 endSlot,
			"excluded_stake_lamports":  0,
			"staked_pubkeys":           []string{},
			"staked_lamports":          []int64{},
			"leader_slots":             []int{}, // Empty for Monad
		},
	}

	return conn.WriteJSON(epochMsg)
}

// Send periodic updates
func sendFiredancerUpdates(conn *websocket.Conn) {
	// Update every 200ms to catch all blocks (Monad block time is 400ms)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	pingID := 0

	for {
		select {
		case <-ticker.C:
			// Fetch fresh metrics directly from Monad on each update
			// This ensures we don't miss any blocks
			consensus, err := monadClient.GetConsensusMetrics()
			if err != nil {
				log.Printf("Error fetching consensus metrics: %v", err)
				continue
			}

			// Get cached metrics for other data
			metrics := getCurrentMetrics()
			// Update with fresh consensus data
			metrics.Consensus = *consensus

			// Send ping
			pingID++
			pingMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "ping",
				Value: nil,
				ID:    &pingID,
			}
			if err := conn.WriteJSON(pingMsg); err != nil {
				log.Printf("Error sending ping: %v", err)
				return
			}

			// Send estimated slot (block height)
			estimatedSlotMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "estimated_slot",
				Value: metrics.Consensus.CurrentHeight,
			}
			if err := conn.WriteJSON(estimatedSlotMsg); err != nil {
				log.Printf("Error sending estimated_slot: %v", err)
				return
			}

			// Also send as root_slot and completed_slot for compatibility
			rootSlotMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "root_slot",
				Value: metrics.Consensus.CurrentHeight,
			}
			if err := conn.WriteJSON(rootSlotMsg); err != nil {
				log.Printf("Error sending root_slot: %v", err)
				return
			}

			completedSlotMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "completed_slot",
				Value: metrics.Consensus.CurrentHeight,
			}
			if err := conn.WriteJSON(completedSlotMsg); err != nil {
				log.Printf("Error sending completed_slot: %v", err)
				return
			}

			// Calculate different TPS metrics from subscriber
			var oneSecondTPS, avgTPS, instantTPS float64
			if monadSubscriber != nil && monadSubscriber.IsConnected() {
				oneSecondTPS = monadSubscriber.calculateOneSecondTPS()
				avgTPS = monadSubscriber.calculateAverageTPS()
				instantTPS = monadSubscriber.getInstantTPS()

				// Add to history for charting
				monadSubscriber.addTPSToHistory(oneSecondTPS, avgTPS, instantTPS)
			} else {
				// Fallback to current metrics
				oneSecondTPS = metrics.Execution.TPS
				avgTPS = metrics.Execution.TPS
				instantTPS = metrics.Execution.TPS
			}

			// Send estimated TPS with 3 different metrics
			// total: 1-second TPS (most recent second)
			// nonvote_success: Average TPS (smoothed over ~4 seconds)
			// nonvote_failed: Instant TPS (per block, shows spikes)
			estimatedTpsMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "estimated_tps",
				Value: map[string]interface{}{
					"total":           oneSecondTPS,  // 1-second TPS
					"vote":            0,
					"nonvote_success": avgTPS,        // Average TPS
					"nonvote_failed":  instantTPS,    // Instant TPS per block
				},
			}
			if err := conn.WriteJSON(estimatedTpsMsg); err != nil {
				log.Printf("Error sending estimated_tps: %v", err)
				return
			}

			// Send Monad waterfall (NEW: Monad lifecycle-aligned)
			// Generate waterfall data using new Monad-specific structure
			monadWaterfallData := GenerateMonadWaterfall()

			// Debug: Log waterfall data source
			if metadata, ok := monadWaterfallData["metadata"].(map[string]interface{}); ok {
				if source, ok := metadata["source"].(string); ok {
					log.Printf("🌊 Monad Waterfall source: %s", source)
				}
			}

			// Send NEW waterfall format (nodes + links for Sankey diagram)
			waterfallMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "monad_waterfall_v2",
				Value: monadWaterfallData,
			}
			if err := conn.WriteJSON(waterfallMsg); err != nil {
				log.Printf("Error sending Monad waterfall v2: %v", err)
				return
			}

			// Also send legacy waterfall format for backward compatibility
			// TODO: Remove after frontend is fully migrated to v2
			legacyWaterfallData := GenerateWaterfallFromSubscriber()
			waterfallIn := legacyWaterfallData["in"].(map[string]interface{})
			waterfallOut := legacyWaterfallData["out"].(map[string]interface{})

			legacyWaterfallMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "live_txn_waterfall",
				Value: map[string]interface{}{
					"next_leader_slot": nil,
					"waterfall": map[string]interface{}{
						"in": map[string]interface{}{
							"quic":           waterfallIn["rpc"],
							"udp":            waterfallIn["p2p"],
							"gossip":         waterfallIn["gossip"],
							"pack_cranked":   0,
							"pack_retained":  0,
							"resolv_retained": 0,
							"block_engine":   0,
						},
						"out": map[string]interface{}{
							"net_overrun":           0,
							"quic_overrun":          0,
							"quic_frag_drop":        0,
							"quic_abandoned":        0,
							"tpu_quic_invalid":      0,
							"tpu_udp_invalid":       0,
							"verify_overrun":        0,
							"verify_parse":          0,
							"verify_failed":         waterfallOut["verify_failed"],
							"verify_duplicate":      waterfallOut["nonce_failed"],
							"dedup_duplicate":       waterfallOut["nonce_failed"],
							"resolv_lut_failed":     waterfallOut["balance_failed"],
							"resolv_expired":        waterfallOut["pool_fee_dropped"],
							"resolv_no_ledger":      0,
							"resolv_ancient":        0,
							"resolv_retained":       0,
							"pack_invalid":          0,
							"pack_invalid_bundle":   0,
							"pack_retained":         0,
							"pack_leader_slow":      0,
							"pack_wait_full":        waterfallOut["pool_full"],
							"pack_expired":          0,
							"bank_invalid":          waterfallOut["exec_failed"],
							"block_success":         waterfallOut["exec_parallel"],
							"block_fail":            waterfallOut["exec_sequential"],
						},
					},
				},
			}
			if err := conn.WriteJSON(legacyWaterfallMsg); err != nil {
				log.Printf("Error sending legacy waterfall: %v", err)
				return
			}

			// Send MonadBFT consensus state
			consensusTracker := GetConsensusTracker()
			if consensusTracker != nil {
				consensusStateMsg := FiredancerMessage{
					Topic: "summary",
					Key:   "monad_consensus_state",
					Value: consensusTracker.GetConsensusState(),
				}
				if err := conn.WriteJSON(consensusStateMsg); err != nil {
					log.Printf("Error sending consensus state: %v", err)
					return
				}
			}

			// Send vote distance
			voteDistanceMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "vote_distance",
				Value: 0,
			}
			if err := conn.WriteJSON(voteDistanceMsg); err != nil {
				log.Printf("Error sending vote_distance: %v", err)
				return
			}

			// Send TPS history for the chart
			// Get accumulated history from subscriber
			var tpsHistoryData [][]float64
			if monadSubscriber != nil && monadSubscriber.IsConnected() {
				history := monadSubscriber.getTPSHistory()
				// Convert [][4]float64 to [][]float64
				tpsHistoryData = make([][]float64, len(history))
				for i, h := range history {
					tpsHistoryData[i] = []float64{h[0], h[1], h[2], h[3]}
				}
			} else {
				// Fallback: send single point
				tpsHistoryData = [][]float64{
					{oneSecondTPS, 0, avgTPS, instantTPS},
				}
			}

			tpsHistoryMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "tps_history",
				Value: tpsHistoryData,
			}
			if err := conn.WriteJSON(tpsHistoryMsg); err != nil {
				log.Printf("Error sending tps_history: %v", err)
				return
			}

			// Debug: log message count
			log.Printf("Sent Firedancer updates: ping=%d, slot=%d, 1s=%.2f, avg=%.2f, instant=%.2f, history=%d",
				pingID, metrics.Consensus.CurrentHeight, oneSecondTPS, avgTPS, instantTPS, len(tpsHistoryData))
		}
	}
}

// Handle incoming client messages
func handleFiredancerClientMessage(conn *websocket.Conn, msgBytes []byte) error {
	var clientMsg map[string]interface{}
	if err := json.Unmarshal(msgBytes, &clientMsg); err != nil {
		return err
	}

	log.Printf("Received client message: %v", clientMsg)

	// Handle subscription requests
	if topic, ok := clientMsg["topic"].(string); ok {
		if topic == "summary" {
			// Client is subscribing to summary topic
			// We already send summary updates periodically
		}
	}

	return nil
}
