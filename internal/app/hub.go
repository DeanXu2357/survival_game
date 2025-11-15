package app

import (
	"context"
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
	hubCommandCh chan protocol.RequestCommand
	ctx          context.Context
	cancel       context.CancelFunc
	shutdownOnce sync.Once
}

func NewHub(ctx context.Context, idGen IDGenerator) *Hub {
	hubCtx, cancel := context.WithCancel(ctx)

	return &Hub{
		rooms:        make(map[string]*game.Room),
		clients:      NewClientRegistry(idGen),
		hubCommandCh: make(chan protocol.RequestCommand, 256),
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
			case protocol.ListRoomsEnvelope:
				clientID := cmd.ClientID

				rooms := make([]protocol.RoomInfo, 0, len(h.rooms))
				for roomID, room := range h.rooms {
					rooms = append(rooms, protocol.RoomInfo{
						RoomID:      roomID,
						PlayerCount: room.PlayerCount(),
					})
				}

				if client, ok := h.clients.Get(clientID); ok {
					ctx, cancel := context.WithTimeout(h.ctx, 2*time.Second)
					if err := client.Send(ctx, protocol.ListRoomsResponseEnvelope, protocol.ListRoomsResponse{Rooms: rooms}); err != nil {
						log.Printf("Failed to send ListRooms response to client %s: %v", clientID, err)
					}
					cancel()
				} else {
					log.Printf("Client %s not found for ListRooms response", clientID)
				}

			case protocol.RequestJoinEnvelope:
				clientID := cmd.ClientID

				_, valid := cmd.ParsedPayload.(protocol.RequestJoinPayload)
				if !valid {
					log.Printf("Invalid join payload from client %s", clientID)
					continue
				}

				// Join Room
				// for testing, only support joining the default room
				responseType := protocol.JoinRoomSuccessEnvelope
				var responsePayload protocol.ErrorPayload
				if err := h.JoinRoom(clientID, DefaultRoomName); err != nil {
					responseType = protocol.ErrorResponseEnvelope
					responsePayload = protocol.ErrorPayload{
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
		room.SendStaticData([]string{clientID})
		log.Printf("Client %s joined room %s", clientID, roomID)
		return nil
	}

	handler := func(cmd protocol.RequestCommand) {
		if cmd.EnvelopeType != protocol.PlayerInputEnvelope {
			log.Printf("Ignoring non-input command from client %s", client.ID())
			return
		}

		// TODO: Add handling logic for other EnvelopeTypes in future
		input, valid := cmd.ParsedPayload.(protocol.PlayerInput)
		if !valid {
			log.Printf("Invalid input payload from client %s", client.ID())
			return
		}

		inputCmd := protocol.Command{
			ClientID: client.ID(),
			Input:    input,
		}

		room.SendCommand(inputCmd)
	}

	if err := client.Subscribe(handler); err != nil {
		return fmt.Errorf("failed to subscribe client %s to room %s: %w", clientID, roomID, err)
	}

	return nil
}

func (h *Hub) DispatchConnection(ctx context.Context, conn protocol.RawConnection, gameName, clientID, name, sessionID string) error {
	client := newWebsocketClient(h.ctx, clientID, name, conn, protocol.NewJsonCodec())

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

	if err := client.Subscribe(func(cmd protocol.RequestCommand) {
		h.hubCommandCh <- cmd
	}); err != nil {
		return fmt.Errorf("failed to subscribe to client %s: %w", clientID, err)
	}

	return nil
}

func (h *Hub) initializeDefaultGame() {
	roomID := DefaultRoomName

	room := createDefaultRoom(h.ctx, roomID)

	h.rooms[roomID] = room

	go room.Run()

	handler := func(msg game.UpdateMessage) {
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

			if err := client.Send(context.Background(), protocol.GameUpdateEnvelope, msg.Envelope); err != nil {
				log.Printf("Failed to send message to client %s: %v", client.ID(), err)
			}
		}
	}

	if err := room.SubscribeResponse(handler); err != nil {
		panic(err)
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
