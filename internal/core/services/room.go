package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"survival/internal/core/domain"
	"survival/internal/core/ports"
	"survival/internal/utils"
)

type UpdateMessage struct {
	ToSessions []string // an empty slice means broadcast to all
	Envelope   ports.ResponseEnvelope
}

type Room struct {
	ID         string
	mapConfig  *domain.MapConfig
	state      *domain.State
	logic      *domain.Logic
	players    *domain.PlayerRegistry
	subManager *Manager[UpdateMessage]

	commands chan ports.Command
	outgoing chan UpdateMessage

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRoom(ctx context.Context, id string) *Room {
	roomCTX, cancel := context.WithCancel(ctx)

	return &Room{
		ID:         id,
		mapConfig:  nil, // No map configuration
		state:      domain.NewGameState(),
		logic:      domain.NewGameLogic(),
		players:    domain.NewPlayerRegistry(),
		subManager: NewManager[UpdateMessage](utils.NewSequentialIDGenerator(fmt.Sprintf("room%s-sub-", id))),

		commands: make(chan ports.Command, 200),
		outgoing: make(chan UpdateMessage, 400),

		ctx:    roomCTX,
		cancel: cancel,
	}
}

func NewRoomWithMap(ctx context.Context, id string, mapConfig *domain.MapConfig) *Room {
	roomCTX, cancel := context.WithCancel(ctx)

	return &Room{
		ID:         id,
		mapConfig:  mapConfig,
		state:      domain.NewGameStateFromMap(mapConfig),
		logic:      domain.NewGameLogic(),
		players:    domain.NewPlayerRegistry(),
		subManager: NewManager[UpdateMessage](utils.NewSequentialIDGenerator(fmt.Sprintf("room%s-sub-", id))),

		commands: make(chan ports.Command, 200),
		outgoing: make(chan UpdateMessage, 400),

		ctx:    roomCTX,
		cancel: cancel,
	}
}

func (r *Room) SendCommand(cmd ports.Command) {
	select {
	case r.commands <- cmd:
	default:
		log.Printf("Room %s command channel full, dropping command from client %s", r.ID, cmd.SessionID)
	}
}

func (r *Room) SubscribeResponse(handler func(msg UpdateMessage)) error {
	// todo: check room is running

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	_, err := r.subManager.Add(handler)
	if err != nil {
		return fmt.Errorf("failed to add subscription to room %s: %w", r.ID, err)
	}

	return nil
}

// OutgoingChannel returns a read-only channel for outgoing messages.
func (r *Room) OutgoingChannel() <-chan UpdateMessage {
	return r.outgoing
}

func (r *Room) responsePump() {
	for msg := range r.outgoing {

		for _, sub := range r.subManager.All() {
			if err := sub.Push(msg); err != nil {
				log.Printf("Failed to push message to subscription %s: %v", sub.ID(), err)
			}
		}

		select {
		case <-r.ctx.Done():
			return
		default:
		}
	}
}

func (r *Room) Run() {
	defer func() {
		if rcv := recover(); rcv != nil {
			log.Printf("Room %s panic: %v", r.ID, rcv)
		}
		log.Printf("Room %s stopped", r.ID)
	}()

	// TODO: should extract  game run and member management to separate methods.

	go r.responsePump()

	ticker := time.NewTicker(time.Second / ports.TargetTickRate)
	defer ticker.Stop()

	currentInputs := make(map[string]ports.PlayerInput)

	for {
		select {
		case cmd := <-r.commands:
			playerID, ok := r.players.PlayerID(cmd.SessionID)
			if !ok {
				log.Printf("Warning: No player ID found for client %s", cmd.SessionID)
				continue
			}
			currentInputs[playerID] = cmd.Input
		case <-ticker.C:
			r.logic.Update(r.state, currentInputs, ports.DeltaTime)
			currentInputs = make(map[string]ports.PlayerInput) // Reset inputs after processing
			r.broadcastGameUpdate()
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Room) Shutdown(ctx context.Context) error {
	defer r.cancel()

	log.Printf("Room %s shutdown initiated", r.ID)

	r.players.Clear()
	r.subManager.Clear()

	log.Printf("Room %s shutdown completed", r.ID)
	return nil
}

// AddPlayer creates a new player or reconnects existing players and registers them to the room.
func (r *Room) AddPlayer(client Client) error {
	var player *domain.Player
	var err error

	sessionID := client.SessionID()
	_, exist := r.players.PlayerID(sessionID)
	if exist {
		return nil
	}

	// Use spawn point if map is configured
	if r.mapConfig != nil {
		spawnPoint := r.mapConfig.GetRandomSpawnPoint()
		if spawnPoint != nil {
			player, err = r.state.NewPlayerAtPosition(spawnPoint.Position)
		} else {
			player, err = r.state.NewPlayer() // fallback to default position
		}
	} else {
		player, err = r.state.NewPlayer()
	}
	if err != nil {
		return fmt.Errorf("failed to create new player: %w", err)
	}

	r.players.Register(client.SessionID(), player.ID)

	log.Printf("Player created and registered successfully - Client: %s, Player: %s, Position: %+v", client, player.ID, player.Position)
	log.Printf("Total players in room: %d", len(r.state.Players))

	return nil
}

// SendStaticData sends static map data (walls, dimensions) to specific clients
func (r *Room) SendStaticData(sessionIDs []string) {
	staticData := r.state.ToStaticData()

	payloadBytes, err := json.Marshal(staticData)
	if err != nil {
		return
	}

	envelope := ports.ResponseEnvelope{
		EnvelopeType: ports.StaticDataEnvelope,
		Payload:      json.RawMessage(payloadBytes),
	}

	r.outgoing <- UpdateMessage{
		ToSessions: sessionIDs,
		Envelope:   envelope,
	}
}

// RemovePlayer removes a player from the game state and the registry.
func (r *Room) RemovePlayer(clientID string) {
	if playerID, ok := r.players.PlayerID(clientID); ok {
		r.players.Unregister(clientID)
		log.Printf("Player %s (Client %s) removed from room %s", playerID, clientID, r.ID)
	}
}

func (r *Room) broadcastGameUpdate() {
	gameUpdate := r.state.ToClientState()

	payloadBytes, err := json.Marshal(gameUpdate)
	if err != nil {
		log.Printf("Failed to marshal game update: %v", err)
		return
	}

	envelope := ports.ResponseEnvelope{
		EnvelopeType: ports.GameUpdateEnvelope,
		Payload:      json.RawMessage(payloadBytes),
	}

	// Push to all clients in the room temporarily
	r.outgoing <- UpdateMessage{
		ToSessions: r.players.AllSessionIDs(),
		Envelope:   envelope,
	}
}

func (r *Room) PlayerCount() int {
	return len(r.state.Players)
}

func (r *Room) Name() string {
	if r.mapConfig != nil {
		return r.mapConfig.Name
	}
	return r.ID
}

func (r *Room) MaxPlayers() int {
	return 0
}
