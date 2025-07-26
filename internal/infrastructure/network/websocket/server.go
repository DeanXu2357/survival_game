package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"survival/internal/app"
	"survival/internal/utils"
)

type server struct {
	hub      *app.Hub
	http     *http.Server
	upgrader websocket.Upgrader
}

func (s *server) Start() error {
	if err := s.hub.Run(); err != nil {
		return fmt.Errorf("failed to start hub: %w", err)
	}

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
		log.Printf("Error shutting down HTTP server: %v", err)
		return err
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
	// Read connection data from request body
	var connReq ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&connReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
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
		conn.Close()
		return
	}
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func NewServer(port string) app.Server {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // TODO: Add proper origin validation
		},
	}

	idGen := utils.NewSequentialIDGenerator("session")

	s := &server{
		hub:      app.NewHub(idGen),
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
