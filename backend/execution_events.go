package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// ExecutionEvent represents a single event from Monad's execution engine
type ExecutionEvent struct {
	// Fixed 64-byte header
	Header ExecutionEventHeader `json:"header"`
	// Variable-length payload
	Payload []byte `json:"payload"`
	// Parsed payload data
	Data interface{} `json:"data,omitempty"`
}

// ExecutionEventHeader is the fixed 64-byte header for each event
type ExecutionEventHeader struct {
	SequenceNumber uint64    `json:"sequence_number"`
	Timestamp      uint64    `json:"timestamp"`
	EventType      uint32    `json:"event_type"`
	PayloadSize    uint32    `json:"payload_size"`
	TransactionID  [32]byte  `json:"transaction_id"`
	Reserved       [16]byte  `json:"-"` // Reserved space
}

// Event types based on Monad's execution pipeline
const (
	EventTypeTransactionStart = iota + 1
	EventTypeTransactionEnd
	EventTypeStateRead
	EventTypeStateWrite
	EventTypeLogEmitted
	EventTypeContractCall
	EventTypeGasUsage
	EventTypeError
)

// Parsed event data structures
type TransactionStartEvent struct {
	Sender    string `json:"sender"`
	To        string `json:"to"`
	GasLimit  uint64 `json:"gas_limit"`
	GasPrice  uint64 `json:"gas_price"`
	Value     string `json:"value"`
	Data      string `json:"data"`
}

type TransactionEndEvent struct {
	Success   bool   `json:"success"`
	GasUsed   uint64 `json:"gas_used"`
	ExitCode  uint32 `json:"exit_code"`
	Duration  uint64 `json:"duration_ns"`
}

