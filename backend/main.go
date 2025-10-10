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
	}

	// WebSocket endpoint
	r.GET("/ws", handleWebSocket)

	// Start metrics collection
	go startMetricsCollection()

	log.Println("Monad Dashboard starting on :3000")
	log.Fatal(r.Run(":3000"))
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "0.1.0",
	})
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
			if err := conn.WriteJSON(metrics); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}