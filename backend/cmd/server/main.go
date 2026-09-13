package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/radiopx/backend/internal/middleware"
	ws "github.com/radiopx/backend/pkg/websocket"
	"github.com/rs/cors"
)

var (
	hub      *ws.Hub
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	// Initialize WebSocket hub
	hub = ws.NewHub()
	go hub.Run()

	// Get port from environment or default to 8080
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Setup routes
	setupRoutes()

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Start server
	handler := c.Handler(http.DefaultServeMux)
	addr := fmt.Sprintf(":%s", port)
	log.Printf("RadioPX Backend starting on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws/audio/{channel_id}", port)
	log.Printf("Health check: http://localhost:%s/health", port)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("Server error:", err)
	}
}

func setupRoutes() {
	// Health check
	http.HandleFunc("/health", healthHandler)

	// Auth routes (public)
	http.HandleFunc("/api/v1/auth/register", registerHandler)
	http.HandleFunc("/api/v1/auth/login", loginHandler)

	// Protected routes
	http.Handle("/api/v1/channels", middleware.AuthMiddleware(http.HandlerFunc(channelsHandler)))
	http.Handle("/api/v1/channels/join", middleware.AuthMiddleware(http.HandlerFunc(joinChannelHandler)))
	http.Handle("/api/v1/channels/leave", middleware.AuthMiddleware(http.HandlerFunc(leaveChannelHandler)))
	http.Handle("/api/v1/location/update", middleware.AuthMiddleware(http.HandlerFunc(updateLocationHandler)))

	// WebSocket endpoint
	http.HandleFunc("/ws/audio/", websocketHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","service":"radiopx-backend","connections":%d}`, hub.GetClientCount())
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message":"User registered successfully"}`)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Login successful","token":"placeholder-jwt-token"}`)
}

func channelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"channels":[]}`)
}

func joinChannelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Joined channel successfully"}`)
}

func leaveChannelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Left channel successfully"}`)
}

func updateLocationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Location updated successfully"}`)
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Extract channel ID from URL
	channelID := r.URL.Query().Get("channel")
	if channelID == "" {
		http.Error(w, `{"error":"Channel ID required"}`, http.StatusBadRequest)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Create client
	client := ws.NewClient(hub, conn, "anonymous")

	// Register client
	hub.RegisterClient(client)

	// Join room
	hub.JoinRoom(client, channelID)

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()
}
