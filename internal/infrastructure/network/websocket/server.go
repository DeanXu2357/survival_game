package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"survival/internal/app"
	"survival/internal/protocol"
	"survival/internal/utils"
)

type server struct {
	hub      *app.Hub
	http     *http.Server
	upgrader websocket.Upgrader
}

func (s *server) Start() error {
	go func() {
		if err := s.hub.Run(); err != nil {
			log.Printf("Hub error: %v", err)
		}
	}()

	log.Printf("WebSocket server starting on port %s", s.http.Addr)

	return s.http.ListenAndServe()
}

func (s *server) Shutdown(ctx context.Context) error {
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.hub.Shutdown(c); err != nil {
		log.Printf("Error shutting down hub: %v", err)
		return err
	}

	if err := s.http.Shutdown(c); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	log.Println("WebSocket server shut down gracefully")
	return nil
}

type ConnectionRequest struct {
	GameName  string `json:"game_name"`
	ClientID  string `json:"client_id"`
	Name      string `json:"name"`
	SessionID string `json:"session_id,omitempty"`
}

func (s *server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connection, Upgrade, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Protocol")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For WebSocket upgrade, get connection data from query parameters
	var connReq ConnectionRequest
	if r.Header.Get("Upgrade") == "websocket" {
		// WebSocket connection - use query parameters
		connReq = ConnectionRequest{
			ClientID:  r.URL.Query().Get("client_id"),
			GameName:  r.URL.Query().Get("game_name"),
			Name:      r.URL.Query().Get("name"),
			SessionID: r.URL.Query().Get("session_id"),
		}
	} else {
		// Regular POST request - read from body
		if err := json.NewDecoder(r.Body).Decode(&connReq); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	// Validate required fields
	if connReq.ClientID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if connReq.GameName == "" {
		connReq.GameName = "default_room"
	}
	if connReq.Name == "" {
		connReq.Name = connReq.ClientID
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create connection wrapper that implements protocol.RawConnection
	wsConn := NewWebSocketConnection(conn)

	// Dispatch the connection to the hub
	if err := s.hub.DispatchConnection(r.Context(), wsConn, connReq.GameName, connReq.ClientID, connReq.Name, connReq.SessionID); err != nil {
		log.Printf("Failed to dispatch connection: %v", err)

		// Handle specific error types
		if errors.Is(err, app.ErrClientSessionValidationFailed) {
			// Send session invalidation message before closing
			s.sendSessionInvalidMessage(conn, "Session validation failed")
		}

		conn.Close()
		return
	}
}

// sendSessionInvalidMessage sends an error_invalid_session message to the client
func (s *server) sendSessionInvalidMessage(conn *websocket.Conn, message string) {
	errorEnvelope := protocol.ResponseEnvelope{
		Type: protocol.ErrInvalidSession,
	}

	// Create the error payload
	errorPayload := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}

	payloadBytes, err := json.Marshal(errorPayload)
	if err != nil {
		log.Printf("Failed to marshal error payload: %v", err)
		return
	}

	errorEnvelope.Payload = payloadBytes

	// Send the error message
	if err := conn.WriteJSON(errorEnvelope); err != nil {
		log.Printf("Failed to send session invalid message: %v", err)
	} else {
		log.Printf("Sent session invalid message to client: %s", message)
	}

	// Give the client a moment to process the message before connection closes
	time.Sleep(100 * time.Millisecond)
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func NewServer(ctx context.Context, port string) app.Server {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // TODO: Add proper origin validation
		},
	}

	idGen := utils.NewSequentialIDGenerator("session")

	s := &server{
		hub:      app.NewHub(ctx, idGen),
		upgrader: upgrader,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)

	s.http = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return s
}
