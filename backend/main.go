package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

//go:embed frontend/dist
var static embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
	Subprotocols: []string{"compress-zstd"},
}

// WebSocket client registry for broadcasting transaction logs
var (
	wsClients   = make(map[*websocket.Conn]bool)
	wsClientsMu sync.RWMutex
)

// registerWSClient adds a WebSocket connection to the registry
func registerWSClient(conn *websocket.Conn) {
	wsClientsMu.Lock()
	defer wsClientsMu.Unlock()
	wsClients[conn] = true
	log.Printf("WebSocket client registered. Total clients: %d", len(wsClients))
}

// unregisterWSClient removes a WebSocket connection from the registry
func unregisterWSClient(conn *websocket.Conn) {
	wsClientsMu.Lock()
	defer wsClientsMu.Unlock()
	delete(wsClients, conn)
	log.Printf("WebSocket client unregistered. Total clients: %d", len(wsClients))
}

// broadcastToAllClients sends a message to all connected WebSocket clients
func broadcastToAllClients(msg interface{}) {
	wsClientsMu.RLock()
	defer wsClientsMu.RUnlock()

	for client := range wsClients {
		if err := client.WriteJSON(msg); err != nil {
			log.Printf("Error broadcasting to client: %v", err)
			// Don't remove client here - it will be removed when connection closes
		}
	}
}

func main() {
	r := gin.Default()

	// Serve static files
	staticFiles, err := fs.Sub(static, "frontend/dist")
	if err != nil {
		log.Fatal("Failed to get static files:", err)
	}

	r.StaticFS("/assets", http.FS(staticFiles))

	// Serve index.html for root and any non-API routes
	r.NoRoute(func(c *gin.Context) {
		if c.Request.URL.Path != "/" && !gin.IsDebugging() {
			c.Header("Cache-Control", "no-cache")
		}

		// Try to serve static files first
		if c.Request.URL.Path != "/" && c.Request.URL.Path != "/websocket" &&
		   !strings.HasPrefix(c.Request.URL.Path, "/api") {
			file, err := static.ReadFile("frontend/dist" + c.Request.URL.Path)
			if err == nil {
				// Determine content type
				ext := filepath.Ext(c.Request.URL.Path)
				var contentType string
				switch ext {
				case ".js":
					contentType = "application/javascript"
				case ".css":
					contentType = "text/css"
				case ".html":
					contentType = "text/html; charset=utf-8"
				case ".json":
					contentType = "application/json"
				case ".png":
					contentType = "image/png"
				case ".svg":
					contentType = "image/svg+xml"
				default:
					contentType = "application/octet-stream"
				}
				c.Data(http.StatusOK, contentType, file)
				return
			}
		}

		// Fall back to index.html for SPA routing
		indexHTML, err := static.ReadFile("frontend/dist/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Frontend not built. Run 'make frontend' first.")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	// API Routes
	api := r.Group("/api/v1")
	{
		api.GET("/health", handleHealth)
		api.GET("/metrics", handleMetrics)
		api.GET("/waterfall", handleWaterfall)  // Legacy waterfall
		api.GET("/waterfall/v2", handleWaterfallV2)  // New Monad lifecycle waterfall
		api.GET("/consensus", handleConsensusState)  // MonadBFT consensus state
		api.GET("/event-rings", handleEventRingsStatus)
	}

	// WebSocket endpoint (Firedancer uses /websocket)
	r.GET("/websocket", handleWebSocket)

	// Initialize Consensus Tracker for MonadBFT phase tracking
	InitializeConsensusTracker()
	log.Printf("✅ MonadBFT Consensus Tracker initialized")

	// Initialize event rings connection
	if err := InitializeEventRings(); err != nil {
		log.Printf("Event rings not available: %v", err)
		log.Printf("Dashboard will use RPC-only mode")
	} else {
		// Start event processing if event rings are available
		go StartEventProcessing()
	}

	// Initialize Prometheus metrics collector for accurate TPS
	promEndpoint := os.Getenv("PROMETHEUS_ENDPOINT")
	if promEndpoint == "" {
		promEndpoint = "http://127.0.0.1:8889/metrics" // Default OTEL endpoint
	}
	log.Printf("Attempting to connect to Prometheus endpoint at %s...", promEndpoint)
	if err := InitializePrometheusCollector(promEndpoint); err != nil {
		log.Printf("Prometheus collector not available: %v", err)
		log.Printf("Will calculate TPS from block data")
	} else {
		log.Printf("✅ Prometheus collector initialized - using accurate TPS from monad_execution_ledger_num_tx_commits")
	}

	// Initialize IPC metrics collector for real metrics
	ipcPath := os.Getenv("MONAD_IPC_PATH")
	if ipcPath == "" {
		ipcPath = "/home/monad/monad-bft/mempool.sock" // Default path
	}
	log.Printf("Attempting to connect to Monad IPC at %s...", ipcPath)
	if err := InitializeIPCCollector(ipcPath); err != nil {
		log.Printf("IPC metrics collector not available: %v", err)
		log.Printf("Will use estimation-based metrics")
	} else {
		log.Printf("✅ IPC metrics collector initialized - using real Monad metrics")
	}

	// Try to initialize real-time WebSocket subscription
	wsURL := "ws://127.0.0.1:8081"
	log.Printf("Attempting to connect to Monad WebSocket at %s...", wsURL)
	if err := InitializeSubscriber(wsURL); err != nil {
		log.Printf("Failed to initialize WebSocket subscriber: %v", err)
		log.Printf("Falling back to polling mode")
		// Start metrics collection via polling as fallback
		go startMetricsCollection()
	} else {
		log.Printf("Successfully initialized real-time WebSocket subscription")
	}

	port := ":4000" // Changed from 3000 to 4000
	log.Printf("Monad Dashboard starting on %s", port)
	log.Fatal(r.Run(port))
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "0.1.0",
	})
}

func handleEventRingsStatus(c *gin.Context) {
	reader := GetExecutionEventReader()
	if reader == nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
			"message":   "Event rings not initialized",
		})
		return
	}

	stats := reader.GetStats()
	c.JSON(http.StatusOK, stats)
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket client connected from %s", c.Request.RemoteAddr)

	// Register this client for broadcasts
	registerWSClient(conn)
	defer unregisterWSClient(conn)

	// Send initial Firedancer protocol messages
	if err := sendInitialSummaryMessages(conn); err != nil {
		log.Printf("Error sending initial messages: %v", err)
		return
	}

	// Send peers message to remove startup screen
	if err := sendPeersMessage(conn); err != nil {
		log.Printf("Error sending peers message: %v", err)
		return
	}

	// Send epoch information
	if err := sendEpochMessage(conn); err != nil {
		log.Printf("Error sending epoch message: %v", err)
		return
	}

	// Start goroutine to handle incoming client messages
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}
			if err := handleFiredancerClientMessage(conn, message); err != nil {
				log.Printf("Error handling client message: %v", err)
			}
		}
	}()

	// Send periodic updates using Firedancer protocol
	go sendFiredancerUpdates(conn)

	// Wait for connection to close
	<-done
	log.Printf("WebSocket client disconnected")
}