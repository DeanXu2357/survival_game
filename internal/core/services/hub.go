package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"survival/internal/adapters/repository/maploader"
	"survival/internal/core/ports"
	"survival/internal/utils"
)

type IDGenerator interface {
	GenerateID() string
}

const (
	DefaultRoomName = "default_room"
	LocalUser       = "localhost"
)

type Hub struct {
	rooms        map[string]*Room
	clients      *ClientRegistry
	hubCommandCh chan ports.RequestCommand
	ctx          context.Context
	cancel       context.CancelFunc
	shutdownOnce sync.Once
}

func NewHub(ctx context.Context, idGen IDGenerator) *Hub {
	hubCtx, cancel := context.WithCancel(ctx)

	return &Hub{
		rooms:        make(map[string]*Room),
		clients:      NewClientRegistry(idGen),
		hubCommandCh: make(chan ports.RequestCommand, 256),
		ctx:          hubCtx,
		cancel:       cancel,
	}
}

func (h *Hub) Run() error {
	log.Println("Hub is running...")
	h.initializeDefaultGame()

	return h.hubLoop()
}

func (h *Hub) hubLoop() error {
	for {
		select {
		case cmd := <-h.hubCommandCh:
			switch cmd.EnvelopeType {
			case ports.ListRoomsEnvelope:
				clientID := cmd.ClientID

				rooms := make([]ports.RoomInfo, 0, len(h.rooms))
				for roomID, room := range h.rooms {
					rooms = append(rooms, ports.RoomInfo{
						RoomID:      roomID,
						Name:        room.Name(),
						PlayerCount: room.PlayerCount(),
						MaxPlayers:  room.MaxPlayers(),
					})
				}

				if client, ok := h.clients.Get(clientID); ok {
					ctx, cancel := context.WithTimeout(h.ctx, 2*time.Second)
					if err := client.Send(ctx, ports.ListRoomsResponseEnvelope, ports.ListRoomsResponse{Rooms: rooms}); err != nil {
						log.Printf("Failed to send ListRooms response to client %s: %v", clientID, err)
					}
					cancel()
				} else {
					log.Printf("Client %s not found for ListRooms response", clientID)
				}

			case ports.RequestJoinEnvelope:
				clientID := cmd.ClientID

				_, valid := cmd.ParsedPayload.(*ports.RequestJoinPayload)
				if !valid {
					log.Printf("Invalid join payload from client %s", clientID)
					continue
				}

				// Join Room
				// for testing, only support joining the default room
				// TODO: JoinRoomSuccess should include player's EntityID in payload.
				// Requires: room.AddPlayer() and hub.JoinRoom() to return EntityID.
				responseType := ports.JoinRoomSuccessEnvelope
				var responsePayload ports.ErrorPayload
				if err := h.JoinRoom(clientID, DefaultRoomName); err != nil {
					responseType = ports.ErrorResponseEnvelope
					responsePayload = ports.ErrorPayload{
						Code:    500,
						Message: fmt.Sprintf("Failed to join room: %v", err),
					}
				}

				// Notify client of join result
				if client, ok := h.clients.Get(clientID); ok {
					ctx, cancel := context.WithTimeout(h.ctx, 2*time.Second)
					if errSend := client.Send(ctx, responseType, responsePayload); errSend != nil {
						log.Printf("Failed to send error message to client %s: %v", clientID, errSend)
					}
					cancel()
				}
			default:
				log.Printf("Unhandled command type: %s from client %s", cmd.EnvelopeType, cmd.ClientID)

			}
		case <-h.ctx.Done():
			log.Println("Hub context is done, shutting down loop.")
			return h.ctx.Err()
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

func (h *Hub) JoinRoom(clientID, roomID string) error {
	room, ok := h.rooms[roomID]
	if !ok {
		return fmt.Errorf("room '%s' not found for client %s", roomID, clientID)
	}

	client, ok := h.clients.Get(clientID)
	if !ok {
		return fmt.Errorf("client '%s' not found in registry", clientID)
	}
	if client == nil {
		return fmt.Errorf("client '%s' is nil in registry", clientID)
	}

	if err := room.AddPlayer(client); err != nil {
		return fmt.Errorf("failed to add player to room %s: %w", roomID, err)
	}

	// Send static data (map, walls, etc.) to the newly joined client
	room.SendStaticData([]string{client.SessionID()})
	log.Printf("[JoinRoom] Client %s (session: %s) joined room %s, SendStaticData called", clientID, client.SessionID(), roomID)

	handler := func(cmd ports.RequestCommand) {
		if cmd.EnvelopeType != ports.PlayerInputEnvelope {
			log.Printf("Ignoring non-input command from client %s", client.ID())
			return
		}

		// TODO: Add handling logic for other EnvelopeTypes in future
		input, valid := cmd.ParsedPayload.(*ports.PlayerInput)
		if !valid {
			log.Printf("Invalid input payload from client %s", client.ID())
			return
		}

		inputCmd := ports.Command{
			SessionID: client.SessionID(),
			Input:     *input,
		}

		room.SendCommand(inputCmd)
	}

	if err := client.Subscribe(handler); err != nil {
		return fmt.Errorf("failed to subscribe client %s to room %s: %w", clientID, roomID, err)
	}

	return nil
}

func (h *Hub) DispatchConnection(ctx context.Context, conn ports.RawConnection, gameName, clientID, name, sessionID string) error {
	client := newWebsocketClient(h.ctx, clientID, name, conn, utils.NewJsonCodec())

	go func() {
		defer func() {
			h.handleLeave(client)
		}()
		for err := range client.Errors() {
			log.Printf("Client %s error: %v", client.ID(), err)
			return
		}
	}()

	if errAdd := h.clients.Add(client, sessionID); errAdd != nil {
		return fmt.Errorf("failed to add client %s to registry: %w", clientID, errAdd)
	}

	if err := client.Subscribe(func(cmd ports.RequestCommand) {
		// Only forward hub-level commands (not player_input, which is handled by room)
		if cmd.EnvelopeType != ports.PlayerInputEnvelope {
			h.hubCommandCh <- cmd
		}
	}); err != nil {
		return fmt.Errorf("failed to subscribe to client %s: %w", clientID, err)
	}

	return nil
}

func (h *Hub) initializeDefaultGame() {
	roomID := DefaultRoomName

	room, err := createDefaultRoom(h.ctx, roomID)
	if err != nil {
		log.Printf("Failed to create default room: %v", err)
		return
	}

	h.rooms[roomID] = room

	go room.Run()

	handler := func(msg UpdateMessage) {
		sessionIDs := msg.ToSessions
		if len(sessionIDs) == 0 {
			return
		}

		for _, sessionID := range sessionIDs {
			client, ok := h.clients.GetBySessionID(sessionID)
			if !ok {
				log.Printf("No client found for session ID %s", sessionID)
				continue
			}

			if err := client.Send(context.Background(), msg.Envelope.EnvelopeType, msg.Envelope.Payload); err != nil {
				log.Printf("Failed to send message to client %s: %v", client.ID(), err)
			}
		}
	}

	if err := room.SubscribeResponse(handler); err != nil {
		panic(err)
	}
}

func createDefaultRoom(ctx context.Context, roomID string) (*Room, error) {
	jsonLoader := maploader.NewJSONMapLoader("./maps")
	mapConfig, err := jsonLoader.LoadMap("office_floor_01")
	if err != nil {
		log.Printf("Failed to load office_floor_01 from JSON: %v, using default map", err)
		return NewRoom(ctx, roomID)
	}

	return NewRoomWithMap(ctx, roomID, mapConfig)
}
