package app

import (
	"context"
	"fmt"
	"log"
	"sync"

	"survival/internal/game"
	"survival/internal/protocol"
)

type IDGenerator interface {
	GenerateID() string
}

const (
	DefaultRoomName = "default_room"
	LocalUser       = "localhost"
)

type Hub struct {
	rooms      map[string]*game.Room // Map of room names to Room objects
	clientsMU  sync.RWMutex
	sessionMap map[string]string // Map of session IDs to Client id
	idGen      IDGenerator
}

func (h *Hub) Run() error {
	for _, room := range h.rooms {
		go room.Run()
	}

	return nil
}

func (h *Hub) Shutdown(ctx context.Context) error {
	for _, room := range h.rooms {
		log.Printf("Shutting down room: %s", room.ID)
		if err := room.Shutdown(ctx); err != nil {
			return err
		}
	}

	defer log.Printf("Hub clean up")

	h.clientsMU.Lock()
	defer h.clientsMU.Unlock()

	h.sessionMap = make(map[string]string) // Clear session map

	defer log.Printf("Session map cleared")

	return nil
}

func (h *Hub) generateSessionID() string {
	return h.idGen.GenerateID()
}

func (h *Hub) DispatchConnection(ctx context.Context, conn protocol.RawConnection, gameName, clientID, name, sessionID string) error {
	client := newWebsocketClient(ctx, clientID, name, conn, protocol.NewJsonCodec())

	h.clientsMU.Lock()
	if id, exists := h.sessionMap[sessionID]; exists { // reconnection
		if id != clientID {
			return fmt.Errorf("session ID %s is already in use by another client", sessionID)
		}
		if err := client.SetSessionID(sessionID); err != nil {
			return fmt.Errorf("failed to set session ID for client %s: %w", clientID, err)
		}
		return nil
	}

	sessionID = h.generateSessionID() // Generate a new session ID if not provided
	if err := client.SetSessionID(sessionID); err != nil {
		return fmt.Errorf("failed to set session ID for client %s: %w", clientID, err)
	}
	h.sessionMap[sessionID] = clientID // Store the session ID in the session map

	h.clientsMU.Unlock()

	room, exists := h.rooms[gameName]
	if !exists {
		return fmt.Errorf("room %s does not exist", gameName)
	}
	if err := room.AddClient(ctx, client); err != nil {
		return fmt.Errorf("failed to add client to room %s: %w", gameName, err)
	}
	client.SetReceiveChannel(room.CommandsChannel())

	defer func() {
		log.Println("clean up for client:", client.ID())
	}()

	// todo: it's wired that a client handle connection here, but it is used in the room. Where should I close the connection , in Room or Hub?
	return client.Pump()
}

func NewHub(ctx context.Context, idGen IDGenerator) *Hub {
	return &Hub{
		clientsMU: sync.RWMutex{},
		rooms: map[string]*game.Room{
			DefaultRoomName: game.NewRoom(ctx, DefaultRoomName),
		},
		sessionMap: make(map[string]string),
		idGen:      idGen,
	}
}
