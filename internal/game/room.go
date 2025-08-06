package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"survival/internal/protocol"
)

type OutgoingMessage struct {
	TargetClientIDs []string // an empty slice means broadcast to all
	Envelope        protocol.ResponseEnvelope
}

type Room struct {
	ID        string
	mapConfig *MapConfig
	state     *State
	logic     *Logic
	players   *PlayerRegistry
	commands  chan protocol.Command
	outgoing  chan OutgoingMessage

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRoom(ctx context.Context, id string) *Room {
	roomCTX, cancel := context.WithCancel(ctx)

	return &Room{
		ID:        id,
		mapConfig: nil, // No map configuration
		state:     NewGameState(),
		logic:     NewGameLogic(),
		players:   NewPlayerRegistry(),
		commands:  make(chan protocol.Command, 200),
		outgoing:  make(chan OutgoingMessage, 400),

		ctx:    roomCTX,
		cancel: cancel,
	}
}

func NewRoomWithMap(ctx context.Context, id string, mapConfig *MapConfig) *Room {
	roomCTX, cancel := context.WithCancel(ctx)

	return &Room{
		ID:        id,
		mapConfig: mapConfig,
		state:     NewGameStateFromMap(mapConfig),
		logic:     NewGameLogic(),
		players:   NewPlayerRegistry(),
		commands:  make(chan protocol.Command, 200),
		outgoing:  make(chan OutgoingMessage, 400),

		ctx:    roomCTX,
		cancel: cancel,
	}
}

func (r *Room) CommandsChannel() chan protocol.Command {
	return r.commands
}

// OutgoingChannel returns a read-only channel for outgoing messages.
func (r *Room) OutgoingChannel() <-chan OutgoingMessage {
	return r.outgoing
}

func (r *Room) Run() {
	defer func() {
		if rcv := recover(); rcv != nil {
			log.Printf("Room %s panic: %v", r.ID, rcv)
		}
		log.Printf("Room %s stopped", r.ID)
	}()

	ticker := time.NewTicker(time.Second / targetTickRate)
	defer ticker.Stop()

	currentInputs := make(map[string]protocol.PlayerInput)

	for {
		select {
		case cmd := <-r.commands:
			playerID, ok := r.players.GetPlayerID(cmd.ClientID)
			if !ok {
				log.Printf("Warning: No player ID found for client %s", cmd.ClientID)
				continue
			}
			currentInputs[playerID] = cmd.Input
		case <-ticker.C:
			r.logic.Update(r.state, currentInputs, deltaTime)
			currentInputs = make(map[string]protocol.PlayerInput) // Reset inputs after processing
			r.broadcastGameUpdate()
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Room) Shutdown(ctx context.Context) error {
	defer r.cancel()

	// In the new architecture, the room is no longer responsible for closing client connections.
	// It just needs to clean up its own resources.
	log.Printf("Room %s shutdown initiated", r.ID)

	// save game state if needed
	r.players.Clear()

	log.Printf("Room %s shutdown completed", r.ID)
	return nil
}

// AddPlayer creates a new player in the game state and associates it with a client ID.
func (r *Room) AddPlayer(clientID string) error {
	var player *Player
	var err error

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
	r.players.Register(clientID, player.ID)

	log.Printf("Player created and registered successfully - Client: %s, Player: %s, Position: %+v",
		clientID, player.ID, player.Position)
	log.Printf("Total players in room: %d", len(r.state.Players))

	return nil
}

// RemovePlayer removes a player from the game state and the registry.
func (r *Room) RemovePlayer(clientID string) {
	if playerID, ok := r.players.GetPlayerID(clientID); ok {
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

	envelope := protocol.ResponseEnvelope{
		Type:    protocol.GameUpdateEnvelope,
		Payload: json.RawMessage(payloadBytes),
	}

	// Push to all clients in the room temporarily
	r.outgoing <- OutgoingMessage{
		TargetClientIDs: r.players.AllClientIDs(),
		Envelope:        envelope,
	}
}
