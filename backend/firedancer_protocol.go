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
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	pingID := 0

	for {
		select {
		case <-ticker.C:
			// Get current metrics
			metrics := getCurrentMetrics()

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

			// Send estimated TPS
			estimatedTpsMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "estimated_tps",
				Value: map[string]interface{}{
					"total":          metrics.Execution.TPS,
					"vote":           0,
					"nonvote_success": metrics.Execution.TPS,
					"nonvote_failed":  0,
				},
			}
			if err := conn.WriteJSON(estimatedTpsMsg); err != nil {
				log.Printf("Error sending estimated_tps: %v", err)
				return
			}

			// Send live txn waterfall
			waterfallMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "live_txn_waterfall",
				Value: map[string]interface{}{
					"next_leader_slot": nil,
					"waterfall": map[string]interface{}{
						"in": map[string]interface{}{
							"pack_cranked":   0,
							"pack_retained":  0,
							"resolv_retained": 0,
							"quic":           int(metrics.Waterfall.RPCReceived),
							"udp":            int(metrics.Waterfall.GossipReceived),
							"gossip":         int(metrics.Waterfall.GossipReceived),
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
							"verify_failed":         int(metrics.Waterfall.SignatureFailed),
							"verify_duplicate":      int(metrics.Waterfall.NonceDuplicate),
							"dedup_duplicate":       int(metrics.Waterfall.NonceDuplicate),
							"resolv_lut_failed":     0,
							"resolv_expired":        0,
							"resolv_no_ledger":      0,
							"resolv_ancient":        0,
							"resolv_retained":       0,
							"pack_invalid":          int(metrics.Waterfall.GasInvalid),
							"pack_invalid_bundle":   0,
							"pack_retained":         0,
							"pack_leader_slow":      0,
							"pack_wait_full":        0,
							"pack_expired":          0,
							"bank_invalid":          0,
							"block_success":         int(metrics.Waterfall.EVMParallelExecuted),
							"block_fail":            int(metrics.Waterfall.EVMSequentialFallback),
						},
					},
				},
			}
			if err := conn.WriteJSON(waterfallMsg); err != nil {
				log.Printf("Error sending waterfall: %v", err)
				return
			}

			// Send root slot
			rootSlotMsg := FiredancerMessage{
				Topic: "summary",
				Key:   "root_slot",
				Value: metrics.Consensus.CurrentHeight,
			}
			if err := conn.WriteJSON(rootSlotMsg); err != nil {
				log.Printf("Error sending root_slot: %v", err)
				return
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

			// Debug: log message count
			log.Printf("Sent Firedancer updates: ping=%d, slot=%d, tps=%.2f",
				pingID, metrics.Consensus.CurrentHeight, metrics.Execution.TPS)
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
