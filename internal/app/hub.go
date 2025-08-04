package app

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

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

// clientJoinRequest is a struct to hold a client and the gameName for joining a room.
type clientJoinRequest struct {
	client   Client
	gameName string
}

type Hub struct {
	rooms        map[string]*game.Room
	clients      *ClientRegistry
	idGen        IDGenerator
	hubCh        chan protocol.Command
	clientJoins  chan clientJoinRequest
	clientLeaves chan Client
	ctx          context.Context
	cancel       context.CancelFunc
	shutdownOnce sync.Once
}

func NewHub(ctx context.Context, idGen IDGenerator) *Hub {
	hubCtx, cancel := context.WithCancel(ctx)
	return &Hub{
		rooms: map[string]*game.Room{
			DefaultRoomName: game.NewRoom(hubCtx, DefaultRoomName),
		},
		clients:      NewClientRegistry(),
		idGen:        idGen,
		hubCh:        make(chan protocol.Command, 256),
		clientJoins:  make(chan clientJoinRequest, 10),
		clientLeaves: make(chan Client, 10),
		ctx:          hubCtx,
		cancel:       cancel,
	}
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
		case req := <-h.clientJoins:
			h.handleJoin(req)
		case client := <-h.clientLeaves:
			h.handleLeave(client)
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

func (h *Hub) handleJoin(req clientJoinRequest) {
	client := req.client
	gameName := req.gameName

	log.Printf("Client %s joining hub for game '%s'", client.ID(), gameName)
	h.clients.Add(client)

	if room, ok := h.rooms[gameName]; ok {
		if err := room.AddPlayer(client.ID()); err != nil {
			log.Printf("Failed to add player for client %s to room %s: %v", client.ID(), room.ID, err)
			client.Close()
		}
	} else {
		log.Printf("Room '%s' not found for joining client %s", gameName, client.ID())
		client.Close()
	}
}

func (h *Hub) handleLeave(client Client) {
	log.Printf("Client %s leaving hub", client.ID())
	h.clients.Remove(client.ID())
	// TODO: This needs to know which room the client was in.
	if room, ok := h.rooms[DefaultRoomName]; ok {
		room.RemovePlayer(client.ID())
	}
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

func (h *Hub) generateSessionID() string {
	return h.idGen.GenerateID()
}

func (h *Hub) DispatchConnection(ctx context.Context, conn protocol.RawConnection, gameName, clientID, name, sessionID string) error {
	if client, ok := h.clients.Get(clientID); ok {
		if client.SessionID() == sessionID {
			log.Printf("Client %s reconnected with session %s", clientID, sessionID)
			// TODO: Handle reconnection logic
			return nil
		} else {
			return fmt.Errorf("client ID %s already exists with a different session", clientID)
		}
	}

	client := newWebsocketClient(h.ctx, clientID, name, conn, protocol.NewJsonCodec())

	_, err := client.Subscribe(func(cmd protocol.Command) {
		h.hubCh <- cmd
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to client %s: %w", clientID, err)
	}

	newSessionID := h.generateSessionID()
	if err := client.SetSessionID(newSessionID); err != nil {
		return fmt.Errorf("failed to set session ID for client %s: %w", clientID, err)
	}

	h.clientJoins <- clientJoinRequest{client: client, gameName: gameName}

	go func() {
		defer func() {
			h.clientLeaves <- client
		}()
		for err := range client.Errors() {
			log.Printf("Client %s error: %v", client.ID(), err)
			return
		}
	}()

	return nil
}
