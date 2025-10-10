package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
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
		if c.Request.URL.Path != "/" && c.Request.URL.Path != "/ws" &&
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
		api.GET("/waterfall", handleWaterfall)
		api.GET("/event-rings", handleEventRingsStatus)
	}

	// WebSocket endpoint
	r.GET("/ws", handleWebSocket)

	// Initialize event rings connection
	if err := InitializeEventRings(); err != nil {
		log.Printf("Event rings not available: %v", err)
		log.Printf("Dashboard will use RPC-only mode")
	} else {
		// Start event processing if event rings are available
		go StartEventProcessing()
	}

	// Start metrics collection
	go startMetricsCollection()

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

	// Send real-time metrics updates
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := getCurrentMetrics()

			// Add event rings status to metrics
			reader := GetExecutionEventReader()
			if reader != nil {
				eventStats := reader.GetStats()
				_ = eventStats // Use the variable to avoid compilation error
				metrics.Timestamp = time.Now().Unix()
				// Add event stats to the response (you could extend MonadMetrics struct)
			}

			if err := conn.WriteJSON(metrics); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}