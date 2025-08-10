package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"survival/internal/game"
	maploader "survival/internal/infrastructure/map"
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
	rooms        map[string]*game.Room
	clients      *ClientRegistry
	hubCh        chan protocol.Command
	ctx          context.Context
	cancel       context.CancelFunc
	shutdownOnce sync.Once
}

func NewHub(ctx context.Context, idGen IDGenerator) *Hub {
	hubCtx, cancel := context.WithCancel(ctx)

	defaultRoom := createDefaultRoom(hubCtx, DefaultRoomName)

	return &Hub{
		rooms: map[string]*game.Room{
			DefaultRoomName: defaultRoom,
		},
		clients: NewClientRegistry(idGen),
		hubCh:   make(chan protocol.Command, 256),
		ctx:     hubCtx,
		cancel:  cancel,
	}
}

func createDefaultRoom(ctx context.Context, roomID string) *game.Room {
	jsonLoader := maploader.NewJSONMapLoader("./maps")
	mapConfig, err := jsonLoader.LoadMap("office_floor_01")
	if err != nil {
		log.Printf("Failed to load office_floor_01 from JSON: %v, using empty room", err)
		return game.NewRoom(ctx, roomID)
	}

	return game.NewRoomWithMap(ctx, roomID, mapConfig)
}

func (h *Hub) Run() error {
	log.Println("Hub is running...")
	for _, room := range h.rooms {
		go room.Run()
		go h.routeOutgoing(room)
	}

	return h.hubLoop()
}

func (h *Hub) hubLoop() error {
	for {
		select {
		case cmd := <-h.hubCh:
			// For now, we assume all commands go to the default room.
			// This could be expanded to route to different rooms based on client state.
			if room, ok := h.rooms[DefaultRoomName]; ok {
				room.CommandsChannel() <- cmd
			} else {
				log.Printf("Warning: Room not found for command from client %s", cmd.ClientID)
			}
		case <-h.ctx.Done():
			log.Println("Hub context is done, shutting down loop.")
			return h.ctx.Err()
		}
	}
}

func (h *Hub) routeOutgoing(room *game.Room) {
	outgoingCh := room.OutgoingChannel()
	for {
		select {
		case msg := <-outgoingCh:
			for _, clientID := range msg.TargetClientIDs {
				if client, ok := h.clients.Get(clientID); ok {
					ctx, cancel := context.WithTimeout(h.ctx, 2*time.Second)
					if err := client.Send(ctx, msg.Envelope); err != nil {
						log.Printf("Failed to send message to client %s: %v", clientID, err)
					}
					cancel()
				}
			}
		case <-h.ctx.Done():
			log.Printf("Stopping outgoing router for room %s", room.ID)
			return
		}
	}
}

func (h *Hub) handleLeave(client Client) {
	log.Printf("Client %s leaving hub", client.ID())
	h.clients.Remove(client.ID())

	// Note: Don't immediately remove player from room to support reconnection
	// Players will be removed later by session cleanup or explicit disconnect
	log.Printf("Client %s disconnected but player remains in room for potential reconnection", client.ID())

	client.Close() // Ensure client resources are cleaned up
}

func (h *Hub) Shutdown(ctx context.Context) error {
	h.shutdownOnce.Do(func() {
		log.Println("Hub shutdown initiated.")
		h.cancel() // Cancel the hub's context to stop loops

		// Close all client connections
		for _, client := range h.clients.All() {
			log.Printf("Closing client %s during hub shutdown", client.ID())
			if err := client.Close(); err != nil {
				log.Printf("Error closing client %s: %v", client.ID(), err)
			} else {
				log.Printf("Client %s closed successfully", client.ID())
			}
		}
		h.clients.Clear()

		// Shutdown all rooms
		for _, room := range h.rooms {
			log.Printf("Shutting down room: %s", room.ID)
			if err := room.Shutdown(ctx); err != nil {
				log.Printf("Error shutting down room %s: %v", room.ID, err)
			}
		}
		log.Println("Hub shutdown complete.")
	})
	return nil
}

// whenClientsAddNewConnection handles new client connections by adding them to the appropriate room
func (h *Hub) whenClientsAddNewConnection(client Client, sessionID, gameName string) {
	if room, ok := h.rooms[gameName]; ok {
		if err := room.AddPlayer(client.ID()); err != nil {
			log.Printf("Failed to add player for client %s to room %s: %v", client.ID(), room.ID, err)
			client.Close() // TODO: or maybe send a notify to client that join failed ?
			return
		}

		log.Printf("Client %s joined room %s as new player with session %s", client.ID(), gameName, sessionID)
	} else {
		log.Printf("Room '%s' not found for joining client %s", gameName, client.ID())
		client.Close() // TODO: notify ?
	}
}

// whenClientsDoReconnection handles successful client reconnections
func (h *Hub) whenClientsDoReconnection(client Client, sessionID, gameName string) {
	log.Printf("Client %s successfully reconnected to game %s with session %s", client.ID(), gameName, sessionID)
}

func (h *Hub) DispatchConnection(ctx context.Context, conn protocol.RawConnection, gameName, clientID, name, sessionID string) error {
	client := newWebsocketClient(h.ctx, clientID, name, conn, protocol.NewJsonCodec())

	_, err := client.Subscribe(func(cmd protocol.Command) {
		h.hubCh <- cmd
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to client %s: %w", clientID, err)
	}

	if errAdd := h.clients.Add(client, sessionID, gameName, h.whenClientsAddNewConnection, h.whenClientsDoReconnection); errAdd != nil {
		// For session validation errors, let the server handle the error response
		// For other errors, wrap with additional context
		if errors.Is(errAdd, ErrClientSessionValidationFailed) {
			return errAdd // Return the original error for server to handle
		} else {
			return fmt.Errorf("failed to add client %s to registry: %w", clientID, errAdd)
		}
	}

	go func() {
		defer func() {
			h.handleLeave(client)
		}()
		for err := range client.Errors() {
			log.Printf("Client %s error: %v", client.ID(), err)
			return
		}
	}()

	return nil
}