type StateChangeEvent struct {
	Address string `json:"address"`
	Key     string `json:"key"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

type LogEvent struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

// EventRingReader manages connection to Monad's event rings
type EventRingReader struct {
	conn           net.Conn
	connected      bool
	eventChan      chan ExecutionEvent
	stopChan       chan struct{}
	mutex          sync.RWMutex
	lastSequence   uint64
	missedEvents   uint64

	// Event processing stats
	eventsReceived uint64
	bytesReceived  uint64
	parseErrors    uint64
}

// NewEventRingReader creates a new reader for Monad execution events
func NewEventRingReader(socketPath string) *EventRingReader {
	return &EventRingReader{
		eventChan: make(chan ExecutionEvent, 1000), // Buffer for high throughput
		stopChan:  make(chan struct{}),
	}
}

// Connect establishes connection to the event ring socket
func (r *EventRingReader) Connect(socketPath string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.connected {
		return fmt.Errorf("already connected")
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to event ring %s: %w", socketPath, err)
	}

	r.conn = conn
	r.connected = true

	log.Printf("Connected to Monad event ring: %s", socketPath)

	// Start reading events in background
	go r.readEvents()

	return nil
}

// Disconnect closes the connection to event rings
func (r *EventRingReader) Disconnect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.connected {
		return nil
	}

	close(r.stopChan)

	if r.conn != nil {
		r.conn.Close()
		r.conn = nil
	}

	r.connected = false
	log.Printf("Disconnected from Monad event ring")

	return nil
}

// Events returns the channel for receiving execution events
func (r *EventRingReader) Events() <-chan ExecutionEvent {
	return r.eventChan
}

// GetStats returns current statistics about event processing
func (r *EventRingReader) GetStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return map[string]interface{}{
		"connected":        r.connected,
		"events_received":  r.eventsReceived,
		"bytes_received":   r.bytesReceived,
		"missed_events":    r.missedEvents,
		"parse_errors":     r.parseErrors,
		"last_sequence":    r.lastSequence,
		"buffer_size":      len(r.eventChan),
	}
}

// readEvents continuously reads events from the socket
func (r *EventRingReader) readEvents() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Event reader panic: %v", r)
		}
	}()

	buffer := make([]byte, 4096) // Buffer for reading

	for {
		select {
		case <-r.stopChan:
			return
		default:
		}

		// Set read timeout
		r.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// Read the fixed 64-byte header
		header := ExecutionEventHeader{}
		if err := binary.Read(r.conn, binary.LittleEndian, &header); err != nil {
			if err == io.EOF {
				log.Printf("Event ring connection closed")
				return
			}
			log.Printf("Failed to read event header: %v", err)
			continue
		}

		// Check for missed events
		if header.SequenceNumber > r.lastSequence+1 && r.lastSequence > 0 {
			missed := header.SequenceNumber - r.lastSequence - 1
			r.mutex.Lock()
			r.missedEvents += missed
			r.mutex.Unlock()
			log.Printf("Missed %d events (seq %d -> %d)", missed, r.lastSequence, header.SequenceNumber)
		}
		r.lastSequence = header.SequenceNumber

		// Read payload if present
		var payload []byte
		if header.PayloadSize > 0 {
			if header.PayloadSize > uint32(len(buffer)) {
				buffer = make([]byte, header.PayloadSize)
			}

			n, err := io.ReadFull(r.conn, buffer[:header.PayloadSize])
			if err != nil {
				log.Printf("Failed to read event payload: %v", err)
				r.mutex.Lock()
				r.parseErrors++
				r.mutex.Unlock()
				continue
			}
			payload = make([]byte, n)
			copy(payload, buffer[:n])
		}

		// Create event
		event := ExecutionEvent{
			Header:  header,
			Payload: payload,
		}

		// Parse payload based on event type
		if err := r.parseEventPayload(&event); err != nil {
			log.Printf("Failed to parse event payload (type %d): %v", header.EventType, err)
			r.mutex.Lock()
			r.parseErrors++
			r.mutex.Unlock()
		}

		// Update stats
		r.mutex.Lock()
		r.eventsReceived++
		r.bytesReceived += uint64(64 + len(payload)) // Header + payload
		r.mutex.Unlock()

		// Send event to channel (non-blocking)
		select {
		case r.eventChan <- event:
		default:
			// Channel full, drop event and log warning
			log.Printf("Event channel full, dropping event seq %d", header.SequenceNumber)
		}
	}
}

// parseEventPayload parses the variable-length payload based on event type
func (r *EventRingReader) parseEventPayload(event *ExecutionEvent) error {
	if len(event.Payload) == 0 {
		return nil
	}

	switch event.Header.EventType {
	case EventTypeTransactionStart:
		var data TransactionStartEvent
		if err := json.Unmarshal(event.Payload, &data); err != nil {
			return err
		}
		event.Data = data

	case EventTypeTransactionEnd:
		var data TransactionEndEvent
		if err := json.Unmarshal(event.Payload, &data); err != nil {
			return err
		}
		event.Data = data

	case EventTypeStateRead, EventTypeStateWrite:
		var data StateChangeEvent
		if err := json.Unmarshal(event.Payload, &data); err != nil {
			return err
		}
		event.Data = data

	case EventTypeLogEmitted:
		var data LogEvent
		if err := json.Unmarshal(event.Payload, &data); err != nil {
			return err
		}
		event.Data = data

	default:
		// Unknown event type, keep raw payload
		log.Printf("Unknown event type: %d", event.Header.EventType)
	}

	return nil
}

// Global event ring reader instances
var (
	executionEventReader *EventRingReader
	eventReaderMutex     sync.RWMutex
)

// InitializeEventRings initializes connections to Monad event rings
func InitializeEventRings() error {
	eventReaderMutex.Lock()
	defer eventReaderMutex.Unlock()

	// Initialize execution event reader
	executionEventReader = NewEventRingReader("/tmp/monad-execution-events.sock")

	// Try to connect (will fallback gracefully if socket doesn't exist)
	if err := executionEventReader.Connect("/tmp/monad-execution-events.sock"); err != nil {
		log.Printf("Failed to connect to execution events: %v", err)
		log.Printf("Event ring features will be disabled")
		return err
	}

	log.Printf("Event ring connections initialized successfully")
	return nil
}

// GetExecutionEventReader returns the global execution event reader
func GetExecutionEventReader() *EventRingReader {
	eventReaderMutex.RLock()
	defer eventReaderMutex.RUnlock()
	return executionEventReader
}

// StartEventProcessing starts processing execution events for dashboard metrics
func StartEventProcessing() {
	go func() {
		reader := GetExecutionEventReader()
		if reader == nil {
			return
		}

		log.Printf("Starting execution event processing...")

		for event := range reader.Events() {
			processExecutionEvent(event)
		}
	}()
}

// processExecutionEvent processes individual execution events and updates metrics
func processExecutionEvent(event ExecutionEvent) {
	// Update real-time metrics based on execution events
	switch event.Header.EventType {
	case EventTypeTransactionStart:
		if data, ok := event.Data.(TransactionStartEvent); ok {
			log.Printf("Transaction started: %s -> %s, Gas: %d", data.Sender, data.To, data.GasLimit)
			// Update waterfall metrics: transaction ingress
			updateWaterfallFromEvent("transaction_start", 1)
		}

	case EventTypeTransactionEnd:
		if data, ok := event.Data.(TransactionEndEvent); ok {
			log.Printf("Transaction completed: Success=%t, Gas=%d, Duration=%dns",
				data.Success, data.GasUsed, data.Duration)
			// Update waterfall metrics: transaction completion
			if data.Success {
				updateWaterfallFromEvent("transaction_success", 1)
			} else {
				updateWaterfallFromEvent("transaction_failed", 1)
			}
		}

	case EventTypeStateWrite:
		if data, ok := event.Data.(StateChangeEvent); ok {
			log.Printf("State change: %s[%s] = %s", data.Address, data.Key, data.NewValue)
			updateWaterfallFromEvent("state_write", 1)
		}

	case EventTypeLogEmitted:
		if data, ok := event.Data.(LogEvent); ok {
			log.Printf("Log emitted: %s, topics: %v", data.Address, data.Topics)
			updateWaterfallFromEvent("log_emitted", 1)
		}
	}
}

// updateWaterfallFromEvent updates waterfall metrics based on execution events
func updateWaterfallFromEvent(eventName string, count int64) {
	// This will integrate with the existing metrics system
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	// Update appropriate waterfall counters based on event type
	switch eventName {
	case "transaction_start":
		currentMetrics.Waterfall.RPCReceived += count
	case "transaction_success":
		currentMetrics.Waterfall.EVMParallelExecuted += count
	case "transaction_failed":
		currentMetrics.Waterfall.SignatureFailed += count
	case "state_write":
		currentMetrics.Waterfall.StateUpdated += count
	case "log_emitted":
		// Could add a new metric for logs emitted
	}
}